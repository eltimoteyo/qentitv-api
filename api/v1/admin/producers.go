package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/notifications"
	"github.com/qenti/qenti/internal/pkg/producers"
)

type ProducersHandlers struct {
	producersRepo *producers.Repository
	notifService  *notifications.Service
}

func NewProducersHandlers(producersRepo *producers.Repository, notifService *notifications.Service) *ProducersHandlers {
	return &ProducersHandlers{
		producersRepo: producersRepo,
		notifService:  notifService,
	}
}

// CreateProducerRequest representa el payload para crear un productor
type CreateProducerRequest struct {
	// UserEmail o UserID del usuario que será el productor
	UserID      string `json:"user_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}

// UpdateProducerRequest representa el payload para actualizar un productor
type UpdateProducerRequest struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

// GetProducers lista todos los productores (super_admin only)
func (h *ProducersHandlers) GetProducers(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.producersRepo.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch producers"})
		return
	}
	if list == nil {
		list = []models.ProducerWithEmail{}
	}
	c.JSON(http.StatusOK, gin.H{"producers": list})
}

// GetProducerByID obtiene un productor por ID
func (h *ProducersHandlers) GetProducerByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid producer ID"})
		return
	}
	p, err := h.producersRepo.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Producer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"producer": p})
}

// CreateProducer crea un nuevo productor y asigna el rol al usuario
func (h *ProducersHandlers) CreateProducer(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateProducerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id (must be UUID)"})
		return
	}

	p := &models.Producer{
		UserID:      userID,
		Name:        req.Name,
		Slug:        req.Slug,
		LogoURL:     req.LogoURL,
		Description: req.Description,
		IsActive:    true,
		// Super_admin crea productores ya aprobados (sin pasar por flujo de onboarding)
		Status:      "active",
	}

	if err := h.producersRepo.Create(ctx, p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create producer", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"producer": p})
}

// UpdateProducer actualiza un productor existente
func (h *ProducersHandlers) UpdateProducer(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid producer ID"})
		return
	}

	existing, err := h.producersRepo.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Producer not found"})
		return
	}

	var req UpdateProducerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	p := &existing.Producer
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.LogoURL != "" {
		p.LogoURL = req.LogoURL
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	if err := h.producersRepo.Update(ctx, p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update producer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"producer": p})
}

// DeleteProducer elimina un productor
func (h *ProducersHandlers) DeleteProducer(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid producer ID"})
		return
	}
	if err := h.producersRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete producer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Producer deleted successfully"})
}

// ApproveProducer aprueba un tenant pendiente, activándolo para acceso completo al panel.
func (h *ProducersHandlers) ApproveProducer(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid producer ID"})
		return
	}

	// Obtener el productor antes de cambiar estado (necesitamos user_id y nombre)
	producer, err := h.producersRepo.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Producer not found"})
		return
	}

	if err := h.producersRepo.SetStatus(ctx, id, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve producer"})
		return
	}

	// Enviar push notification al productor de forma asíncrona (best-effort)
	if h.notifService != nil {
		go h.notifService.NotifyProducerApproved(ctx, producer.UserID, producer.Name)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Producer approved", "status": "active"})
}

// SuspendProducer suspende un tenant (acceso bloqueado hasta reactivación).
func (h *ProducersHandlers) SuspendProducer(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid producer ID"})
		return
	}
	if err := h.producersRepo.SetStatus(ctx, id, "suspended"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suspend producer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Producer suspended", "status": "suspended"})
}

