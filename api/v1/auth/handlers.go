package auth

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/auth"
	"github.com/qenti/qenti/internal/pkg/invitations"
	"github.com/qenti/qenti/internal/pkg/jwt"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/producers"
	"github.com/qenti/qenti/internal/pkg/users"
)

type Handlers struct {
	authService       *auth.Service
	jwtService        *jwt.Service
	refreshTokenRepo  *auth.RefreshTokenRepository
	usersRepo         *users.Repository
	producersRepo     *producers.Repository
	invitationsRepo   *invitations.Repository
	superAdminEmail   string // email fijo para auto-provisionar super_admin
}

func NewHandlers(
	authService *auth.Service,
	jwtService *jwt.Service,
	db *sql.DB,
	usersRepo *users.Repository,
	producersRepo *producers.Repository,
	invitationsRepo *invitations.Repository,
	superAdminEmail string,
) *Handlers {
	return &Handlers{
		authService:     authService,
		jwtService:      jwtService,
		refreshTokenRepo: auth.NewRefreshTokenRepository(db),
		usersRepo:       usersRepo,
		producersRepo:   producersRepo,
		invitationsRepo: invitationsRepo,
		superAdminEmail: superAdminEmail,
	}
}

// LoginRequest representa el payload de login
type LoginRequest struct {
	FirebaseToken string `json:"firebase_token" binding:"required"`
}

// LoginResponse representa la respuesta de login
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo contiene información básica del usuario
type UserInfo struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	IsPremium      bool   `json:"is_premium"`
	CoinBalance    int    `json:"coin_balance"`
	Role           string `json:"role"` // "user", "admin", "super_admin", "producer"
	ProducerID     string `json:"producer_id,omitempty"`
	// NeedsOnboarding es true si el usuario no tiene rol de producer aún
	// — el frontend debe mostrar el formulario de creación de productora.
	NeedsOnboarding bool   `json:"needs_onboarding,omitempty"`
	// ProducerStatus indica el estado del tenant: pending | active | suspended
	ProducerStatus  string `json:"producer_status,omitempty"`
}

// RefreshRequest representa el payload de refresh token
type RefreshRequest struct {
	Token string `json:"token" binding:"required"`
}

// Login autentica un usuario usando Firebase token y devuelve JWT con rol
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Verificar token de Firebase
	user, err := h.authService.VerifyToken(req.FirebaseToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid Firebase token",
		})
		return
	}

	// Obtener o crear usuario en DB
	ctx := c.Request.Context()
	dbUser, err := h.authService.GetOrCreateUser(ctx, user.FirebaseUID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get or create user",
		})
		return
	}

	// Obtener rol multi-tenant (super_admin / admin / producer / user)
	role, producerID, _ := h.authService.GetUserRole(user.FirebaseUID)

	// Auto-provisionar super_admin para el email configurado en SUPER_ADMIN_EMAIL.
	// Solo aplica cuando la cuenta aún no tiene ningún rol privilegiado.
	if h.superAdminEmail != "" && dbUser.Email == h.superAdminEmail && role == "user" {
		if err := h.authService.GrantRole(ctx, dbUser.ID, "super_admin", dbUser.ID); err != nil {
			log.Printf("Warning: Failed to grant super_admin role: %v", err)
		} else {
			role = "super_admin"
			log.Printf("✅ super_admin auto-provisioned for %s", dbUser.Email)
		}
	}

	// Generar access token (24 horas)
	accessToken, _, err := h.jwtService.GenerateToken(
		dbUser.ID,
		dbUser.Email,
		role,
		producerID,
		24,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	// Para producers, verificar el status de su productora
	userInfo := UserInfo{
		ID:          dbUser.ID.String(),
		Email:       dbUser.Email,
		IsPremium:   dbUser.IsPremium,
		CoinBalance: dbUser.CoinBalance,
		Role:        role,
		ProducerID:  producerID,
	}

	if role == "producer" && producerID != "" {
		// Resolver producer_status para que el frontend sepa si está pendiente o activo
		if pID, err2 := parseProducerUUID(producerID); err2 == nil {
			if p, err3 := h.producersRepo.GetByID(ctx, pID); err3 == nil && p != nil {
				userInfo.ProducerStatus = p.Status
			}
		}
	} else if role == "user" {
		// Usuario sin productora → necesita completar onboarding
		userInfo.NeedsOnboarding = true
	}
	
	// Generar refresh token (7 días)
	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})
		return
	}
	
	// Guardar refresh token en DB
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 días
	if err := h.refreshTokenRepo.Create(ctx, refreshToken, dbUser.ID, refreshExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save refresh token",
		})
		return
	}
	
	expiresAt := time.Now().Add(24 * time.Hour)
	
	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: userInfo,
	})
}

// Refresh refresca un access token usando un refresh token
func (h *Handlers) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// Validar refresh token
	userID, err := h.refreshTokenRepo.Validate(ctx, req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	// Obtener usuario desde DB para obtener email y verificar rol
	user, err := h.usersRepo.GetByID(ctx, *userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	
	// Verificar rol del usuario desde DB
	role, producerID, _ := h.authService.GetUserRole(user.FirebaseUID)

	// Generar nuevo access token con información completa del usuario
	newAccessToken, _, err := h.jwtService.GenerateToken(
		*userID,
		user.Email,
		role,
		producerID,
		24,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate new access token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_at":   time.Now().Add(24 * time.Hour),
	})
}

// OnboardingRequest datos para crear la productora del nuevo usuario
type OnboardingRequest struct {
	ProducerName string `json:"producer_name" binding:"required"`
	LogoURL      string `json:"logo_url"`
	Description  string `json:"description"`
}

// Onboarding completa el registro de un nuevo producer: crea la productora en estado
// 'pending' y asigna el rol 'producer' al usuario. El super_admin la aprobará después.
// Requiere: Authorization header con JWT válido (el token del login inicial).
func (h *Handlers) Onboarding(c *gin.Context) {
	var req OnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	jwtClaims := claims.(*jwt.Claims)
	userID, err := jwtClaims.GetUserID()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	ctx := c.Request.Context()

	// Verificar que el usuario no tenga ya una productora
	existing, err := h.producersRepo.GetByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing producer"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":           "Producer already exists",
			"producer_status": existing.Status,
		})
		return
	}

	// Crear productora en estado pending
	producer := &models.Producer{
		UserID:      userID,
		Name:        req.ProducerName,
		LogoURL:     req.LogoURL,
		Description: req.Description,
		IsActive:    false,
		Status:      "pending",
	}
	if err := h.producersRepo.Create(ctx, producer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create producer"})
		return
	}

	// Asignar rol 'producer' al usuario
	user, _ := h.usersRepo.GetByID(ctx, userID)
	if err := h.authService.GrantRole(ctx, userID, "producer", userID); err != nil {
		log.Printf("Warning: failed to grant producer role: %v", err)
	}

	// Generar nuevos tokens con el rol 'producer' ya asignado
	firebaseUID := ""
	if user != nil {
		firebaseUID = user.FirebaseUID
	}
	_ = firebaseUID

	newToken, _, err := h.jwtService.GenerateToken(userID, jwtClaims.Email, "producer", producer.ID.String(), 24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, _ := jwt.GenerateRefreshToken()
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	_ = h.refreshTokenRepo.Create(ctx, refreshToken, userID, refreshExpiresAt)

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Producer created — awaiting super admin approval",
		"producer_status": "pending",
		"producer_id":     producer.ID.String(),
		"access_token":    newToken,
		"refresh_token":   refreshToken,
		"expires_at":      time.Now().Add(24 * time.Hour),
	})
}

// AcceptInviteRequest payload para aceptar un link de invitación
type AcceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}

// AcceptInvite acepta un link de invitación. El usuario ya debe estar autenticado
// (hizo login con Google). Se le asigna el rol y se vincula al tenant del productor.
func (h *Handlers) AcceptInvite(c *gin.Context) {
	var req AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, _ := c.Get("claims")
	jwtClaims := claims.(*jwt.Claims)
	userID, err := jwtClaims.GetUserID()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	ctx := c.Request.Context()

	inv, err := h.invitationsRepo.GetByToken(ctx, req.Token)
	if err != nil || inv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found or already used"})
		return
	}
	if inv.UsedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Invitation already used"})
		return
	}
	if time.Now().After(inv.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "Invitation has expired"})
		return
	}

	// Vincular usuario al tenant: inserta en producer_members + user_roles
	if err := h.authService.AddMemberToProducer(ctx, inv.ProducerID, userID, inv.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}
	_ = h.invitationsRepo.MarkUsed(ctx, req.Token, userID)

	// Emitir nuevos tokens con el rol asignado
	newToken, _, _ := h.jwtService.GenerateToken(userID, jwtClaims.Email, inv.Role, inv.ProducerID.String(), 24)
	refreshToken, _ := jwt.GenerateRefreshToken()
	_ = h.refreshTokenRepo.Create(ctx, refreshToken, userID, time.Now().Add(7*24*time.Hour))

	// Obtener el status actual del tenant para que el frontend sepa si puede acceder
	producerStatus := "active"
	if p, err2 := h.producersRepo.GetByID(ctx, inv.ProducerID); err2 == nil && p != nil {
		producerStatus = p.Status
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Invitation accepted",
		"producer_id":     inv.ProducerID.String(),
		"role":            inv.Role,
		"producer_status": producerStatus,
		"access_token":    newToken,
		"refresh_token":   refreshToken,
		"expires_at":      time.Now().Add(24 * time.Hour),
	})
}

// GetInviteInfo devuelve información pública de una invitación (para preview antes de aceptar).
func (h *Handlers) GetInviteInfo(c *gin.Context) {
	token := c.Param("token")
	ctx := c.Request.Context()

	inv, err := h.invitationsRepo.GetByToken(ctx, token)
	if err != nil || inv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
		return
	}
	if inv.UsedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Invitation already used"})
		return
	}
	if time.Now().After(inv.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "Invitation has expired"})
		return
	}

	producer, _ := h.producersRepo.GetByID(ctx, inv.ProducerID)
	producerName := ""
	if producer != nil {
		producerName = producer.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":         true,
		"producer_id":   inv.ProducerID.String(),
		"producer_name": producerName,
		"role":          inv.Role,
		"expires_at":    inv.ExpiresAt,
	})
}

// parseProducerUUID convierte el string de producer_id del JWT a uuid.UUID
func parseProducerUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
