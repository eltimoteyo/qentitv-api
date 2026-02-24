package producers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetAll retorna todos los productores con el email de su usuario vinculado y métricas básicas.
func (r *Repository) GetAll(ctx context.Context) ([]models.ProducerWithEmail, error) {
	query := `
		SELECT p.id, p.user_id, p.name, p.slug, p.logo_url, p.description,
		       p.is_active, p.status, p.created_at, p.updated_at, u.email,
		       COALESCE((SELECT COUNT(*) FROM series s WHERE s.producer_id = p.id), 0) AS series_count,
		       COALESCE((SELECT COUNT(*) FROM producer_members pm WHERE pm.producer_id = p.id), 0) AS members_count
		FROM producers p
		JOIN users u ON u.id = p.user_id
		ORDER BY p.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query producers: %w", err)
	}
	defer rows.Close()

	var result []models.ProducerWithEmail
	for rows.Next() {
		var p models.ProducerWithEmail
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.Slug, &p.LogoURL, &p.Description,
			&p.IsActive, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.Email,
			&p.SeriesCount, &p.MembersCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan producer: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}

// GetByID retorna un productor por ID.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.ProducerWithEmail, error) {
	query := `
		SELECT p.id, p.user_id, p.name, p.slug, p.logo_url, p.description,
		       p.is_active, p.status, p.created_at, p.updated_at, u.email
		FROM producers p
		JOIN users u ON u.id = p.user_id
		WHERE p.id = $1`

	var p models.ProducerWithEmail
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.LogoURL, &p.Description,
		&p.IsActive, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.Email,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("producer not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get producer: %w", err)
	}
	return &p, nil
}

// GetByUserID retorna el productor vinculado a un usuario.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Producer, error) {
	query := `SELECT id, user_id, name, slug, logo_url, description, is_active, status, created_at, updated_at
	          FROM producers WHERE user_id = $1 LIMIT 1`
	var p models.Producer
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Slug, &p.LogoURL, &p.Description,
		&p.IsActive, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Create crea un nuevo productor y le asigna el rol 'producer' al usuario.
func (r *Repository) Create(ctx context.Context, p *models.Producer) error {
	p.ID = uuid.New()
	// Generar slug automático si no se proporcionó
	if p.Slug == "" {
		p.Slug = slugify(p.Name) + "-" + p.ID.String()[:8]
	}
	// Status por defecto: pending (requiere aprobación del super_admin)
	if p.Status == "" {
		p.Status = "pending"
	}
	query := `INSERT INTO producers (id, user_id, name, slug, logo_url, description, is_active, status)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		p.ID, p.UserID, p.Name, p.Slug, p.LogoURL, p.Description, p.IsActive, p.Status,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

// SetStatus actualiza el status de un productor (approve/reject/suspend).
// Cuando se activa, también actualiza is_active = true; al suspender lo pone en false.
func (r *Repository) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	isActive := status == "active"
	_, err := r.db.ExecContext(ctx,
		`UPDATE producers SET status = $1, is_active = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`,
		status, isActive, id,
	)
	return err
}

// Update actualiza un productor existente.
func (r *Repository) Update(ctx context.Context, p *models.Producer) error {
	query := `UPDATE producers
	          SET name = $1, logo_url = $2, description = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $5 RETURNING updated_at`
	err := r.db.QueryRowContext(ctx, query,
		p.Name, p.LogoURL, p.Description, p.IsActive, p.ID,
	).Scan(&p.UpdatedAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("producer not found")
	}
	return err
}

// Delete elimina un productor (hard delete — las series quedan huérfanas con producer_id = NULL).
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM producers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("producer not found")
	}
	return nil
}

// slugify convierte un nombre en slug URL-friendly simple.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
