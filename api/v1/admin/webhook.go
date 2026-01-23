package admin

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/qenti/qenti/internal/pkg/payment"
	"github.com/qenti/qenti/internal/pkg/users"
)

type WebhookHandlers struct {
	paymentService *payment.Service
	usersRepo      *users.Repository
}

func NewWebhookHandlers(
	paymentService *payment.Service,
	usersRepo *users.Repository,
) *WebhookHandlers {
	return &WebhookHandlers{
		paymentService: paymentService,
		usersRepo:      usersRepo,
	}
}

// HandleRevenueCatWebhook procesa webhooks de RevenueCat para actualizar estado premium
func (h *WebhookHandlers) HandleRevenueCatWebhook(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Leer body completo
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}
	
	// Obtener firma del header
	signature := c.GetHeader("Authorization")
	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Missing authorization signature",
		})
		return
	}
	
	// Procesar webhook
	event, err := h.paymentService.ProcessWebhook(body, signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to process webhook",
			"details": err.Error(),
		})
		return
	}
	
	// Procesar según tipo de evento
	switch event.Event.Type {
	case "INITIAL_PURCHASE", "RENEWAL":
		// Usuario compró o renovó suscripción
		// Obtener usuario por app_user_id (que debería ser firebase_uid)
		user, err := h.usersRepo.GetByFirebaseUID(ctx, event.Event.AppUserID)
		if err != nil {
			// Si no existe, crear usuario (o manejar error según lógica de negocio)
			c.JSON(http.StatusOK, gin.H{
				"message": "Webhook processed, but user not found",
			})
			return
		}
		
		// Actualizar estado premium
		if err := h.usersRepo.UpdatePremiumStatus(ctx, user.ID, true); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update premium status",
			})
			return
		}
		
	case "CANCELLATION", "EXPIRATION":
		// Usuario canceló o expiró suscripción
		user, err := h.usersRepo.GetByFirebaseUID(ctx, event.Event.AppUserID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "Webhook processed, but user not found",
			})
			return
		}
		
		// Actualizar estado premium a false
		if err := h.usersRepo.UpdatePremiumStatus(ctx, user.ID, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update premium status",
			})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook processed successfully",
	})
}

