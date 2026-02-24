package admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/episodes"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/notifications"
	"github.com/qenti/qenti/internal/pkg/series"
	"github.com/qenti/qenti/internal/pkg/storage"
)

type Handlers struct {
	seriesRepo     *series.Repository
	episodesRepo   *episodes.Repository
	videoProvider  storage.VideoProvider
	notifService   *notifications.Service
	// maxFileSizeMB límite de tamaño en MB para uploads (configurable vía VideoUploadConfig)
	maxFileSizeMB  int64
	warnFileSizeMB int64
	// Cliff pricing: precio automático según posición del episodio
	cliffStart     int // primer ep en precio alto (ej. 8)
	basePrice      int // precio ep 1..(cliffStart-1) (ej. 10 monedas)
	cliffPrice     int // precio ep >= cliffStart (ej. 20 monedas)
}

func NewHandlers(
	seriesRepo *series.Repository,
	episodesRepo *episodes.Repository,
	videoProvider storage.VideoProvider,
	notifService *notifications.Service,
	maxFileSizeMB int64,
	warnFileSizeMB int64,
	cliffStart int,
	basePrice int,
	cliffPrice int,
) *Handlers {
	return &Handlers{
		seriesRepo:     seriesRepo,
		episodesRepo:   episodesRepo,
		videoProvider:  videoProvider,
		notifService:   notifService,
		maxFileSizeMB:  maxFileSizeMB,
		warnFileSizeMB: warnFileSizeMB,
		cliffStart:     cliffStart,
		basePrice:      basePrice,
		cliffPrice:     cliffPrice,
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

// producerIDFromContext extrae el producer_id del contexto gin (vacío para super_admin).
func producerIDFromContext(c *gin.Context) *uuid.UUID {
	pidStr, _ := c.Get("producer_id")
	s, ok := pidStr.(string)
	if !ok || s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

// GetSeries lista las series del panel: filtra por producer si aplica
func (h *Handlers) GetSeries(c *gin.Context) {
	ctx := c.Request.Context()
	producerID := producerIDFromContext(c)

	seriesList, err := h.seriesRepo.GetAllAdmin(ctx, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch series"})
		return
	}
	if seriesList == nil {
		seriesList = []models.Series{}
	}
	c.JSON(http.StatusOK, gin.H{"series": seriesList})
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

// CreateSeries crea una nueva serie y la asigna al productor si aplica
func (h *Handlers) CreateSeries(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	series := &models.Series{
		Title:            req.Title,
		Description:      req.Description,
		HorizontalPoster: req.HorizontalPoster,
		VerticalPoster:   req.VerticalPoster,
		IsActive:         req.IsActive,
		ProducerID:       producerIDFromContext(c), // nil para super_admin
	}

	if err := h.seriesRepo.Create(ctx, series); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create series"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"series": series})
}

// UpdateSeries actualiza una serie existente verificando propiedad si es producer
func (h *Handlers) UpdateSeries(c *gin.Context) {
	ctx := c.Request.Context()
	seriesIDStr := c.Param("id")

	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	// Verificar propiedad
	pidStr, _ := c.Get("producer_id")
	if owns, err := h.seriesRepo.BelongsToProducer(ctx, seriesID, fmt.Sprintf("%v", pidStr)); err != nil || !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "Series not found or not owned by you"})
		return
	}

	var req UpdateSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
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

// DeleteSeries elimina una serie (soft delete) verificando propiedad
func (h *Handlers) DeleteSeries(c *gin.Context) {
	ctx := c.Request.Context()
	seriesIDStr := c.Param("id")

	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	// Verificar propiedad
	pidStr, _ := c.Get("producer_id")
	if owns, err := h.seriesRepo.BelongsToProducer(ctx, seriesID, fmt.Sprintf("%v", pidStr)); err != nil || !owns {
		c.JSON(http.StatusForbidden, gin.H{"error": "Series not found or not owned by you"})
		return
	}

	if err := h.seriesRepo.Delete(ctx, seriesID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete series"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Series deleted successfully"})
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

	// Cliff pricing: si el admin no especificó precio y el episodio no es gratuito,
	// asignamos el precio automáticamente según la posición del episodio.
	if !req.IsFree && req.PriceCoins == 0 {
		if h.cliffStart > 0 && req.EpisodeNumber >= h.cliffStart {
			episode.PriceCoins = h.cliffPrice
		} else {
			episode.PriceCoins = h.basePrice
		}
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

// GetUploadURL genera una URL de upload directo al proveedor de video activo.
// El cliente hace PUT directo a esa URL (no pasa por el servidor).
// Endpoint: POST /admin/episodes/{id}/upload-url
func (h *Handlers) GetUploadURL(c *gin.Context) {
	episodeIDStr := c.Param("id")

	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	ctx := c.Request.Context()
	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Episode not found"})
		return
	}

	uploadResult, err := h.videoProvider.CreateVideo(episode.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate upload URL",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadResult.UploadURL,
		"video_id":   uploadResult.ExternalID,
		"episode_id": episodeID,
		"provider":   h.videoProvider.ProviderName(),
	})
}

// UploadVideo recibe el archivo de video, lo valida y lo sube al proveedor activo.
// Endpoint: POST /admin/episodes/{id}/upload
func (h *Handlers) UploadVideo(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")

	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Episode not found"})
		return
	}

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided", "details": err.Error()})
		return
	}

	// --- Validación de tipo MIME ---
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}
	if !strings.HasPrefix(contentType, "video/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Tipo de archivo no permitido: %s. Solo se aceptan videos.", contentType),
		})
		return
	}

	// --- Validación de tamaño ---
	maxBytes := h.maxFileSizeMB * 1024 * 1024
	if file.Size > maxBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf(
				"El archivo pesa %.1f MB y supera el límite de %d MB. Comprimir el video antes de subir.",
				float64(file.Size)/1024/1024, h.maxFileSizeMB,
			),
			"max_mb":  h.maxFileSizeMB,
			"size_mb": float64(file.Size) / 1024 / 1024,
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file", "details": err.Error()})
		return
	}
	defer src.Close()

	// Crear o reutilizar el ID externo del video
	var externalID string
	if episode.VideoIDBunny != "" {
		externalID = episode.VideoIDBunny
	} else {
		uploadResult, err := h.videoProvider.CreateVideo(episode.Title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   fmt.Sprintf("Failed to create video in %s", h.videoProvider.ProviderName()),
				"details": err.Error(),
			})
			return
		}
		externalID = uploadResult.ExternalID
	}

	if err := h.videoProvider.UploadVideo(externalID, src, contentType, file.Size); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Failed to upload video to %s", h.videoProvider.ProviderName()),
			"details": err.Error(),
		})
		return
	}

	if err := h.episodesRepo.UpdateVideoID(ctx, episodeID, externalID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update episode video ID", "details": err.Error()})
		return
	}

	h.videoProvider.CompleteUpload(externalID) // no crítico

	// Avisar si el video es más grande de lo recomendado
	warning := ""
	warnBytes := h.warnFileSizeMB * 1024 * 1024
	if file.Size > warnBytes {
		warning = fmt.Sprintf(
			"El video pesa %.1f MB. Para micro-dramas verticales se recomienda < %d MB. Considerar re-comprimir con H.264/H.265 CRF 28-30.",
			float64(file.Size)/1024/1024, h.warnFileSizeMB,
		)
	}

	resp := gin.H{
		"message":   "Video uploaded successfully",
		"video_id":  externalID,
		"provider":  h.videoProvider.ProviderName(),
		"size_mb":   float64(file.Size) / 1024 / 1024,
	}
	if warning != "" {
		resp["warning"] = warning
	}
	c.JSON(http.StatusOK, resp)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid episode ID"})
		return
	}

	var req CompleteUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := h.episodesRepo.UpdateVideoID(ctx, episodeID, req.VideoIDBunny); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update episode video ID"})
		return
	}

	h.videoProvider.CompleteUpload(req.VideoIDBunny) // no crítico

	// Ahora que el video está listo, notificar a los fans de la serie (best-effort)
	go func() {
		if h.notifService == nil {
			return
		}
		bgCtx := context.Background()
		ep, err := h.episodesRepo.GetByID(bgCtx, episodeID)
		if err != nil || ep == nil {
			return
		}
		s, err := h.seriesRepo.GetByID(bgCtx, ep.SeriesID)
		seriesTitle := "Nueva actualización"
		if err == nil && s != nil {
			seriesTitle = s.Title
		}
		h.notifService.NotifyNewEpisode(bgCtx, ep.SeriesID, seriesTitle, ep.EpisodeNumber, ep.Title)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":  "Upload completed successfully",
		"provider": h.videoProvider.ProviderName(),
	})
}

// ValidateStorageConnection valida la conexión con el proveedor de video activo.
// Endpoint: GET /admin/validate/bunny
func (h *Handlers) ValidateBunnyConnection(c *gin.Context) {
	if err := h.videoProvider.ValidateConnection(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "error",
			"provider": h.videoProvider.ProviderName(),
			"error":    err.Error(),
			"message":  "No se pudo conectar con el proveedor de video. Verifica tus credenciales.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"provider": h.videoProvider.ProviderName(),
		"message":  "Conexión con el proveedor de video exitosa.",
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

