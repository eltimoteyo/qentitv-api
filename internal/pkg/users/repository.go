package users

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

// GetByID retorna un usuario por ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	query := `SELECT id, email, firebase_uid, coin_balance, is_premium, created_at, updated_at 
	          FROM users WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.FirebaseUID, &u.CoinBalance,
		&u.IsPremium, &u.CreatedAt, &u.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &u, nil
}

// GetByFirebaseUID retorna un usuario por Firebase UID
func (r *Repository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	var u models.User
	query := `SELECT id, email, firebase_uid, coin_balance, is_premium, created_at, updated_at 
	          FROM users WHERE firebase_uid = $1`
	
	err := r.db.QueryRowContext(ctx, query, firebaseUID).Scan(
		&u.ID, &u.Email, &u.FirebaseUID, &u.CoinBalance,
		&u.IsPremium, &u.CreatedAt, &u.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &u, nil
}

// UpdateCoinBalance actualiza el balance de monedas de un usuario
func (r *Repository) UpdateCoinBalance(ctx context.Context, userID uuid.UUID, newBalance int) error {
	query := `UPDATE users SET coin_balance = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, newBalance, userID)
	if err != nil {
		return fmt.Errorf("failed to update coin balance: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// UpdatePremiumStatus actualiza el estado premium de un usuario
func (r *Repository) UpdatePremiumStatus(ctx context.Context, userID uuid.UUID, isPremium bool) error {
	query := `UPDATE users SET is_premium = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, isPremium, userID)
	if err != nil {
		return fmt.Errorf("failed to update premium status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

