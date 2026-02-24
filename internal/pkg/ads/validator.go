package ads

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Validator valida anuncios para prevenir fraude
type Validator struct {
	db *sql.DB
}

func NewValidator(db *sql.DB) *Validator {
	return &Validator{db: db}
}

// AdValidationResult representa el resultado de la validación
type AdValidationResult struct {
	Valid   bool
	Reason  string
	AdID    string
	UserID  string
	EpisodeID string
}

// ValidateAd valida que un anuncio fue realmente visto
func (v *Validator) ValidateAd(ctx context.Context, adID, userID, episodeID string) (*AdValidationResult, error) {
	// 1. Verificar formato del ad_id
	if len(adID) < 10 {
		return &AdValidationResult{
			Valid:  false,
			Reason: "Invalid ad ID format",
			AdID:   adID,
		}, nil
	}
	
	// 2. Verificar que el mismo ad_id no se haya usado recientemente (prevenir reutilización)
	// Guardamos un registro de ad_ids usados en los últimos 5 minutos
	var count int
	query := `SELECT COUNT(*) FROM ad_validations 
	          WHERE ad_id = $1 AND user_id = $2 AND created_at > NOW() - INTERVAL '5 minutes'`
	
	err := v.db.QueryRowContext(ctx, query, adID, userID).Scan(&count)
	if err == nil && count > 0 {
		return &AdValidationResult{
			Valid:  false,
			Reason: "Ad ID already used recently",
			AdID:   adID,
		}, nil
	}
	
	// 3. Registrar el ad_id usado (para prevenir reutilización)
	// Nota: Necesitamos crear la tabla ad_validations si no existe
	// Por ahora, usamos una tabla temporal o la creamos en migraciones
	
	// 4. Validación básica pasada
	// TODO: En producción, aquí se integraría con el SDK de ads real
	// Por ahora validamos formato y reutilización
	
	return &AdValidationResult{
		Valid:     true,
		Reason:    "Valid",
		AdID:      adID,
		UserID:    userID,
		EpisodeID: episodeID,
	}, nil
}

// RecordAdValidation registra que un anuncio fue validado
func (v *Validator) RecordAdValidation(ctx context.Context, adID, userID, episodeID string) error {
	// Insertar registro (la tabla se crea en las migraciones)
	insertQuery := `INSERT INTO ad_validations (ad_id, user_id, episode_id) VALUES ($1, $2, $3)`
	_, err := v.db.ExecContext(ctx, insertQuery, adID, userID, episodeID)
	if err != nil {
		return fmt.Errorf("failed to record ad validation: %w", err)
	}
	
	return nil
}

// CleanupOldValidations limpia validaciones antiguas (más de 1 hora)
func (v *Validator) CleanupOldValidations(ctx context.Context) error {
	query := `DELETE FROM ad_validations WHERE created_at < NOW() - INTERVAL '1 hour'`
	_, err := v.db.ExecContext(ctx, query)
	return err
}


// AdRewardValidationResult representa el resultado de la validación para recompensas
type AdRewardValidationResult struct {
	Valid              bool
	Reason             string
	DailyCount         int
	HourlyCount        int
	DailyLimitRemaining int
	HourlyLimitRemaining int
	CooldownSeconds    int
}

// ValidateAdReward valida que un anuncio puede otorgar recompensa (monedas)
func (v *Validator) ValidateAdReward(ctx context.Context, adID, userID string, cooldownMinutes, dailyLimit, hourlyLimit int) (*AdRewardValidationResult, error) {
	if len(adID) < 10 {
		return &AdRewardValidationResult{
			Valid:  false,
			Reason: "Invalid ad ID format",
		}, nil
	}

	var lastAdTime sql.NullTime
	// Obtenemos el último anuncio visto en las últimas 24h por este usuario.
	// Luego comparamos en Go si el tiempo transcurrido es menor que cooldownMinutes.
	cooldownQuery := `SELECT MAX(created_at) FROM ad_validations
	                  WHERE user_id = $1 AND created_at > NOW() - INTERVAL '24 hours'`
	err := v.db.QueryRowContext(ctx, cooldownQuery, userID).Scan(&lastAdTime)
	if err == nil && lastAdTime.Valid {
		elapsed := time.Since(lastAdTime.Time)
		if int(elapsed.Minutes()) < cooldownMinutes {
			remainingSecs := cooldownMinutes*60 - int(elapsed.Seconds())
			if remainingSecs < 0 {
				remainingSecs = 0
			}
			return &AdRewardValidationResult{
				Valid:           false,
				Reason:          fmt.Sprintf("Cooldown activo. Espera %d segundos antes de ver otro anuncio.", remainingSecs),
				CooldownSeconds: remainingSecs,
			}, nil
		}
	}

	var dailyCount int
	dailyQuery := `SELECT COUNT(*) FROM ad_validations 
	               WHERE user_id = $1 AND created_at >= CURRENT_DATE`
	err = v.db.QueryRowContext(ctx, dailyQuery, userID).Scan(&dailyCount)
	if err != nil {
		dailyCount = 0
	}

	if dailyCount >= dailyLimit {
		return &AdRewardValidationResult{
			Valid:              false,
			Reason:             "Daily ad limit reached",
			DailyCount:         dailyCount,
			DailyLimitRemaining: 0,
		}, nil
	}

	var hourlyCount int
	hourlyQuery := `SELECT COUNT(*) FROM ad_validations 
	                WHERE user_id = $1 AND created_at > NOW() - INTERVAL '1 hour'`
	err = v.db.QueryRowContext(ctx, hourlyQuery, userID).Scan(&hourlyCount)
	if err != nil {
		hourlyCount = 0
	}

	if hourlyCount >= hourlyLimit {
		return &AdRewardValidationResult{
			Valid:               false,
			Reason:              "Hourly ad limit reached",
			HourlyCount:         hourlyCount,
			HourlyLimitRemaining: 0,
		}, nil
	}

	var adIDCount int
	adIDQuery := `SELECT COUNT(*) FROM ad_validations 
	              WHERE ad_id = $1 AND user_id = $2 AND created_at > NOW() - INTERVAL '5 minutes'`
	err = v.db.QueryRowContext(ctx, adIDQuery, adID, userID).Scan(&adIDCount)
	if err == nil && adIDCount > 0 {
		return &AdRewardValidationResult{
			Valid:  false,
			Reason: "This ad has already been used recently",
		}, nil
	}

	return &AdRewardValidationResult{
		Valid:               true,
		Reason:              "Valid",
		DailyCount:          dailyCount,
		HourlyCount:         hourlyCount,
		DailyLimitRemaining: dailyLimit - dailyCount - 1,
		HourlyLimitRemaining: hourlyLimit - hourlyCount - 1,
	}, nil
}
