package transactions

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

// Transaction representa una transacción
type Transaction struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Amount    int
	EpisodeID *uuid.UUID
	Method    string
	CreatedAt time.Time
}

// Create crea una nueva transacción
func (r *Repository) Create(ctx context.Context, tx *Transaction) error {
	tx.ID = uuid.New()
	query := `INSERT INTO transactions (id, user_id, type, amount, episode_id, method) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at`
	
	var episodeIDValue interface{}
	if tx.EpisodeID != nil {
		episodeIDValue = *tx.EpisodeID
	} else {
		episodeIDValue = nil
	}
	
	err := r.db.QueryRowContext(ctx, query,
		tx.ID, tx.UserID, tx.Type, tx.Amount, episodeIDValue, tx.Method,
	).Scan(&tx.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	
	return nil
}

// GetUserHistory retorna el historial de transacciones de un usuario
func (r *Repository) GetUserHistory(ctx context.Context, userID uuid.UUID, limit int) ([]Transaction, error) {
	query := `SELECT id, user_id, type, amount, episode_id, method, created_at 
	          FROM transactions 
	          WHERE user_id = $1 
	          ORDER BY created_at DESC 
	          LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()
	
	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		var episodeID sql.NullString
		
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &episodeID, &tx.Method, &tx.CreatedAt)
		if err != nil {
			continue
		}
		
		if episodeID.Valid {
			epID, _ := uuid.Parse(episodeID.String)
			tx.EpisodeID = &epID
		}
		
		transactions = append(transactions, tx)
	}
	
	return transactions, nil
}

