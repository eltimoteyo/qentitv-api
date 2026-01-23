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

// GetDashboard retorna analytics para las gráficas del dashboard
func (h *DashboardHandlers) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Obtener métricas básicas
	var totalSeries, totalEpisodes, totalUsers, activeUsers, premiumUsers int
	
	err := h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM series WHERE is_active = TRUE").Scan(&totalSeries)
	if err != nil {
		totalSeries = 0
	}
	
	err = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM episodes").Scan(&totalEpisodes)
	if err != nil {
		totalEpisodes = 0
	}
	
	err = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		totalUsers = 0
	}
	
	// Usuarios activos (últimos 7 días)
	err = h.db.QueryRowContext(ctx, 
		"SELECT COUNT(DISTINCT user_id) FROM views WHERE created_at > NOW() - INTERVAL '7 days' AND user_id IS NOT NULL",
	).Scan(&activeUsers)
	if err != nil {
		activeUsers = 0
	}
	
	// Usuarios premium
	err = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_premium = TRUE").Scan(&premiumUsers)
	if err != nil {
		premiumUsers = 0
	}
	
	// Top dramas por reproducciones (últimos 30 días)
	type TopDrama struct {
		SeriesID uuid.UUID `json:"series_id"`
		Title    string    `json:"title"`
		Views    int       `json:"views"`
	}
	
	topDramasQuery := `
		SELECT s.id, s.title, COUNT(v.id) as views
		FROM series s
		JOIN episodes e ON e.series_id = s.id
		LEFT JOIN views v ON v.episode_id = e.id AND v.created_at > NOW() - INTERVAL '30 days'
		WHERE s.is_active = TRUE
		GROUP BY s.id, s.title
		ORDER BY views DESC
		LIMIT 10
	`
	
	rows, err := h.db.QueryContext(ctx, topDramasQuery)
	var topDramas []TopDrama
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var td TopDrama
			if err := rows.Scan(&td.SeriesID, &td.Title, &td.Views); err == nil {
				topDramas = append(topDramas, td)
			}
		}
	}
	
	// Retención por episodio (tasa de completación)
	type RetentionData struct {
		EpisodeNumber int     `json:"episode_number"`
		CompletionRate float64 `json:"completion_rate"`
	}
	
	retentionQuery := `
		SELECT e.episode_number, 
		       COALESCE(COUNT(CASE WHEN v.completed = TRUE THEN 1 END)::float / NULLIF(COUNT(v.id), 0), 0) as completion_rate
		FROM episodes e
		LEFT JOIN views v ON v.episode_id = e.id
		GROUP BY e.series_id, e.episode_number
		ORDER BY e.series_id, e.episode_number
		LIMIT 20
	`
	
	rows2, err := h.db.QueryContext(ctx, retentionQuery)
	var retentionData []RetentionData
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var rd RetentionData
			if err := rows2.Scan(&rd.EpisodeNumber, &rd.CompletionRate); err == nil {
				retentionData = append(retentionData, rd)
			}
		}
	}
	
	// Usuarios activos últimos 30 días
	var activeUsers30d int
	err = h.db.QueryRowContext(ctx, 
		"SELECT COUNT(DISTINCT user_id) FROM views WHERE created_at > NOW() - INTERVAL '30 days' AND user_id IS NOT NULL",
	).Scan(&activeUsers30d)
	if err != nil {
		activeUsers30d = 0
	}
	
	// Revenue total últimos 30 días (placeholder)
	totalRevenue30d := 0
	
	c.JSON(http.StatusOK, gin.H{
		"metrics": gin.H{
			"total_series":      totalSeries,
			"total_episodes":    totalEpisodes,
			"total_users":       totalUsers,
			"active_users_7d":   activeUsers,
			"active_users_30d":  activeUsers30d,
			"premium_users":     premiumUsers,
			"total_revenue_30d": totalRevenue30d,
		},
		"charts": gin.H{
			"retention_by_episode": retentionData,
			"top_dramas":           topDramas,
			"revenue":              []interface{}{}, // TODO: Implementar cuando haya datos de RevenueCat
		},
	})
}

