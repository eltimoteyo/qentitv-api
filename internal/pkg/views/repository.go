package views

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ContinueWatchingItem representa un episodio en curso que el usuario no terminó.
type ContinueWatchingItem struct {
	SeriesID       uuid.UUID `json:"series_id"`
	SeriesTitle    string    `json:"series_title"`
	VerticalPoster string    `json:"vertical_poster"`
	EpisodeID      uuid.UUID `json:"episode_id"`
	EpisodeNumber  int       `json:"episode_number"`
	EpisodeTitle   string    `json:"episode_title"`
	Duration       int       `json:"duration"`
	WatchedSeconds int       `json:"watched_seconds"`
	Completed      bool      `json:"completed"`
	LastWatched    time.Time `json:"last_watched"`
}

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

// UpdateWatchProgress hace un UPSERT del progreso de visualización del usuario.
// Requiere que exista el índice único parcial idx_views_user_episode (ver migraciones).
func (r *Repository) UpdateWatchProgress(ctx context.Context, userID, episodeID uuid.UUID, watchedSeconds int, completed bool) error {
	query := `
		INSERT INTO views (user_id, episode_id, watched_seconds, completed, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, episode_id) WHERE user_id IS NOT NULL
		DO UPDATE SET
			watched_seconds = EXCLUDED.watched_seconds,
			completed       = EXCLUDED.completed,
			updated_at      = NOW()
	`
	if _, err := r.db.ExecContext(ctx, query, userID, episodeID, watchedSeconds, completed); err != nil {
		return fmt.Errorf("failed to update watch progress: %w", err)
	}
	return nil
}

// GetContinueWatching devuelve las series en curso del usuario (mayor episodio visto, no completado).
// Devuelve máximo `limit` items ordenados por última actividad DESC.
func (r *Repository) GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]ContinueWatchingItem, error) {
	query := `
		SELECT series_id, series_title, vertical_poster,
		       episode_id, episode_number, episode_title, duration,
		       watched_seconds, completed, last_watched
		FROM (
			SELECT DISTINCT ON (e.series_id)
				s.id             AS series_id,
				s.title          AS series_title,
				s.vertical_poster,
				e.id             AS episode_id,
				e.episode_number,
				e.title          AS episode_title,
				e.duration,
				v.watched_seconds,
				v.completed,
				v.updated_at     AS last_watched
			FROM views v
			JOIN episodes e ON e.id = v.episode_id
			JOIN series   s ON s.id = e.series_id
			WHERE v.user_id  = $1
			  AND v.watched_seconds > 0
			  AND v.completed = FALSE
			  AND s.is_active = TRUE
			ORDER BY e.series_id, v.updated_at DESC
		) latest
		ORDER BY last_watched DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query continue watching: %w", err)
	}
	defer rows.Close()

	var result []ContinueWatchingItem
	for rows.Next() {
		var item ContinueWatchingItem
		if err := rows.Scan(
			&item.SeriesID, &item.SeriesTitle, &item.VerticalPoster,
			&item.EpisodeID, &item.EpisodeNumber, &item.EpisodeTitle,
			&item.Duration, &item.WatchedSeconds, &item.Completed, &item.LastWatched,
		); err != nil {
			return nil, fmt.Errorf("failed to scan continue watching: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

