package invitations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GenerateToken genera un token criptográficamente seguro de 32 bytes (64 hex chars).
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Create crea una nueva invitación y devuelve el registro completo.
func (r *Repository) Create(ctx context.Context, producerID uuid.UUID, role string, createdBy *uuid.UUID, expiresIn time.Duration) (*models.Invitation, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	inv := &models.Invitation{
		Token:      token,
		ProducerID: producerID,
		Role:       role,
		CreatedBy:  createdBy,
		ExpiresAt:  time.Now().Add(expiresIn),
	}

	query := `
		INSERT INTO invitations (token, producer_id, role, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err = r.db.QueryRowContext(ctx, query,
		inv.Token, inv.ProducerID, inv.Role, createdBy, inv.ExpiresAt,
	).Scan(&inv.ID, &inv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}
	return inv, nil
}

// GetByToken busca una invitación válida (no usada y no expirada) por su token.
func (r *Repository) GetByToken(ctx context.Context, token string) (*models.Invitation, error) {
	query := `
		SELECT id, token, producer_id, role, created_by, expires_at, used_at, used_by, created_at
		FROM invitations
		WHERE token = $1`

	inv := &models.Invitation{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&inv.ID, &inv.Token, &inv.ProducerID, &inv.Role,
		&inv.CreatedBy, &inv.ExpiresAt, &inv.UsedAt, &inv.UsedBy, &inv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	return inv, nil
}

// MarkUsed marca una invitación como usada.
func (r *Repository) MarkUsed(ctx context.Context, token string, usedBy uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE invitations SET used_at = CURRENT_TIMESTAMP, used_by = $1 WHERE token = $2`,
		usedBy, token,
	)
	return err
}

// ListByProducer retorna todas las invitaciones de un productor (para mostrar en su panel).
func (r *Repository) ListByProducer(ctx context.Context, producerID uuid.UUID) ([]models.Invitation, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, token, producer_id, role, created_by, expires_at, used_at, used_by, created_at
		 FROM invitations WHERE producer_id = $1 ORDER BY created_at DESC`,
		producerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Invitation
	for rows.Next() {
		var inv models.Invitation
		if err := rows.Scan(
			&inv.ID, &inv.Token, &inv.ProducerID, &inv.Role,
			&inv.CreatedBy, &inv.ExpiresAt, &inv.UsedAt, &inv.UsedBy, &inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, inv)
	}
	return result, nil
}
