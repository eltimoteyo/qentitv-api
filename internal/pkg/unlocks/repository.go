package unlocks

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// IsUnlocked verifica si un usuario tiene desbloqueado un episodio
func (r *Repository) IsUnlocked(ctx context.Context, userID, episodeID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM unlocks WHERE user_id = $1 AND episode_id = $2`
	
	err := r.db.QueryRowContext(ctx, query, userID, episodeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check unlock: %w", err)
	}
	
	return count > 0, nil
}

// Create crea un nuevo desbloqueo
func (r *Repository) Create(ctx context.Context, unlock *models.Unlock) error {
	unlock.ID = uuid.New()
	query := `INSERT INTO unlocks (id, user_id, episode_id, method) 
	          VALUES ($1, $2, $3, $4) RETURNING unlocked_at`
	
	err := r.db.QueryRowContext(ctx, query,
		unlock.ID, unlock.UserID, unlock.EpisodeID, unlock.Method,
	).Scan(&unlock.UnlockedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create unlock: %w", err)
	}
	
	return nil
}

// GetUnlockedEpisodes retorna todos los episodios desbloqueados por un usuario
func (r *Repository) GetUnlockedEpisodes(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT episode_id FROM unlocks WHERE user_id = $1`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unlocks: %w", err)
	}
	defer rows.Close()
	
	var episodeIDs []uuid.UUID
	for rows.Next() {
		var episodeID uuid.UUID
		if err := rows.Scan(&episodeID); err != nil {
			return nil, fmt.Errorf("failed to scan episode_id: %w", err)
		}
		episodeIDs = append(episodeIDs, episodeID)
	}
	
	return episodeIDs, nil
}

