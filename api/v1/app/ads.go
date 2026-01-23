package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/transactions"
)

// UnlockEpisodeWithAdRequest representa el payload para desbloquear con anuncio
type UnlockEpisodeWithAdRequest struct {
	EpisodeID uuid.UUID `json:"episode_id" binding:"required"`
	AdID      string    `json:"ad_id" binding:"required"` // ID del anuncio visto (para tracking)
}

// UnlockEpisodeWithAd desbloquea un episodio después de ver un anuncio
func (h *Handlers) UnlockEpisodeWithAd(c *gin.Context) {
	ctx := c.Request.Context()
	
	var req UnlockEpisodeWithAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
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
	episode, err := h.episodesRepo.GetByID(ctx, req.EpisodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Episode not found",
		})
		return
	}
	
	// Verificar si ya está desbloqueado
	alreadyUnlocked, err := h.unlocksRepo.IsUnlocked(ctx, uid, req.EpisodeID)
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
	
	// Validar anuncio con el validador
	validation, err := h.adsValidator.ValidateAd(ctx, req.AdID, uid.String(), req.EpisodeID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate ad",
		})
		return
	}
	
	if !validation.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ad",
			"reason": validation.Reason,
		})
		return
	}
	
	// Registrar validación del anuncio (para prevenir reutilización)
	if err := h.adsValidator.RecordAdValidation(ctx, req.AdID, uid.String(), req.EpisodeID.String()); err != nil {
		// Log error pero continuar (no es crítico)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record ad validation",
		})
		return
	}
	
	// TODO: En producción, aquí se integraría con el SDK de ads real (AdMob, Unity Ads, etc.)
	// para verificar que el anuncio fue realmente visto desde el lado del cliente
	
	// Crear desbloqueo con método AD
	unlock := &models.Unlock{
		UserID:    uid,
		EpisodeID: req.EpisodeID,
		Method:    models.UnlockMethodAd,
	}
	
	if err := h.unlocksRepo.Create(ctx, unlock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unlock episode",
		})
		return
	}
	
	// TODO: Registrar visualización del anuncio para analytics
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Episode unlocked successfully",
		"method": "ad",
		"ad_id": req.AdID,
	})
}



// RewardCoinsForAdRequest representa el payload para obtener monedas por ver anuncio
type RewardCoinsForAdRequest struct {
	AdID   string `json:"ad_id" binding:"required"`   // ID del anuncio visto (del SDK)
	AdType string `json:"ad_type" binding:"required"` // Tipo: rewarded, interstitial, banner
}

// RewardCoinsForAd otorga monedas al usuario por ver un anuncio
func (h *Handlers) RewardCoinsForAd(c *gin.Context) {
	ctx := c.Request.Context()

	var req RewardCoinsForAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}
	uid := userID.(uuid.UUID)

	coinsPerAd := 10
	dailyLimit := 10
	hourlyLimit := 3
	cooldownMinutes := 5

	validation, err := h.adsValidator.ValidateAdReward(ctx, req.AdID, uid.String(), cooldownMinutes, dailyLimit, hourlyLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate ad",
		})
		return
	}

	if !validation.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":                "Invalid ad or limit reached",
			"reason":               validation.Reason,
			"daily_limit_remaining": validation.DailyLimitRemaining,
			"hourly_limit_remaining": validation.HourlyLimitRemaining,
			"cooldown_seconds":     validation.CooldownSeconds,
		})
		return
	}

	user, err := h.usersRepo.GetByID(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user",
		})
		return
	}

	coinsToReward := coinsPerAd
	if req.AdType == "interstitial" {
		coinsToReward = coinsPerAd / 2
	} else if req.AdType == "banner" {
		coinsToReward = 1
	}

	newBalance := user.CoinBalance + coinsToReward
	if err := h.usersRepo.UpdateCoinBalance(ctx, uid, newBalance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update coin balance",
		})
		return
	}

	if err := h.adsValidator.RecordAdValidation(ctx, req.AdID, uid.String(), ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record ad validation",
		})
		return
	}

	transactionsRepo := transactions.NewRepository(h.db)
	tx := &transactions.Transaction{
		UserID:    uid,
		Type:      "ad_reward",
		Amount:    coinsToReward,
		EpisodeID: nil,
		Method:    "AD",
	}
	if err := transactionsRepo.Create(ctx, tx); err != nil {
		// Log error pero no fallar
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                "Coins rewarded successfully",
		"coins_earned":            coinsToReward,
		"new_balance":            newBalance,
		"daily_limit_remaining":  validation.DailyLimitRemaining,
		"hourly_limit_remaining": validation.HourlyLimitRemaining,
		"ad_id":                  req.AdID,
		"ad_type":                req.AdType,
	})
}
