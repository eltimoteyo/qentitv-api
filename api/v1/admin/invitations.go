package admin

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/invitations"
	"github.com/qenti/qenti/internal/pkg/jwt"
	"github.com/qenti/qenti/internal/pkg/models"
)

type InvitationsHandlers struct {
	invitationsRepo *invitations.Repository
}

func NewInvitationsHandlers(invitationsRepo *invitations.Repository) *InvitationsHandlers {
	return &InvitationsHandlers{invitationsRepo: invitationsRepo}
}

// CreateInviteRequest payload para generar un link de invitación
type CreateInviteRequest struct {
	// Role del usuario invitado dentro del tenant. Por defecto 'producer' (acceso completo).
	// En el futuro podrá ser 'editor', 'analyst', etc.
	Role      string `json:"role"`
	ExpiresIn int    `json:"expires_in_days"` // Días hasta expiración. Default: 7.
}

// CreateInvite genera un link de invitación copiable para un tenant.
// Solo accesible por el tenant admin (producer) o super_admin.
// El link contiene el token que el invitado usa para unirse al tenant.
func (h *InvitationsHandlers) CreateInvite(c *gin.Context) {
	claims, _ := c.Get("claims")
	jwtClaims := claims.(*jwt.Claims)

	// El producer_id viene del JWT del tenant admin; si es super_admin lo toma del param
	producerIDStr := jwtClaims.ProducerID
	if producerIDStr == "" {
		// super_admin debe especificar el producer_id en el body
		producerIDStr = c.Query("producer_id")
	}
	producerID, err := uuid.Parse(producerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "producer_id required"})
		return
	}

	var req CreateInviteRequest
	_ = c.ShouldBindJSON(&req)
	if req.Role == "" {
		req.Role = "producer"
	}
	if req.ExpiresIn <= 0 {
		req.ExpiresIn = 7
	}

	userID, _ := jwtClaims.GetUserID()
	ctx := c.Request.Context()

	inv, err := h.invitationsRepo.Create(ctx, producerID, req.Role, &userID, time.Duration(req.ExpiresIn)*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":       inv.Token,
		"expires_at":  inv.ExpiresAt,
		"role":        inv.Role,
		"producer_id": producerID.String(),
		// El frontend construye el link completo: https://admin.qenti.tv/invite/{token}
	})
}

// ListInvites lista las invitaciones activas del tenant del producer autenticado.
func (h *InvitationsHandlers) ListInvites(c *gin.Context) {
	claims, _ := c.Get("claims")
	jwtClaims := claims.(*jwt.Claims)

	producerID, err := uuid.Parse(jwtClaims.ProducerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No producer associated with this account"})
		return
	}

	ctx := c.Request.Context()
	list, err := h.invitationsRepo.ListByProducer(ctx, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
		return
	}
	if list == nil {
		list = []models.Invitation{}
	}
	c.JSON(http.StatusOK, gin.H{"invitations": list})
}
