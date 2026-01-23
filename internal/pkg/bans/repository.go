package bans

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Ban representa un ban de usuario
type Ban struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Reason    string
	BannedBy  *uuid.UUID
	ExpiresAt *time.Time
	IsActive  bool
	CreatedAt time.Time
}

// Create crea un nuevo ban
func (r *Repository) Create(ctx context.Context, ban *Ban) error {
	ban.ID = uuid.New()
	query := `INSERT INTO bans (id, user_id, reason, banned_by, expires_at, is_active) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	
	var bannedByValue interface{}
	if ban.BannedBy != nil {
		bannedByValue = *ban.BannedBy
	} else {
		bannedByValue = nil
	}
	
	var expiresAtValue interface{}
	if ban.ExpiresAt != nil {
		expiresAtValue = *ban.ExpiresAt
	} else {
		expiresAtValue = nil
	}
	
	err := r.db.QueryRowContext(ctx, query,
		ban.ID, ban.UserID, ban.Reason, bannedByValue, expiresAtValue, ban.IsActive,
	).Scan(&ban.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create ban: %w", err)
	}
	
	return nil
}

// IsBanned verifica si un usuario está baneado actualmente
func (r *Repository) IsBanned(ctx context.Context, userID uuid.UUID) (bool, *Ban, error) {
	query := `SELECT id, user_id, reason, banned_by, expires_at, is_active, created_at 
	          FROM bans 
	          WHERE user_id = $1 AND is_active = TRUE 
	          AND (expires_at IS NULL OR expires_at > NOW())
	          ORDER BY created_at DESC 
	          LIMIT 1`
	
	var ban Ban
	var bannedBy sql.NullString
	var expiresAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&ban.ID, &ban.UserID, &ban.Reason, &bannedBy, &expiresAt, &ban.IsActive, &ban.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to check ban: %w", err)
	}
	
	if bannedBy.Valid {
		bannedByUUID, _ := uuid.Parse(bannedBy.String)
		ban.BannedBy = &bannedByUUID
	}
	
	if expiresAt.Valid {
		ban.ExpiresAt = &expiresAt.Time
	}
	
	return true, &ban, nil
}

// Revoke revoca un ban (lo marca como inactivo)
func (r *Repository) Revoke(ctx context.Context, banID uuid.UUID) error {
	query := `UPDATE bans SET is_active = FALSE WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, banID)
	if err != nil {
		return fmt.Errorf("failed to revoke ban: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("ban not found")
	}
	
	return nil
}

// GetUserBans retorna todos los bans de un usuario
func (r *Repository) GetUserBans(ctx context.Context, userID uuid.UUID) ([]Ban, error) {
	query := `SELECT id, user_id, reason, banned_by, expires_at, is_active, created_at 
	          FROM bans 
	          WHERE user_id = $1 
	          ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bans: %w", err)
	}
	defer rows.Close()
	
	var bans []Ban
	for rows.Next() {
		var ban Ban
		var bannedBy sql.NullString
		var expiresAt sql.NullTime
		
		err := rows.Scan(
			&ban.ID, &ban.UserID, &ban.Reason, &bannedBy, &expiresAt, &ban.IsActive, &ban.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		if bannedBy.Valid {
			bannedByUUID, _ := uuid.Parse(bannedBy.String)
			ban.BannedBy = &bannedByUUID
		}
		
		if expiresAt.Valid {
			ban.ExpiresAt = &expiresAt.Time
		}
		
		bans = append(bans, ban)
	}
	
	return bans, nil
}

