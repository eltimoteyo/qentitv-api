package auth

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qenti/qenti/internal/pkg/auth"
	"github.com/qenti/qenti/internal/pkg/jwt"
	"github.com/qenti/qenti/internal/pkg/users"
)

type Handlers struct {
	authService       *auth.Service
	jwtService        *jwt.Service
	refreshTokenRepo  *auth.RefreshTokenRepository
	usersRepo         *users.Repository
}

func NewHandlers(authService *auth.Service, jwtService *jwt.Service, db *sql.DB, usersRepo *users.Repository) *Handlers {
	return &Handlers{
		authService:      authService,
		jwtService:       jwtService,
		refreshTokenRepo: auth.NewRefreshTokenRepository(db),
		usersRepo:        usersRepo,
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
	ID          string `json:"id"`
	Email       string `json:"email"`
	IsPremium   bool   `json:"is_premium"`
	CoinBalance int    `json:"coin_balance"`
	Role        string `json:"role"` // "user" o "admin"
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

	// Verificar si es admin
	isAdmin, _ := h.authService.IsAdmin(user.FirebaseUID)
	
	// Si es modo mock (FirebaseUID es "mock-firebase-uid"), otorgar admin automáticamente
	if user.FirebaseUID == "mock-firebase-uid" && !isAdmin {
		// Otorgar rol de admin al usuario mock para desarrollo
		if err := h.authService.GrantAdminRole(ctx, dbUser.ID, dbUser.ID); err != nil {
			log.Printf("Warning: Failed to grant admin role to mock user: %v", err)
		} else {
			isAdmin = true
			log.Println("✅ Granted admin role to mock user for development")
		}
	}
	
	role := "user"
	if isAdmin {
		role = "admin"
	}

	// Generar access token (24 horas)
	accessToken, _, err := h.jwtService.GenerateToken(
		dbUser.ID,
		dbUser.Email,
		role,
		24, // 24 horas de expiración
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
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
		User: UserInfo{
			ID:          dbUser.ID.String(),
			Email:       dbUser.Email,
			IsPremium:   dbUser.IsPremium,
			CoinBalance: dbUser.CoinBalance,
			Role:        role,
		},
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
	isAdmin, _ := h.authService.IsAdmin(user.FirebaseUID)
	role := "user"
	if isAdmin {
		role = "admin"
	}
	
	// Generar nuevo access token con información completa del usuario
	newAccessToken, _, err := h.jwtService.GenerateToken(
		*userID,
		user.Email,
		role,
		24, // 24 horas de expiración
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

