package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create crea un nuevo refresh token en la DB
func (r *RefreshTokenRepository) Create(ctx context.Context, tokenID string, userID uuid.UUID, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (id, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, tokenID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

// Validate verifica que un refresh token existe, no está revocado y no ha expirado
func (r *RefreshTokenRepository) Validate(ctx context.Context, tokenID string) (*uuid.UUID, error) {
	var userID uuid.UUID
	var expiresAt time.Time
	var revoked bool
	
	query := `SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, tokenID).Scan(&userID, &expiresAt, &revoked)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}
	
	if revoked {
		return nil, fmt.Errorf("refresh token has been revoked")
	}
	
	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}
	
	return &userID, nil
}

// Revoke revoca un refresh token
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found")
	}
	
	return nil
}

// RevokeAllUserTokens revoca todos los refresh tokens de un usuario
func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}
	return nil
}

// CleanupExpired elimina tokens expirados (para mantenimiento)
func (r *RefreshTokenRepository) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

