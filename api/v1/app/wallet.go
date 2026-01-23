package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/transactions"
)

// GetWallet obtiene el balance de monedas del usuario
func (h *Handlers) GetWallet(c *gin.Context) {
	ctx := c.Request.Context()
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}
	uid := userID.(uuid.UUID)
	
	user, err := h.usersRepo.GetByID(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get wallet",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"coin_balance": user.CoinBalance,
		"is_premium": user.IsPremium,
	})
}

// GetWalletHistory obtiene el historial de transacciones del usuario
func (h *Handlers) GetWalletHistory(c *gin.Context) {
	ctx := c.Request.Context()
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}
	uid := userID.(uuid.UUID)
	
	// Obtener historial de transacciones
	transactionsRepo := transactions.NewRepository(h.db)
	txHistory, err := transactionsRepo.GetUserHistory(ctx, uid, 50) // Últimas 50 transacciones
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get wallet history",
		})
		return
	}
	
	type TransactionResponse struct {
		ID        string    `json:"id"`
		Type      string    `json:"type"`
		Method    string    `json:"method"`
		Amount    int       `json:"amount"`
		EpisodeID *string   `json:"episode_id,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}
	
	history := make([]TransactionResponse, 0, len(txHistory))
	for _, tx := range txHistory {
		tr := TransactionResponse{
			ID:        tx.ID.String(),
			Type:      tx.Type,
			Method:    tx.Method,
			Amount:    tx.Amount,
			CreatedAt: tx.CreatedAt,
		}
		if tx.EpisodeID != nil {
			epIDStr := tx.EpisodeID.String()
			tr.EpisodeID = &epIDStr
		}
		history = append(history, tr)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"total_transactions": len(history),
	})
}

