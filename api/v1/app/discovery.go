package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qenti/qenti/internal/pkg/models"
)

// GetTrending devuelve series ordenadas por score de actividad real (vistas + unlocks × 2)
// en los últimos 7 días. Endpoint público.
//
// GET /api/v1/app/trending?limit=20&days=7&producer_slug=slug
func (h *Handlers) GetTrending(c *gin.Context) {
	ctx := c.Request.Context()

	limit := 20
	days := 7

	producerID, err := resolveProducerSlug(ctx, h.db, c.Query("producer_slug"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve producer"})
		return
	}

	seriesList, err := h.seriesRepo.GetTrendingFiltered(ctx, limit, days, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch trending series",
		})
		return
	}

	if seriesList == nil {
		seriesList = []models.Series{}
	}

	c.JSON(http.StatusOK, gin.H{
		"series":      seriesList,
		"days_window": days,
	})
}

// GetMostViewed devuelve series ordenadas por total de vistas histórico (all-time).
//
// GET /api/v1/app/most-viewed?limit=30
func (h *Handlers) GetMostViewed(c *gin.Context) {
	ctx := c.Request.Context()

	limit := 30

	seriesList, err := h.seriesRepo.GetMostViewed(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch most viewed"})
		return
	}

	if seriesList == nil {
		seriesList = []models.Series{}
	}

	c.JSON(http.StatusOK, gin.H{"series": seriesList})
}

// GetNewReleases devuelve series publicadas recientemente (por defecto últimos 30 días).
//
// GET /api/v1/app/new-releases?limit=30&days=30
func (h *Handlers) GetNewReleases(c *gin.Context) {
	ctx := c.Request.Context()

	limit := 30
	days := 30

	seriesList, err := h.seriesRepo.GetNewReleases(ctx, limit, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch new releases"})
		return
	}

	// Si no hay series nuevas en el período, devolver las más recientes sin restricción
	if len(seriesList) == 0 {
		seriesList, _ = h.seriesRepo.GetAll(ctx)
		if seriesList == nil {
			seriesList = []models.Series{}
		}
	}

	c.JSON(http.StatusOK, gin.H{"series": seriesList})
}

// Search busca series por título o descripción (case-insensitive, mínimo 2 chars).
//
// GET /api/v1/app/search?q=drama&producer_slug=slug
func (h *Handlers) Search(c *gin.Context) {
	ctx := c.Request.Context()

	q := strings.TrimSpace(c.Query("q"))
	if len(q) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query must be at least 2 characters",
		})
		return
	}

	limit := 30

	producerID, err := resolveProducerSlug(ctx, h.db, c.Query("producer_slug"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve producer"})
		return
	}

	seriesList, err := h.seriesRepo.SearchFiltered(ctx, q, limit, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed",
		})
		return
	}

	if seriesList == nil {
		seriesList = []models.Series{}
	}

	c.JSON(http.StatusOK, gin.H{
		"series": seriesList,
		"query":  q,
		"count":  len(seriesList),
	})
}
