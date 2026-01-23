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
	query := `INSERT INTO series (id, title, description, horizontal_poster, vertical_poster, is_active) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		series.ID, series.Title, series.Description,
		series.HorizontalPoster, series.VerticalPoster, series.IsActive,
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

