package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/producers"
)

// MyProducerHandlers gestiona los endpoints de configuración de la propia productora.
type MyProducerHandlers struct {
	repo *producers.Repository
}

func NewMyProducerHandlers(repo *producers.Repository) *MyProducerHandlers {
	return &MyProducerHandlers{repo: repo}
}

// GetMyProducer devuelve los datos de la productora del usuario autenticado.
// Endpoint: GET /admin/my-producer
func (h *MyProducerHandlers) GetMyProducer(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	p, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la productora"})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No tienes una productora vinculada"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// UpdateMyProducerRequest payload para actualizar datos de la propia productora.
type UpdateMyProducerRequest struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

// UpdateMyProducer actualiza nombre, logo y descripción de la productora propia.
// Endpoint: PUT /admin/my-producer
func (h *MyProducerHandlers) UpdateMyProducer(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	p, err := h.repo.GetByUserID(ctx, userID)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No tienes una productora vinculada"})
		return
	}

	var req UpdateMyProducerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido", "details": err.Error()})
		return
	}

	// Solo actualizar campos provistos (no vacíos)
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.LogoURL != "" {
		p.LogoURL = req.LogoURL
	}
	if req.Description != "" {
		p.Description = req.Description
	}

	if err := h.repo.Update(ctx, p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar la productora"})
		return
	}

	c.JSON(http.StatusOK, p)
}
