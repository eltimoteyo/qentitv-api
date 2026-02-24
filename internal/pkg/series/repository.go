package series

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

// GetAll retorna todas las series activas
func (r *Repository) GetAll(ctx context.Context) ([]models.Series, error) {
	query := `SELECT id, title, description, horizontal_poster, vertical_poster, 
	          is_active, created_at, updated_at 
	          FROM series WHERE is_active = TRUE ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query series: %w", err)
	}
	defer rows.Close()
	
	var series []models.Series
	for rows.Next() {
		var s models.Series
		err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}
		series = append(series, s)
	}
	
	return series, nil
}

// GetByID retorna una serie por ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Series, error) {
	var s models.Series
	query := `SELECT id, title, description, horizontal_poster, vertical_poster, 
	          is_active, created_at, updated_at 
	          FROM series WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
		&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("series not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get series: %w", err)
	}
	
	return &s, nil
}

// Create crea una nueva serie
func (r *Repository) Create(ctx context.Context, series *models.Series) error {
	series.ID = uuid.New()
	query := `INSERT INTO series (id, title, description, horizontal_poster, vertical_poster, is_active, producer_id) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		series.ID, series.Title, series.Description,
		series.HorizontalPoster, series.VerticalPoster, series.IsActive, series.ProducerID,
	).Scan(&series.CreatedAt, &series.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create series: %w", err)
	}
	
	return nil
}

// Update actualiza una serie existente
func (r *Repository) Update(ctx context.Context, series *models.Series) error {
	query := `UPDATE series 
	          SET title = $1, description = $2, horizontal_poster = $3, 
	              vertical_poster = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $6 RETURNING updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		series.Title, series.Description, series.HorizontalPoster,
		series.VerticalPoster, series.IsActive, series.ID,
	).Scan(&series.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return fmt.Errorf("series not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update series: %w", err)
	}
	
	return nil
}

// Delete elimina una serie (soft delete marcando is_active = false)
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE series SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete series: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("series not found")
	}
	
	return nil
}

// GetTrending devuelve series ordenadas por score = views*1 + unlocks*2 en los últimos `days` días.
// Garantiza al menos `limit` resultados aunque no haya actividad reciente (rellena con series recientes).
func (r *Repository) GetTrending(ctx context.Context, limit, days int) ([]models.Series, error) {
	query := `
		SELECT s.id, s.title, s.description, s.horizontal_poster, s.vertical_poster,
		       s.is_active, s.created_at, s.updated_at
		FROM series s
		LEFT JOIN (
			SELECT e.series_id, COUNT(v.id) AS view_count
			FROM views v
			JOIN episodes e ON e.id = v.episode_id
			WHERE v.created_at > NOW() - ($2 || ' days')::INTERVAL
			GROUP BY e.series_id
		) vc ON vc.series_id = s.id
		LEFT JOIN (
			SELECT e.series_id, COUNT(u.id) AS unlock_count
			FROM unlocks u
			JOIN episodes e ON e.id = u.episode_id
			WHERE u.unlocked_at > NOW() - ($2 || ' days')::INTERVAL
			GROUP BY e.series_id
		) uc ON uc.series_id = s.id
		WHERE s.is_active = TRUE
		ORDER BY (COALESCE(vc.view_count, 0) + COALESCE(uc.unlock_count, 0) * 2) DESC,
		         s.created_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trending series: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// GetAllAdmin retorna las series para el panel de admin.
// Si producerID != nil filtra por productor; si es nil devuelve todas (super_admin).
func (r *Repository) GetAllAdmin(ctx context.Context, producerID *uuid.UUID) ([]models.Series, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if producerID == nil {
		query := `SELECT id, title, description, horizontal_poster, vertical_poster,
		          is_active, producer_id, created_at, updated_at
		          FROM series ORDER BY created_at DESC`
		rows, err = r.db.QueryContext(ctx, query)
	} else {
		query := `SELECT id, title, description, horizontal_poster, vertical_poster,
		          is_active, producer_id, created_at, updated_at
		          FROM series WHERE producer_id = $1 ORDER BY created_at DESC`
		rows, err = r.db.QueryContext(ctx, query, *producerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query series for admin: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.ProducerID, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// BelongsToProducer verifica que una serie pertenece a un productor dado.
// Devuelve true si la serie tiene ese producer_id, o si producerID está vacío (super_admin, sin restricción).
func (r *Repository) BelongsToProducer(ctx context.Context, seriesID uuid.UUID, producerID string) (bool, error) {
	if producerID == "" {
		return true, nil // super_admin, sin restricción
	}
	pID, err := uuid.Parse(producerID)
	if err != nil {
		return false, fmt.Errorf("invalid producer_id")
	}
	var count int
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM series WHERE id = $1 AND producer_id = $2`,
		seriesID, pID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Search busca series por título o descripción (case-insensitive).
func (r *Repository) Search(ctx context.Context, q string, limit int) ([]models.Series, error) {
	query := `
		SELECT id, title, description, horizontal_poster, vertical_poster,
		       is_active, created_at, updated_at
		FROM series
		WHERE is_active = TRUE
		  AND (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
		ORDER BY
		  CASE WHEN title ILIKE $1 || '%' THEN 0
		       WHEN title ILIKE '%' || $1 || '%' THEN 1
		       ELSE 2 END,
		  title ASC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, q, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search series: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// ── Variantes filtradas por productor (tenant isolation) ──────────────────────

// GetAllFiltered retorna series activas, opcionalmente filtradas por productor.
// Si producerID es nil devuelve todas (comportamiento legacy / super_admin).
func (r *Repository) GetAllFiltered(ctx context.Context, producerID *uuid.UUID) ([]models.Series, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if producerID == nil {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, title, description, horizontal_poster, vertical_poster,
			        is_active, created_at, updated_at
			 FROM series WHERE is_active = TRUE ORDER BY created_at DESC`)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, title, description, horizontal_poster, vertical_poster,
			        is_active, created_at, updated_at
			 FROM series WHERE is_active = TRUE AND producer_id = $1 ORDER BY created_at DESC`,
			*producerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query series: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan series: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// GetTrendingFiltered retorna series trending, opcionalmente filtradas por productor.
func (r *Repository) GetTrendingFiltered(ctx context.Context, limit, days int, producerID *uuid.UUID) ([]models.Series, error) {
	var (
		rows *sql.Rows
		err  error
	)
	base := `
		SELECT s.id, s.title, s.description, s.horizontal_poster, s.vertical_poster,
		       s.is_active, s.created_at, s.updated_at
		FROM series s
		LEFT JOIN (
			SELECT e.series_id, COUNT(v.id) AS view_count
			FROM views v
			JOIN episodes e ON e.id = v.episode_id
			WHERE v.created_at > NOW() - ($2 || ' days')::INTERVAL
			GROUP BY e.series_id
		) vc ON vc.series_id = s.id
		LEFT JOIN (
			SELECT e.series_id, COUNT(u.id) AS unlock_count
			FROM unlocks u
			JOIN episodes e ON e.id = u.episode_id
			WHERE u.unlocked_at > NOW() - ($2 || ' days')::INTERVAL
			GROUP BY e.series_id
		) uc ON uc.series_id = s.id
		WHERE s.is_active = TRUE`

	if producerID == nil {
		rows, err = r.db.QueryContext(ctx, base+`
		ORDER BY (COALESCE(vc.view_count, 0) + COALESCE(uc.unlock_count, 0) * 2) DESC,
		         s.created_at DESC
		LIMIT $1`, limit, days)
	} else {
		rows, err = r.db.QueryContext(ctx, base+`
		  AND s.producer_id = $3
		ORDER BY (COALESCE(vc.view_count, 0) + COALESCE(uc.unlock_count, 0) * 2) DESC,
		         s.created_at DESC
		LIMIT $1`, limit, days, *producerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query trending: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trending series: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// SearchFiltered busca series por texto, opcionalmente filtrado por productor.
func (r *Repository) SearchFiltered(ctx context.Context, q string, limit int, producerID *uuid.UUID) ([]models.Series, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if producerID == nil {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, title, description, horizontal_poster, vertical_poster,
			       is_active, created_at, updated_at
			FROM series
			WHERE is_active = TRUE
			  AND (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
			ORDER BY
			  CASE WHEN title ILIKE $1 || '%' THEN 0
			       WHEN title ILIKE '%' || $1 || '%' THEN 1
			       ELSE 2 END, title ASC
			LIMIT $2`, q, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, title, description, horizontal_poster, vertical_poster,
			       is_active, created_at, updated_at
			FROM series
			WHERE is_active = TRUE
			  AND producer_id = $3
			  AND (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
			ORDER BY
			  CASE WHEN title ILIKE $1 || '%' THEN 0
			       WHEN title ILIKE '%' || $1 || '%' THEN 1
			       ELSE 2 END, title ASC
			LIMIT $2`, q, limit, *producerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search series: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// GetMostViewed retorna series ordenadas por total de vistas histórico (all-time).
func (r *Repository) GetMostViewed(ctx context.Context, limit int) ([]models.Series, error) {
	query := `
		SELECT s.id, s.title, s.description, s.horizontal_poster, s.vertical_poster,
		       s.is_active, s.created_at, s.updated_at
		FROM series s
		LEFT JOIN (
			SELECT e.series_id, COUNT(v.id) AS total_views
			FROM views v
			JOIN episodes e ON e.id = v.episode_id
			GROUP BY e.series_id
		) vc ON vc.series_id = s.id
		WHERE s.is_active = TRUE
		ORDER BY COALESCE(vc.total_views, 0) DESC, s.created_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query most viewed: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan most viewed: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}

// GetNewReleases retorna series publicadas recientemente (últimos `days` días),
// ordenadas por fecha de creación descendente.
func (r *Repository) GetNewReleases(ctx context.Context, limit, days int) ([]models.Series, error) {
	query := `
		SELECT id, title, description, horizontal_poster, vertical_poster,
		       is_active, created_at, updated_at
		FROM series
		WHERE is_active = TRUE
		  AND created_at > NOW() - ($2 || ' days')::INTERVAL
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query new releases: %w", err)
	}
	defer rows.Close()

	var result []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan new releases: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}
