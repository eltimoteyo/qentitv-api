package views

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// RecordView registra una visualización de un episodio
func (r *Repository) RecordView(ctx context.Context, userID *uuid.UUID, episodeID uuid.UUID, watchedSeconds int, completed bool) error {
	query := `INSERT INTO views (user_id, episode_id, watched_seconds, completed) 
	          VALUES ($1, $2, $3, $4)`
	
	var userIDValue interface{}
	if userID != nil {
		userIDValue = *userID
	} else {
		userIDValue = nil
	}
	
	_, err := r.db.ExecContext(ctx, query, userIDValue, episodeID, watchedSeconds, completed)
	if err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}
	
	return nil
}

// GetEpisodeViews retorna el número de vistas de un episodio
func (r *Repository) GetEpisodeViews(ctx context.Context, episodeID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM views WHERE episode_id = $1`
	err := r.db.QueryRowContext(ctx, query, episodeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get episode views: %w", err)
	}
	return count, nil
}

// GetTopEpisodes retorna los episodios más vistos
func (r *Repository) GetTopEpisodes(ctx context.Context, limit int) ([]uuid.UUID, error) {
	query := `SELECT episode_id, COUNT(*) as view_count 
	          FROM views 
	          GROUP BY episode_id 
	          ORDER BY view_count DESC 
	          LIMIT $1`
	
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top episodes: %w", err)
	}
	defer rows.Close()
	
	var episodeIDs []uuid.UUID
	for rows.Next() {
		var episodeID uuid.UUID
		var viewCount int
		if err := rows.Scan(&episodeID, &viewCount); err != nil {
			continue
		}
		episodeIDs = append(episodeIDs, episodeID)
	}
	
	return episodeIDs, nil
}

