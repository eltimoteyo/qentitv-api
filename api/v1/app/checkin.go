package app

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/transactions"
)

// DailyCheckIn registra el check-in diario del usuario y le otorga monedas.
// Si ya reclamó hoy, devuelve 409 con already_claimed=true.
//
// POST /api/v1/app/checkin
func (h *Handlers) DailyCheckIn(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	uid := userID.(uuid.UUID)

	// Inicio del día UTC
	todayStart := time.Now().UTC().Truncate(24 * time.Hour)

	// Verificar si ya hizo check-in hoy
	var existingCount int
	err := h.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions
		 WHERE user_id = $1 AND method = 'DAILY_CHECKIN' AND created_at >= $2`,
		uid, todayStart,
	).Scan(&existingCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check status"})
		return
	}
	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":          "Already claimed today",
			"already_claimed": true,
		})
		return
	}

	// Calcular racha (días consecutivos)
	streak := 1
	yesterday := todayStart.Add(-24 * time.Hour)
	var lastCheckinTime sql.NullTime
	_ = h.db.QueryRowContext(ctx,
		`SELECT MAX(created_at) FROM transactions
		 WHERE user_id = $1 AND method = 'DAILY_CHECKIN'`,
		uid,
	).Scan(&lastCheckinTime)

	if lastCheckinTime.Valid && lastCheckinTime.Time.After(yesterday) {
		// Hay check-in de ayer → contar racha actual
		var totalDays int
		_ = h.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM (
			   SELECT date_trunc('day', created_at) AS day
			   FROM transactions
			   WHERE user_id = $1 AND method = 'DAILY_CHECKIN'
			   GROUP BY day
			 ) t`,
			uid,
		).Scan(&totalDays)
		streak = totalDays + 1 // +1 por hoy
	}

	// Tabla de monedas por día de la racha (ciclo semanal)
	coinsTable := []int{10, 15, 20, 25, 30, 40, 100}
	dayIndex := (streak - 1) % 7
	coinsEarned := coinsTable[dayIndex]

	// Obtener balance actual
	user, err := h.usersRepo.GetByID(ctx, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	newBalance := user.CoinBalance + coinsEarned

	// Actualizar balance
	if err := h.usersRepo.UpdateCoinBalance(ctx, uid, newBalance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	// Registrar transacción
	txRepo := transactions.NewRepository(h.db)
	tx := &transactions.Transaction{
		UserID: uid,
		Type:   "EARN",
		Amount: coinsEarned,
		Method: "DAILY_CHECKIN",
	}
	_ = txRepo.Create(ctx, tx)

	c.JSON(http.StatusOK, gin.H{
		"coins_earned": coinsEarned,
		"new_balance":  newBalance,
		"streak":       streak,
		"day_index":    dayIndex,
	})
}
