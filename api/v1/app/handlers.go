package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/config"
	"github.com/qenti/qenti/internal/pkg/ads"
	"github.com/qenti/qenti/internal/pkg/episodes"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/notifications"
	"github.com/qenti/qenti/internal/pkg/payment"
	"github.com/qenti/qenti/internal/pkg/series"
	"github.com/qenti/qenti/internal/pkg/storage"
	"github.com/qenti/qenti/internal/pkg/transactions"
	"github.com/qenti/qenti/internal/pkg/unlocks"
	"github.com/qenti/qenti/internal/pkg/users"
	"github.com/qenti/qenti/internal/pkg/views"
)

type Handlers struct {
	seriesRepo     *series.Repository
	episodesRepo   *episodes.Repository
	usersRepo      *users.Repository
	unlocksRepo    *unlocks.Repository
	videoProvider  storage.VideoProvider
	adsValidator   *ads.Validator
	paymentService *payment.Service
	notifService   *notifications.Service
	db             *sql.DB // Para acceso a vistas y transacciones
	cfg            *config.Config
}

func NewHandlers(
	seriesRepo *series.Repository,
	episodesRepo *episodes.Repository,
	usersRepo *users.Repository,
	unlocksRepo *unlocks.Repository,
	videoProvider storage.VideoProvider,
	paymentService *payment.Service,
	notifService *notifications.Service,
	db *sql.DB,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		seriesRepo:     seriesRepo,
		episodesRepo:   episodesRepo,
		usersRepo:      usersRepo,
		unlocksRepo:    unlocksRepo,
		videoProvider:  videoProvider,
		adsValidator:   ads.NewValidator(db),
		paymentService: paymentService,
		notifService:   notifService,
		db:             db,
		cfg:            cfg,
	}
}

// GetSeries lista las series disponibles.
// Acepta ?producer_slug=slug para filtrar por tenant (multi-tenancy móvil).
func (h *Handlers) GetSeries(c *gin.Context) {
	ctx := c.Request.Context()

	producerID, err := resolveProducerSlug(ctx, h.db, c.Query("producer_slug"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve producer"})
		return
	}

	seriesList, err := h.seriesRepo.GetAllFiltered(ctx, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch series",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"series": seriesList,
	})
}

// GetSeriesByID obtiene el detalle de una serie
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
	
	series, err := h.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Series not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"series": series,
	})
}

// GetSeriesEpisodes obtiene la lista de episodios de una serie (solo metadatos)
func (h *Handlers) GetSeriesEpisodes(c *gin.Context) {
	ctx := c.Request.Context()
	seriesIDStr := c.Param("id")
	
	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid series ID",
		})
		return
	}
	
	// Verificar que la serie existe
	_, err = h.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Series not found",
		})
		return
	}
	
	// Obtener todos los episodios de la serie
	episodesList, err := h.episodesRepo.GetBySeriesID(ctx, seriesID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch episodes",
		})
		return
	}
	
	// Obtener información del usuario (si está autenticado)
	userID, _ := c.Get("user_id")
	isPremium, _ := c.Get("is_premium")
	
	var unlockedEpisodes map[uuid.UUID]bool
	if userID != nil {
		uid := userID.(uuid.UUID)
		unlockedList, _ := h.unlocksRepo.GetUnlockedEpisodes(ctx, uid)
		unlockedEpisodes = make(map[uuid.UUID]bool)
		for _, epID := range unlockedList {
			unlockedEpisodes[epID] = true
		}
	}
	
	// Construir respuesta con información de acceso (sin URLs de video)
	type EpisodeMetadata struct {
		ID            uuid.UUID `json:"id"`
		EpisodeNumber int       `json:"episode_number"`
		Title         string    `json:"title"`
		Duration      int       `json:"duration"`
		IsFree        bool      `json:"is_free"`
		PriceCoins    int       `json:"price_coins"`
		Locked        bool      `json:"locked"`
	}
	
	var episodes []EpisodeMetadata
	for _, ep := range episodesList {
		item := EpisodeMetadata{
			ID:            ep.ID,
			EpisodeNumber: ep.EpisodeNumber,
			Title:         ep.Title,
			Duration:      ep.Duration,
			IsFree:        ep.IsFree,
			PriceCoins:    ep.PriceCoins,
			Locked:        false,
		}
		
		// Determinar si el episodio está desbloqueado
		isUnlocked := ep.IsFree
		if !isUnlocked && userID != nil {
			if isPremium != nil && isPremium.(bool) {
				isUnlocked = true
			} else {
				isUnlocked = unlockedEpisodes[ep.ID]
			}
		}
		
		if !isUnlocked {
			item.Locked = true
		}
		
		episodes = append(episodes, item)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"episodes": episodes,
	})
}

// GetEpisodeStream obtiene la URL firmada de un episodio si el usuario tiene acceso
func (h *Handlers) GetEpisodeStream(c *gin.Context) {
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
	
	// Verificar acceso del usuario
	userID, exists := c.Get("user_id")
	isPremium, _ := c.Get("is_premium")
	
	hasAccess := episode.IsFree
	var uid uuid.UUID
	
	if !hasAccess && exists {
		uid = userID.(uuid.UUID)
		
		// Verificar si es premium
		if isPremium != nil && isPremium.(bool) {
			hasAccess = true
		} else {
			// Verificar si está desbloqueado
			unlocked, _ := h.unlocksRepo.IsUnlocked(ctx, uid, episodeID)
			hasAccess = unlocked
		}
	}
	
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Episode is locked",
			"episode_id": episodeID,
			"is_free": episode.IsFree,
			"price_coins": episode.PriceCoins,
		})
		return
	}
	
	if episode.VideoIDBunny == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Video not available",
			"message": "Este episodio no tiene video configurado. Por favor verifica en el panel de administración.",
		})
		return
	}
	
	// Generar URL firmada (expira en 1 hora)
	signedURL, err := h.videoProvider.GetPlaybackURL(episode.VideoIDBunny, 60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate video URL",
			"message": fmt.Sprintf("Error al generar URL del video: %v", err),
		})
		return
	}
	
	// Validar que la URL no esté vacía
	if signedURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Empty video URL",
			"message": "La URL del video está vacía. Verifica la configuración de Bunny.net CDN.",
		})
		return
	}
	
	// Registrar vista inicial (solo si el usuario está autenticado)
	if exists {
		viewsRepo := views.NewRepository(h.db)
		go viewsRepo.RecordView(ctx, &uid, episodeID, 0, false)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"video_url": signedURL,
		"expires_in": 3600, // segundos
	})
}

// UnlockEpisode desbloquea un episodio usando monedas
func (h *Handlers) UnlockEpisode(c *gin.Context) {
	ctx := c.Request.Context()
	episodeIDStr := c.Param("id")
	
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid episode ID",
		})
		return
	}
	
	// Obtener información del usuario
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}
	uid := userID.(uuid.UUID)
	
	// Obtener episodio
	episode, err := h.episodesRepo.GetByID(ctx, episodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Episode not found",
		})
		return
	}
	
	// Verificar si ya está desbloqueado
	alreadyUnlocked, err := h.unlocksRepo.IsUnlocked(ctx, uid, episodeID)
	if err == nil && alreadyUnlocked {
		c.JSON(http.StatusOK, gin.H{
			"message": "Episode already unlocked",
		})
		return
	}
	
	// Verificar si es gratis
	if episode.IsFree {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Episode is already free",
		})
		return
	}
	
	// Obtener usuario
	user, err := h.usersRepo.GetByID(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user",
		})
		return
	}
	
	// Verificar balance de monedas
	if user.CoinBalance < episode.PriceCoins {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Insufficient coins",
			"required": episode.PriceCoins,
			"available": user.CoinBalance,
		})
		return
	}
	
	// Crear desbloqueo
	unlock := &models.Unlock{
		UserID:    uid,
		EpisodeID: episodeID,
		Method:    models.UnlockMethodCoin,
	}
	
	if err := h.unlocksRepo.Create(ctx, unlock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unlock episode",
		})
		return
	}
	
	// Descontar monedas
	newBalance := user.CoinBalance - episode.PriceCoins
	if err := h.usersRepo.UpdateCoinBalance(ctx, uid, newBalance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update coin balance",
		})
		return
	}
	
	// Registrar transacción
	transactionsRepo := transactions.NewRepository(h.db)
	tx := &transactions.Transaction{
		UserID:    uid,
		Type:      "unlock",
		Amount:    -episode.PriceCoins,
		EpisodeID: &episodeID,
		Method:    models.UnlockMethodCoin,
	}
	transactionsRepo.Create(ctx, tx)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Episode unlocked successfully",
		"remaining_coins": newBalance,
	})
}

// GetUserProfile obtiene el perfil del usuario
func (h *Handlers) GetUserProfile(c *gin.Context) {
	ctx := c.Request.Context()
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}
	uid := userID.(uuid.UUID)
	
	user, err := h.usersRepo.GetByID(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user profile",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

