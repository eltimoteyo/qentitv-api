package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeviceToken representa un token FCM de un dispositivo móvil.
type DeviceToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	Platform  string    `db:"platform"` // "android" | "ios"
	CreatedAt time.Time `db:"created_at"`
}

// Repository gestiona la tabla device_tokens.
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveToken inserta o actualiza el token FCM de un usuario para un dispositivo.
// Usa UPSERT para evitar duplicados (ON CONFLICT en token).
func (r *Repository) SaveToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	query := `
		INSERT INTO device_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE
			SET user_id = EXCLUDED.user_id,
			    platform = EXCLUDED.platform,
			    created_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, userID, token, platform)
	if err != nil {
		return fmt.Errorf("failed to save device token: %w", err)
	}
	return nil
}

// DeleteToken elimina un token FCM específico (al hacer logout).
func (r *Repository) DeleteToken(ctx context.Context, token string) error {
	query := `DELETE FROM device_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete device token: %w", err)
	}
	return nil
}

// GetTokensForFavoriters devuelve todos los tokens FCM de usuarios que tienen
// la serie indicada en favoritos.
func (r *Repository) GetTokensForFavoriters(ctx context.Context, seriesID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT dt.token
		FROM device_tokens dt
		INNER JOIN favorites f ON f.user_id = dt.user_id
		WHERE f.series_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens for favoriters: %w", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}
// GetTokensForUser devuelve todos los tokens FCM registrados por un usuario específico.
func (r *Repository) GetTokensForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT token FROM device_tokens WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens for user: %w", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}