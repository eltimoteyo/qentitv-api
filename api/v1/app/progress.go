package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/views"
)

// UpdateWatchProgressRequest payload para guardar progreso de reproducción.
type UpdateWatchProgressRequest struct {
	WatchedSeconds int  `json:"watched_seconds" binding:"required,min=0"`
	Completed      bool `json:"completed"`
}

// UpdateWatchProgress guarda (upsert) cuántos segundos lleva el usuario en un episodio.
// Se llama durante reproducción (ej. cada 30 s) o al pausar/salir.
//
// POST /api/v1/app/episodes/:id/progress
func (h *Handlers) UpdateWatchProgress(c *gin.Context) {
	ctx := c.Request.Context()

	episodeIDStr := c.Param("id")
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	uid := userID.(uuid.UUID)

	var req UpdateWatchProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	viewsRepo := views.NewRepository(h.db)
	if err := viewsRepo.UpdateWatchProgress(ctx, uid, episodeID, req.WatchedSeconds, req.Completed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Progress saved",
		"watched_seconds": req.WatchedSeconds,
		"completed":       req.Completed,
	})
}

// GetContinueWatching devuelve las series en curso del usuario (no finalizadas).
// Máximo 10 items, ordenadas por última actividad DESC.
//
// GET /api/v1/app/continue-watching
func (h *Handlers) GetContinueWatching(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	uid := userID.(uuid.UUID)

	viewsRepo := views.NewRepository(h.db)
	items, err := viewsRepo.GetContinueWatching(ctx, uid, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch continue watching"})
		return
	}

	if items == nil {
		items = []views.ContinueWatchingItem{}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
