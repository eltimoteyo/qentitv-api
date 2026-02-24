package admin

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DashboardHandlers struct {
	db *sql.DB
}

func NewDashboardHandlers(db *sql.DB) *DashboardHandlers {
	return &DashboardHandlers{db: db}
}

// GetDashboard retorna analytics para las gráficas del dashboard.
// Si el rol es "producer" filtra métricas por sus propias series.
func (h *DashboardHandlers) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	// Determinar si hay filtro por producer
	pidStr, _ := c.Get("producer_id")
	producerFilter, _ := pidStr.(string)
	var producerID *uuid.UUID
	if producerFilter != "" {
		id, err := uuid.Parse(producerFilter)
		if err == nil {
			producerID = &id
		}
	}

	var totalSeries, totalEpisodes, totalUsers, activeUsers, premiumUsers int

	if producerID != nil {
		// ── Métricas de productor ──────────────────────────────────────────
		h.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM series WHERE is_active = TRUE AND producer_id = $1`, *producerID,
		).Scan(&totalSeries)

		h.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM episodes e JOIN series s ON s.id = e.series_id WHERE s.producer_id = $1`, *producerID,
		).Scan(&totalEpisodes)

		// Para producer: usuarios únicos que vieron su contenido (últimos 7d)
		h.db.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT v.user_id) FROM views v
			 JOIN episodes e ON e.id = v.episode_id
			 JOIN series s ON s.id = e.series_id
			 WHERE s.producer_id = $1 AND v.created_at > NOW() - INTERVAL '7 days' AND v.user_id IS NOT NULL`, *producerID,
		).Scan(&activeUsers)

		// Total de usuarios registrados (global, útil como contexto)
		h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers)
		h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_premium = TRUE`).Scan(&premiumUsers)
	} else {
		// ── Métricas globales (super_admin) ───────────────────────────────
		h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM series WHERE is_active = TRUE`).Scan(&totalSeries)
		h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM episodes`).Scan(&totalEpisodes)
		h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers)
		h.db.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT user_id) FROM views WHERE created_at > NOW() - INTERVAL '7 days' AND user_id IS NOT NULL`,
		).Scan(&activeUsers)
		h.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_premium = TRUE`).Scan(&premiumUsers)
	}

	// ── Top dramas ────────────────────────────────────────────────────────────
	type TopDrama struct {
		SeriesID uuid.UUID `json:"series_id"`
		Title    string    `json:"title"`
		Views    int       `json:"views"`
	}

	var topDramasRows *sql.Rows
	var tdErr error
	if producerID != nil {
		topDramasRows, tdErr = h.db.QueryContext(ctx,
			`SELECT s.id, s.title, COUNT(v.id) as views
			 FROM series s
			 JOIN episodes e ON e.series_id = s.id
			 LEFT JOIN views v ON v.episode_id = e.id AND v.created_at > NOW() - INTERVAL '30 days'
			 WHERE s.is_active = TRUE AND s.producer_id = $1
			 GROUP BY s.id, s.title
			 ORDER BY views DESC LIMIT 10`, *producerID)
	} else {
		topDramasRows, tdErr = h.db.QueryContext(ctx,
			`SELECT s.id, s.title, COUNT(v.id) as views
			 FROM series s
			 JOIN episodes e ON e.series_id = s.id
			 LEFT JOIN views v ON v.episode_id = e.id AND v.created_at > NOW() - INTERVAL '30 days'
			 WHERE s.is_active = TRUE
			 GROUP BY s.id, s.title
			 ORDER BY views DESC LIMIT 10`)
	}
	var topDramas []TopDrama
	if tdErr == nil {
		defer topDramasRows.Close()
		for topDramasRows.Next() {
			var td TopDrama
			if topDramasRows.Scan(&td.SeriesID, &td.Title, &td.Views) == nil {
				topDramas = append(topDramas, td)
			}
		}
	}

	// ── Retención por episodio ────────────────────────────────────────────────
	type RetentionData struct {
		EpisodeNumber  int     `json:"episode_number"`
		CompletionRate float64 `json:"completion_rate"`
	}

	var retRows *sql.Rows
	var retErr error
	if producerID != nil {
		retRows, retErr = h.db.QueryContext(ctx,
			`SELECT e.episode_number,
			        COALESCE(COUNT(CASE WHEN v.completed = TRUE THEN 1 END)::float / NULLIF(COUNT(v.id), 0), 0)
			 FROM episodes e
			 JOIN series s ON s.id = e.series_id
			 LEFT JOIN views v ON v.episode_id = e.id
			 WHERE s.producer_id = $1
			 GROUP BY e.series_id, e.episode_number
			 ORDER BY e.series_id, e.episode_number LIMIT 20`, *producerID)
	} else {
		retRows, retErr = h.db.QueryContext(ctx,
			`SELECT e.episode_number,
			        COALESCE(COUNT(CASE WHEN v.completed = TRUE THEN 1 END)::float / NULLIF(COUNT(v.id), 0), 0)
			 FROM episodes e
			 LEFT JOIN views v ON v.episode_id = e.id
			 GROUP BY e.series_id, e.episode_number
			 ORDER BY e.series_id, e.episode_number LIMIT 20`)
	}
	var retentionData []RetentionData
	if retErr == nil {
		defer retRows.Close()
		for retRows.Next() {
			var rd RetentionData
			if retRows.Scan(&rd.EpisodeNumber, &rd.CompletionRate) == nil {
				retentionData = append(retentionData, rd)
			}
		}
	}

	// Usuarios activos 30d (global o por producer)
	var activeUsers30d int
	if producerID != nil {
		h.db.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT v.user_id) FROM views v
			 JOIN episodes e ON e.id = v.episode_id
			 JOIN series s ON s.id = e.series_id
			 WHERE s.producer_id = $1 AND v.created_at > NOW() - INTERVAL '30 days' AND v.user_id IS NOT NULL`, *producerID,
		).Scan(&activeUsers30d)
	} else {
		h.db.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT user_id) FROM views WHERE created_at > NOW() - INTERVAL '30 days' AND user_id IS NOT NULL`,
		).Scan(&activeUsers30d)
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": gin.H{
			"total_series":      totalSeries,
			"total_episodes":    totalEpisodes,
			"total_users":       totalUsers,
			"active_users_7d":   activeUsers,
			"active_users_30d":  activeUsers30d,
			"premium_users":     premiumUsers,
			"total_revenue_30d": 0,
		},
		"charts": gin.H{
			"retention_by_episode": retentionData,
			"top_dramas":           topDramas,
			"revenue":              []interface{}{},
		},
	})
}

// GetProducerStatus devuelve el estado actual del tenant del usuario autenticado.
// Usado por el frontend para detectar cuando un productor pendiente es aprobado.
func (h *DashboardHandlers) GetProducerStatus(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	var status string
	err := h.db.QueryRowContext(ctx,
		`SELECT status FROM producers WHERE user_id = $1 LIMIT 1`, userID,
	).Scan(&status)
	if err != nil {
		// Super_admin u otros roles sin producer → devolvemos "active" para no bloquear
		c.JSON(http.StatusOK, gin.H{"status": "active"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

