package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/bunny"
	"github.com/qenti/qenti/internal/pkg/episodes"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/series"
)

type Handlers struct {
	seriesRepo   *series.Repository
	episodesRepo *episodes.Repository
	bunnyService *bunny.Service
}

func NewHandlers(
	seriesRepo *series.Repository,
	episodesRepo *episodes.Repository,
	bunnyService *bunny.Service,
) *Handlers {
	return &Handlers{
		seriesRepo:   seriesRepo,
		episodesRepo: episodesRepo,
		bunnyService: bunnyService,
	}
}

// CreateSeriesRequest representa el payload para crear una serie
type CreateSeriesRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	HorizontalPoster string `json:"horizontal_poster"`
	VerticalPoster  string `json:"vertical_poster"`
	IsActive        bool   `json:"is_active"`
}

// UpdateSeriesRequest representa el payload para actualizar una serie
type UpdateSeriesRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	HorizontalPoster string `json:"horizontal_poster"`
	VerticalPoster  string `json:"vertical_poster"`
	IsActive        *bool  `json:"is_active"`
}

// CreateEpisodeRequest representa el payload para crear un episodio
type CreateEpisodeRequest struct {
	SeriesID      uuid.UUID `json:"series_id" binding:"required"`
	EpisodeNumber int       `json:"episode_number" binding:"required"`
	Title         string    `json:"title" binding:"required"`
	Duration      int       `json:"duration"`
	IsFree        bool      `json:"is_free"`
	PriceCoins    int       `json:"price_coins"`
}

// UpdateEpisodeRequest representa el payload para actualizar un episodio
type UpdateEpisodeRequest struct {
	Title      string `json:"title"`
	Duration   int    `json:"duration"`
	IsFree     *bool  `json:"is_free"`
	PriceCoins int    `json:"price_coins"`
}

// GetSeries lista todas las series (admin)
func (h *Handlers) GetSeries(c *gin.Context) {
	ctx := c.Request.Context()
	
	seriesList, err := h.seriesRepo.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch series",
		})
		return
	}
	
	// Asegurar que siempre devolvemos un array (no nil)
	if seriesList == nil {
		seriesList = []models.Series{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"series": seriesList,
	})
}

// GetSeriesByID obtiene una serie por ID
func (h *Handlers) GetSeriesByID(c *gin.Context) {
	ctx := c.Request.Context()
	seriesIDStr := c.Param("id")
	
	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid series ID",
		})
		return
	}
	
	s, err := h.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Series not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"series": s,
	})
}

// CreateSeries crea una nueva serie
func (h *Handlers) CreateSeries(c *gin.Context) {
	ctx := c.Request.Context()
	
	var req CreateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	series := &models.Series{
		Title:            req.Title,
		Description:      req.Description,
		HorizontalPoster: req.HorizontalPoster,
		VerticalPoster:   req.VerticalPoster,
		IsActive:         req.IsActive,
	}
	
	if err := h.seriesRepo.Create(ctx, series); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create series",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"series": series,
	})
}

// UpdateSeries actualiza una serie existente
func (h *Handlers) UpdateSeries(c *gin.Context) {
	ctx := c.Request.Context()
	seriesIDStr := c.Param("id")
	
	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid series ID",
		})
		return
	}
	
	var req UpdateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Obtener serie existente
	series, err := h.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Series not found",
		})
		return
	}
	
	// Actualizar campos
	if req.Title != "" {
		series.Title = req.Title
	}
	if req.Description != "" {
		series.Description = req.Description
	}
	if req.HorizontalPoster != "" {
		series.HorizontalPoster = req.HorizontalPoster
	}
	if req.VerticalPoster != "" {
		series.VerticalPoster = req.VerticalPoster
	}
	if req.IsActive != nil {
		series.IsActive = *req.IsActive
	}
	
	if err := h.seriesRepo.Update(ctx, series); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update series",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"series": series,
	})
}

// DeleteSeries elimina una serie (soft delete)
func (h *Handlers) DeleteSeries(c *gin.Context) {
	ctx := c.Request.Context()
	seriesIDStr := c.Param("id")
	
	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid series ID",
		})
		return
	}
	
	if err := h.seriesRepo.Delete(ctx, seriesID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete series",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Series deleted successfully",
	})
}

// CreateEpisode crea un nuevo episodio
func (h *Handlers) CreateEpisode(c *gin.Context) {
	ctx := c.Request.Context()
	
	var req CreateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	episode := &models.Episode{
		SeriesID:      req.SeriesID,
		EpisodeNumber: req.EpisodeNumber,
		Title:         req.Title,
		Duration:      req.Duration,
		IsFree:        req.IsFree,
		PriceCoins:    req.PriceCoins,
	}
	
	if err := h.episodesRepo.Create(ctx, episode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create episode",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"episode": episode,
	})
}

// UpdateEpisode actualiza un episodio existente
func (h *Handlers) UpdateEpisode(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	var req UpdateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Obtener episodio existente
	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Episode not found",
		})
		return
	}
	
	// Actualizar campos
	if req.Title != "" {
		episode.Title = req.Title
	}
	if req.Duration > 0 {
		episode.Duration = req.Duration
	}
	if req.IsFree != nil {
		episode.IsFree = *req.IsFree
	}
	if req.PriceCoins >= 0 {
		episode.PriceCoins = req.PriceCoins
	}
	
	if err := h.episodesRepo.Update(ctx, episode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update episode",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"episode": episode,
	})
}

// GetUploadURL genera una URL presignada para subir un video a Bunny.net
// Endpoint: POST /admin/episodes/{id}/upload-url
func (h *Handlers) GetUploadURL(c *gin.Context) {
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	// Obtener episodio para usar su título
	ctx := c.Request.Context()
	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Episode not found",
		})
		return
	}
	
	// Generar URL presignada y video ID
	uploadResult, err := h.bunnyService.PresignedUploadURL(episode.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate upload URL",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadResult.UploadURL,
		"video_id":   uploadResult.VideoID,
		"episode_id": episodeID,
	})
}

// UploadVideo recibe el archivo de video y lo sube a Bunny.net
// Endpoint: POST /admin/episodes/{id}/upload
func (h *Handlers) UploadVideo(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	// Obtener episodio
	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Episode not found",
		})
		return
	}
	
	// Obtener archivo del multipart form
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No video file provided",
			"details": err.Error(),
		})
		return
	}
	
	// Abrir archivo
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open file",
			"details": err.Error(),
		})
		return
	}
	defer src.Close()
	
	// Si el episodio ya tiene un video_id, usarlo; si no, crear uno nuevo
	var videoID string
	if episode.VideoIDBunny != "" {
		videoID = episode.VideoIDBunny
	} else {
		// Crear nuevo video en Bunny.net
		uploadResult, err := h.bunnyService.PresignedUploadURL(episode.Title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create video in Bunny.net",
				"details": err.Error(),
			})
			return
		}
		videoID = uploadResult.VideoID
	}
	
	// Subir archivo a Bunny.net
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4" // Default
	}
	
	if err := h.bunnyService.UploadVideo(videoID, src, contentType, file.Size); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload video to Bunny.net",
			"details": err.Error(),
		})
		return
	}
	
	// Actualizar video_id_bunny en el episodio
	if err := h.episodesRepo.UpdateVideoID(ctx, episodeID, videoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update episode video ID",
			"details": err.Error(),
		})
		return
	}
	
	// Marcar video como completado en Bunny (opcional, para re-encoding)
	h.bunnyService.CompleteUpload(videoID) // Ignorar error, no es crítico
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Video uploaded successfully",
		"video_id": videoID,
	})
}

// CompleteUpload marca un episodio como completado después de la subida
// Endpoint: POST /admin/episodes/{id}/complete
type CompleteUploadRequest struct {
	VideoIDBunny string `json:"video_id_bunny" binding:"required"`
}

func (h *Handlers) CompleteUpload(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	var req CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Actualizar video_id_bunny en el episodio
	if err := h.episodesRepo.UpdateVideoID(ctx, episodeID, req.VideoIDBunny); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update episode video ID",
		})
		return
	}
	
	// Marcar video como completado en Bunny (opcional, para re-encoding)
	if err := h.bunnyService.CompleteUpload(req.VideoIDBunny); err != nil {
		// Log error pero no fallar la operación
		c.JSON(http.StatusOK, gin.H{
			"message": "Episode updated successfully, but re-encoding may have failed",
			"warning": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload completed successfully",
	})
}

// ValidateBunnyConnection valida la conexión con Bunny.net
// Endpoint: GET /admin/validate/bunny
func (h *Handlers) ValidateBunnyConnection(c *gin.Context) {
	if err := h.bunnyService.ValidateConnection(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"error":   err.Error(),
			"message": "No se pudo conectar con Bunny.net. Verifica tus credenciales.",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Conexión con Bunny.net exitosa",
	})
}

// GetEpisodes lista episodios con filtros opcionales
func (h *Handlers) GetEpisodes(c *gin.Context) {
	ctx := c.Request.Context()
	
	seriesIDStr := c.Query("series_id")
	var seriesID *uuid.UUID
	
	if seriesIDStr != "" {
		parsedID, err := uuid.Parse(seriesIDStr)
		if err == nil {
			seriesID = &parsedID
		}
	}
	
	episodes, err := h.episodesRepo.GetAll(ctx, seriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch episodes",
		})
		return
	}
	
	// Asegurar que siempre devolvemos un array (no nil)
	if episodes == nil {
		episodes = []models.Episode{}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"episodes": episodes,
	})
}

// GetEpisodeByID obtiene el detalle de un episodio
func (h *Handlers) GetEpisodeByID(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Episode not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"episode": episode,
	})
}

// DeleteEpisode elimina un episodio
func (h *Handlers) DeleteEpisode(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	if err := h.episodesRepo.Delete(ctx, episodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete episode",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Episode deleted successfully",
	})
}

