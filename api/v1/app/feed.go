package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
	"github.com/qenti/qenti/internal/pkg/views"
)

// GetFeed retorna el feed del home con algoritmo de recomendados y trending
func (h *Handlers) GetFeed(c *gin.Context) {
	ctx := c.Request.Context()

	producerID, err := resolveProducerSlug(ctx, h.db, c.Query("producer_slug"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve producer"})
		return
	}

	// Obtener todas las series activas (filtradas por tenant si aplica)
	allSeries, err := h.seriesRepo.GetAllFiltered(ctx, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch series",
		})
		return
	}

	type FeedSection struct {
		Title  string         `json:"title"`
		Series []models.Series `json:"series"`
	}

	// Trending: Basado en vistas de últimas 48 horas
	trendingSeries := h.getTrendingSeries(ctx, allSeries, 10)

	// Recomendados: Personalizado si el usuario está autenticado
	var recommendedSeries []models.Series
	userID, exists := c.Get("user_id")
	if exists {
		uid := userID.(uuid.UUID)
		recommendedSeries = h.getRecommendedSeries(ctx, uid, allSeries)
	} else {
		// Si no está autenticado, mostrar series más populares
		recommendedSeries = h.getTrendingSeries(ctx, allSeries, 15)
	}

	feed := []FeedSection{
		{
			Title:  "Trending",
			Series: trendingSeries,
		},
		{
			Title:  "Recomendados para ti",
			Series: recommendedSeries,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"feed": feed,
	})
}

// getTrendingSeries obtiene las series más vistas en las últimas 48 horas
func (h *Handlers) getTrendingSeries(ctx context.Context, allSeries []models.Series, limit int) []models.Series {
	_ = views.NewRepository(h.db) // viewsRepo no usado aún
	
	// Obtener top episodios vistos en últimas 48 horas
	var topEpisodes []uuid.UUID
	topEpisodesQuery := `
		SELECT e.id, COUNT(v.id) as view_count
		FROM episodes e
		JOIN views v ON v.episode_id = e.id
		WHERE v.created_at > NOW() - INTERVAL '48 hours'
		GROUP BY e.id
		ORDER BY view_count DESC
		LIMIT $1
	`
	
	rows, err := h.db.QueryContext(ctx, topEpisodesQuery, limit*5)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var epID uuid.UUID
			var viewCount int
			if err := rows.Scan(&epID, &viewCount); err == nil {
				topEpisodes = append(topEpisodes, epID)
			}
		}
	}
	
	if len(topEpisodes) == 0 {
		// Si no hay vistas recientes, retornar series más recientes
		if len(allSeries) > limit {
			return allSeries[:limit]
		}
		return allSeries
	}
	
	// Mapear episodios a series y contar vistas por serie
	seriesViews := make(map[uuid.UUID]int)
	for _, epID := range topEpisodes {
		// Obtener serie del episodio
		var seriesID uuid.UUID
		query := `SELECT series_id FROM episodes WHERE id = $1`
		err := h.db.QueryRowContext(ctx, query, epID).Scan(&seriesID)
		if err == nil {
			seriesViews[seriesID]++
		}
	}
	
	// Ordenar series por número de vistas
	type SeriesWithViews struct {
		Series models.Series
		Views  int
	}
	
	var seriesList []SeriesWithViews
	for _, s := range allSeries {
		views := seriesViews[s.ID]
		seriesList = append(seriesList, SeriesWithViews{
			Series: s,
			Views:  views,
		})
	}
	
	// Ordenar por vistas (descendente)
	for i := 0; i < len(seriesList)-1; i++ {
		for j := i + 1; j < len(seriesList); j++ {
			if seriesList[i].Views < seriesList[j].Views {
				seriesList[i], seriesList[j] = seriesList[j], seriesList[i]
			}
		}
	}
	
	// Tomar las top N
	result := make([]models.Series, 0, limit)
	for i := 0; i < len(seriesList) && i < limit; i++ {
		result = append(result, seriesList[i].Series)
	}
	
	// Si no hay suficientes con vistas, completar con series recientes
	if len(result) < limit {
		for _, s := range allSeries {
			found := false
			for _, r := range result {
				if r.ID == s.ID {
					found = true
					break
				}
			}
			if !found {
				result = append(result, s)
				if len(result) >= limit {
					break
				}
			}
		}
	}
	
	return result
}

// getRecommendedSeries obtiene recomendaciones personalizadas para un usuario
func (h *Handlers) getRecommendedSeries(ctx context.Context, userID uuid.UUID, allSeries []models.Series) []models.Series {
	// Obtener series que el usuario ya vio
	var watchedSeries []uuid.UUID
	query := `SELECT DISTINCT e.series_id 
	          FROM views v
	          JOIN episodes e ON e.id = v.episode_id
	          WHERE v.user_id = $1 AND v.completed = TRUE`
	
	rows, err := h.db.QueryContext(ctx, query, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var seriesID uuid.UUID
			if err := rows.Scan(&seriesID); err == nil {
				watchedSeries = append(watchedSeries, seriesID)
			}
		}
	}
	
	// Si el usuario no ha visto nada, retornar trending
	if len(watchedSeries) == 0 {
		return h.getTrendingSeries(ctx, allSeries, 15)
	}
	
	// Filtrar series que el usuario NO ha visto
	recommended := make([]models.Series, 0)
	watchedMap := make(map[uuid.UUID]bool)
	for _, id := range watchedSeries {
		watchedMap[id] = true
	}
	
	for _, s := range allSeries {
		if !watchedMap[s.ID] {
			recommended = append(recommended, s)
		}
	}
	
	// Si hay menos de 10 recomendaciones, agregar algunas que ya vio (pero no completó)
	if len(recommended) < 10 {
		// Obtener series que vio pero no completó
		partialQuery := `SELECT DISTINCT e.series_id 
		                 FROM views v
		                 JOIN episodes e ON e.id = v.episode_id
		                 WHERE v.user_id = $1 AND v.completed = FALSE
		                 LIMIT 5`
		
		rows, err := h.db.QueryContext(ctx, partialQuery, userID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var seriesID uuid.UUID
				if err := rows.Scan(&seriesID); err == nil {
					// Buscar la serie y agregarla si no está ya
					for _, s := range allSeries {
						if s.ID == seriesID {
							found := false
							for _, r := range recommended {
								if r.ID == s.ID {
									found = true
									break
								}
							}
							if !found {
								recommended = append(recommended, s)
								break
							}
						}
					}
				}
			}
		}
	}
	
	// Limitar a 15
	if len(recommended) > 15 {
		recommended = recommended[:15]
	}
	
	return recommended
}
