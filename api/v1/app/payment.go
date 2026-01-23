package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetSubscriptionStatus retorna el estado de suscripción del usuario
func (h *Handlers) GetSubscriptionStatus(c *gin.Context) {
	ctx := c.Request.Context()
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}
	uid := userID.(uuid.UUID)
	
	// Obtener usuario
	user, err := h.usersRepo.GetByID(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user",
		})
		return
	}
	
	// Verificar estado desde DB
	status := "inactive"
	if user.IsPremium {
		status = "active"
	}
	
	// Si está configurado RevenueCat, verificar también allí
	// Por ahora usamos solo el estado de DB
	// TODO: Integrar con RevenueCat API para obtener fecha de expiración real
	
	c.JSON(http.StatusOK, gin.H{
		"status":      status,
		"is_premium":  user.IsPremium,
		"expires_at":  nil, // TODO: Obtener desde RevenueCat cuando esté configurado
		"auto_renew":  false, // TODO: Obtener desde RevenueCat
	})
}

// SubscriptionOffer representa un plan de suscripción
type SubscriptionOffer struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Duration    int      `json:"duration"` // en días
	TrialDays   int      `json:"trial_days"`
	Features    []string `json:"features"`
}

// GetOffer retorna los planes de suscripción disponibles
func (h *Handlers) GetOffer(c *gin.Context) {
	// Planes predefinidos (en producción, estos deberían venir de RevenueCat o DB)
	offers := []SubscriptionOffer{
		{
			ID:          "premium_monthly",
			Name:        "Premium Mensual",
			Description: "Acceso ilimitado a todo el contenido",
			Price:       9.99,
			Currency:    "USD",
			Duration:    30,
			TrialDays:   7,
			Features: []string{
				"Acceso ilimitado a todos los episodios",
				"Sin anuncios",
				"Contenido exclusivo",
				"Descarga para ver offline",
			},
		},
		{
			ID:          "premium_yearly",
			Name:        "Premium Anual",
			Description: "Acceso ilimitado con descuento anual",
			Price:       79.99,
			Currency:    "USD",
			Duration:    365,
			TrialDays:   7,
			Features: []string{
				"Acceso ilimitado a todos los episodios",
				"Sin anuncios",
				"Contenido exclusivo",
				"Descarga para ver offline",
				"Ahorra 33% vs mensual",
			},
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"offers": offers,
	})
}

