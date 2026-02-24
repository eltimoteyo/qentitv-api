package app

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
)

// ToggleFavorite agrega o elimina una serie de los favoritos del usuario.
// Si ya era favorita la elimina; si no, la agrega. Devuelve el nuevo estado.
//
// POST /api/v1/app/favorites/:series_id
func (h *Handlers) ToggleFavorite(c *gin.Context) {
	ctx := c.Request.Context()

	seriesIDStr := c.Param("series_id")
	seriesID, err := uuid.Parse(seriesIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid series ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	uid := userID.(uuid.UUID)

	// Intentar insertar (UPSERT tipo "ignorar si ya existe")
	var newID uuid.UUID
	insertErr := h.db.QueryRowContext(ctx,
		`INSERT INTO favorites (user_id, series_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, series_id) DO NOTHING
		 RETURNING id`,
		uid, seriesID,
	).Scan(&newID)

	if insertErr == nil {
		// Se insertó → favorito AÑADIDO
		c.JSON(http.StatusOK, gin.H{"favorited": true, "series_id": seriesID})
		return
	}

	if insertErr == sql.ErrNoRows {
		// Ya existía → ELIMINAR (toggle off)
		if _, err := h.db.ExecContext(ctx,
			`DELETE FROM favorites WHERE user_id = $1 AND series_id = $2`,
			uid, seriesID,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"favorited": false, "series_id": seriesID})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle favorite"})
}

// GetFavorites devuelve las series marcadas como favoritas por el usuario.
//
// GET /api/v1/app/favorites
func (h *Handlers) GetFavorites(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	uid := userID.(uuid.UUID)

	query := `
		SELECT s.id, s.title, s.description, s.horizontal_poster, s.vertical_poster,
		       s.is_active, s.created_at, s.updated_at
		FROM favorites f
		JOIN series s ON s.id = f.series_id
		WHERE f.user_id = $1
		  AND s.is_active = TRUE
		ORDER BY f.created_at DESC
	`
	rows, err := h.db.QueryContext(ctx, query, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch favorites"})
		return
	}
	defer rows.Close()

	seriesList := []models.Series{}
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(
			&s.ID, &s.Title, &s.Description, &s.HorizontalPoster,
			&s.VerticalPoster, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read favorites"})
			return
		}
		seriesList = append(seriesList, s)
	}

	c.JSON(http.StatusOK, gin.H{"series": seriesList})
}
