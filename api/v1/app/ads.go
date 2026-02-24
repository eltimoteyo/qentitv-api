package app

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
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

	// Obtener configuración desde cfg (sin valores hardcodeados)
	coinsPerAd := h.cfg.AdReward.CoinsPerAd
	hourlyLimit := h.cfg.AdReward.HourlyLimit
	cooldownMinutes := h.cfg.AdReward.CooldownMinutes

	// Límite diario según país: Tier-A (high eCPM) recibe más anuncios permitidos.
	// Cloudflare rellena CF-IPCountry automáticamente; el SDK móvil puede enviar X-Country.
	dailyLimit := h.cfg.AdReward.DailyLimit
	country := c.GetHeader("CF-IPCountry")
	if country == "" {
		country = c.GetHeader("X-Country")
	}
	if country != "" {
		for _, tc := range h.cfg.AdTier.TierACountries {
			if tc == country {
				dailyLimit = h.cfg.AdTier.TierADailyLimit
				break
			}
		}
	}

	validation, err := h.adsValidator.ValidateAdReward(ctx, req.AdID, uid.String(), cooldownMinutes, dailyLimit, hourlyLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate ad",
		})
		return
	}

	if !validation.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":                  "Invalid ad or limit reached",
			"reason":                 validation.Reason,
			"daily_limit_remaining":  validation.DailyLimitRemaining,
			"hourly_limit_remaining": validation.HourlyLimitRemaining,
			"cooldown_seconds":       validation.CooldownSeconds,
		})
		return
	}

	// Calcular monedas según tipo de anuncio
	coinsToReward := coinsPerAd
	if req.AdType == "interstitial" {
		coinsToReward = coinsPerAd / 2
	} else if req.AdType == "banner" {
		coinsToReward = 1
	}

	// Registrar validación del anuncio ANTES de dar monedas
	if err := h.adsValidator.RecordAdValidation(ctx, req.AdID, uid.String(), ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record ad validation",
		})
		return
	}

	// Operación atómica: actualizar balance + insertar registro de transacción.
	// Si alguna falla, hacemos rollback para evitar inconsistencias de balance.
	dbTx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] RewardCoinsForAd: BeginTx failed for user %s: %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to begin transaction",
		})
		return
	}
	defer dbTx.Rollback() //nolint:errcheck — el commit exitoso cancela este defer

	// 1. Actualizar balance y obtener nuevo valor (atómico contra race conditions)
	var newBalance int
	err = dbTx.QueryRowContext(ctx,
		`UPDATE users
		 SET coin_balance = coin_balance + $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING coin_balance`,
		coinsToReward, uid,
	).Scan(&newBalance)
	if err != nil {
		log.Printf("[ERROR] RewardCoinsForAd: UpdateCoinBalance failed for user %s: %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update coin balance",
		})
		return
	}

	// 2. Registrar transacción para auditoría / analytics
	_, err = dbTx.ExecContext(ctx,
		`INSERT INTO transactions (user_id, type, amount, episode_id, method)
		 VALUES ($1, 'ad_reward', $2, NULL, 'AD')`,
		uid, coinsToReward,
	)
	if err != nil {
		log.Printf("[ERROR] RewardCoinsForAd: transaction insert failed for user %s: %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record transaction",
		})
		return
	}

	if err := dbTx.Commit(); err != nil {
		log.Printf("[ERROR] RewardCoinsForAd: Commit failed for user %s: %v", uid, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                "Coins rewarded successfully",
		"coins_earned":           coinsToReward,
		"new_balance":            newBalance,
		"daily_limit_remaining":  validation.DailyLimitRemaining,
		"hourly_limit_remaining": validation.HourlyLimitRemaining,
		"ad_id":                  req.AdID,
		"ad_type":                req.AdType,
	})
}
