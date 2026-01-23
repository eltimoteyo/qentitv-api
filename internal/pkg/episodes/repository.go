package episodes

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

// GetBySeriesID retorna todos los episodios de una serie ordenados por número
func (r *Repository) GetBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]models.Episode, error) {
	query := `SELECT id, series_id, episode_number, title, video_id_bunny, duration, 
	          is_free, price_coins, created_at, updated_at 
	          FROM episodes WHERE series_id = $1 ORDER BY episode_number ASC`
	
	rows, err := r.db.QueryContext(ctx, query, seriesID)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()
	
	var episodes []models.Episode
	for rows.Next() {
		var e models.Episode
		err := rows.Scan(
			&e.ID, &e.SeriesID, &e.EpisodeNumber, &e.Title,
			&e.VideoIDBunny, &e.Duration, &e.IsFree, &e.PriceCoins,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episodes = append(episodes, e)
	}
	
	return episodes, nil
}

// GetByID retorna un episodio por ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Episode, error) {
	var e models.Episode
	query := `SELECT id, series_id, episode_number, title, video_id_bunny, duration, 
	          is_free, price_coins, created_at, updated_at 
	          FROM episodes WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.SeriesID, &e.EpisodeNumber, &e.Title,
		&e.VideoIDBunny, &e.Duration, &e.IsFree, &e.PriceCoins,
		&e.CreatedAt, &e.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("episode not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get episode: %w", err)
	}
	
	return &e, nil
}

// Create crea un nuevo episodio
func (r *Repository) Create(ctx context.Context, episode *models.Episode) error {
	episode.ID = uuid.New()
	query := `INSERT INTO episodes (id, series_id, episode_number, title, video_id_bunny, 
	          duration, is_free, price_coins) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		episode.ID, episode.SeriesID, episode.EpisodeNumber, episode.Title,
		episode.VideoIDBunny, episode.Duration, episode.IsFree, episode.PriceCoins,
	).Scan(&episode.CreatedAt, &episode.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create episode: %w", err)
	}
	
	return nil
}

// Update actualiza un episodio existente
func (r *Repository) Update(ctx context.Context, episode *models.Episode) error {
	query := `UPDATE episodes 
	          SET title = $1, video_id_bunny = $2, duration = $3, 
	              is_free = $4, price_coins = $5, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $6 RETURNING updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		episode.Title, episode.VideoIDBunny, episode.Duration,
		episode.IsFree, episode.PriceCoins, episode.ID,
	).Scan(&episode.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("episode not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update episode: %w", err)
	}
	
	return nil
}

// UpdateVideoID actualiza el video_id_bunny de un episodio después de la subida
func (r *Repository) UpdateVideoID(ctx context.Context, episodeID uuid.UUID, videoID string) error {
	query := `UPDATE episodes SET video_id_bunny = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, videoID, episodeID)
	if err != nil {
		return fmt.Errorf("failed to update video_id: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("episode not found")
	}
	
	return nil
}

// GetAll retorna todos los episodios con filtros opcionales
func (r *Repository) GetAll(ctx context.Context, seriesID *uuid.UUID) ([]models.Episode, error) {
	var query string
	var args []interface{}
	
	if seriesID != nil {
		query = `SELECT id, series_id, episode_number, title, video_id_bunny, duration, 
		         is_free, price_coins, created_at, updated_at 
		         FROM episodes WHERE series_id = $1 ORDER BY episode_number ASC`
		args = []interface{}{*seriesID}
	} else {
		query = `SELECT id, series_id, episode_number, title, video_id_bunny, duration, 
		         is_free, price_coins, created_at, updated_at 
		         FROM episodes ORDER BY created_at DESC`
		args = []interface{}{}
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	defer rows.Close()
	
	var episodes []models.Episode
	for rows.Next() {
		var e models.Episode
		err := rows.Scan(
			&e.ID, &e.SeriesID, &e.EpisodeNumber, &e.Title,
			&e.VideoIDBunny, &e.Duration, &e.IsFree, &e.PriceCoins,
			&e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episodes = append(episodes, e)
	}
	
	return episodes, nil
}

// Delete elimina un episodio
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM episodes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete episode: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("episode not found")
	}
	
	return nil
}

