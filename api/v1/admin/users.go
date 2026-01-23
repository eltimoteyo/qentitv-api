package admin

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/bans"
	"github.com/qenti/qenti/internal/pkg/transactions"
	"github.com/qenti/qenti/internal/pkg/unlocks"
	"github.com/qenti/qenti/internal/pkg/users"
	"github.com/qenti/qenti/internal/pkg/views"
)

type UsersHandlers struct {
	usersRepo        *users.Repository
	bansRepo         *bans.Repository
	transactionsRepo *transactions.Repository
	unlocksRepo      *unlocks.Repository
	viewsRepo        *views.Repository
	db               *sql.DB
}

func NewUsersHandlers(usersRepo *users.Repository, db *sql.DB) *UsersHandlers {
	return &UsersHandlers{
		usersRepo:        usersRepo,
		bansRepo:         bans.NewRepository(db),
		transactionsRepo: transactions.NewRepository(db),
		unlocksRepo:      unlocks.NewRepository(db),
		viewsRepo:        views.NewRepository(db),
		db:               db,
	}
}

// GetUsers lista usuarios con paginación
func (h *UsersHandlers) GetUsers(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Parámetros de paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	// Contar total
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := h.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count users",
		})
		return
	}
	
	// Obtener usuarios
	query := `SELECT id, email, firebase_uid, coin_balance, is_premium, created_at, updated_at 
	          FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	
	rows, err := h.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}
	defer rows.Close()
	
	type UserListItem struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		CoinBalance int    `json:"coin_balance"`
		IsPremium   bool   `json:"is_premium"`
		CreatedAt   string `json:"created_at"`
	}
	
	var usersList []UserListItem
	for rows.Next() {
		var u UserListItem
		var id uuid.UUID
		var createdAt, updatedAt string
		var firebaseUID string
		
		err := rows.Scan(&id, &u.Email, &firebaseUID, &u.CoinBalance, &u.IsPremium, &createdAt, &updatedAt)
		if err != nil {
			continue
		}
		
		u.ID = id.String()
		u.CreatedAt = createdAt
		usersList = append(usersList, u)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"users": usersList,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + limit - 1) / limit,
		},
	})
}

// GetUserByID obtiene el detalle de un usuario con historial
func (h *UsersHandlers) GetUserByID(c *gin.Context) {
	ctx := c.Request.Context()
	userIDStr := c.Param("id")
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	user, err := h.usersRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	
	// Obtener historial de visionado
	unlockedEpisodes, err := h.unlocksRepo.GetUnlockedEpisodes(ctx, userID)
	if err != nil {
		unlockedEpisodes = []uuid.UUID{}
	}
	
	// Obtener tiempo total de visionado
	var totalWatchTime int
	watchTimeQuery := `SELECT COALESCE(SUM(watched_seconds), 0) FROM views WHERE user_id = $1`
	err = h.db.QueryRowContext(ctx, watchTimeQuery, userID).Scan(&totalWatchTime)
	if err != nil {
		totalWatchTime = 0
	}
	
	// Obtener número de episodios completados
	var completedEpisodes int
	completedQuery := `SELECT COUNT(DISTINCT episode_id) FROM views WHERE user_id = $1 AND completed = TRUE`
	err = h.db.QueryRowContext(ctx, completedQuery, userID).Scan(&completedEpisodes)
	if err != nil {
		completedEpisodes = 0
	}
	
	// Obtener historial de transacciones recientes
	txHistory, err := h.transactionsRepo.GetUserHistory(ctx, userID, 10)
	if err != nil {
		txHistory = []transactions.Transaction{}
	}
	
	// Obtener bans del usuario
	userBans, err := h.bansRepo.GetUserBans(ctx, userID)
	if err != nil {
		userBans = []bans.Ban{}
	}
	
	isBanned, activeBan, _ := h.bansRepo.IsBanned(ctx, userID)
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"history": gin.H{
			"unlocked_episodes": len(unlockedEpisodes),
			"completed_episodes": completedEpisodes,
			"total_watch_time":   totalWatchTime, // en segundos
			"recent_transactions": txHistory,
		},
		"bans": gin.H{
			"is_banned":  isBanned,
			"active_ban": activeBan,
			"all_bans":   userBans,
		},
	})
}

// BanUserRequest representa el payload para banear un usuario
type BanUserRequest struct {
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"` // Opcional, si es nil es ban permanente
}

// BanUser banea un usuario
func (h *UsersHandlers) BanUser(c *gin.Context) {
	ctx := c.Request.Context()
	userIDStr := c.Param("id")
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	// Verificar que el usuario existe
	_, err = h.usersRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	
	// Verificar si ya está baneado
	isBanned, _, _ := h.bansRepo.IsBanned(ctx, userID)
	if isBanned {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User is already banned",
		})
		return
	}
	
	var req BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Obtener ID del admin que está haciendo el ban
	adminID, exists := c.Get("user_id")
	var bannedBy *uuid.UUID
	if exists {
		uid := adminID.(uuid.UUID)
		bannedBy = &uid
	}
	
	// Crear ban
	ban := &bans.Ban{
		UserID:    userID,
		Reason:    req.Reason,
		BannedBy:  bannedBy,
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
	}
	
	if err := h.bansRepo.Create(ctx, ban); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to ban user",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "User banned successfully",
		"user_id": userID,
		"ban_id":  ban.ID,
		"expires_at": ban.ExpiresAt,
	})
}

// GiftCoins regala monedas a un usuario manualmente
type GiftCoinsRequest struct {
	Amount int `json:"amount" binding:"required,min=1"`
}

func (h *UsersHandlers) GiftCoins(c *gin.Context) {
	ctx := c.Request.Context()
	userIDStr := c.Param("id")
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}
	
	var req GiftCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}
	
	// Obtener usuario actual
	user, err := h.usersRepo.GetByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	
	// Actualizar balance
	newBalance := user.CoinBalance + req.Amount
	if err := h.usersRepo.UpdateCoinBalance(ctx, userID, newBalance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update coin balance",
		})
		return
	}
	
	// Registrar transacción
	adminID, _ := c.Get("user_id")
	var adminUUID uuid.UUID
	if adminID != nil {
		adminUUID = adminID.(uuid.UUID)
	}
	
	tx := &transactions.Transaction{
		UserID:    userID,
		Type:      "gift",
		Amount:    req.Amount,
		EpisodeID: nil, // No está asociado a un episodio
		Method:    "GIFT",
	}
	if err := h.transactionsRepo.Create(ctx, tx); err != nil {
		// Log error pero no fallar la operación
		log.Printf("Failed to record transaction: %v", err)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Coins gifted successfully",
		"user_id": userID,
		"amount": req.Amount,
		"new_balance": newBalance,
		"gifted_by": adminUUID,
	})
}

