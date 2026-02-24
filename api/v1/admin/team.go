package admin

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/jwt"
)

type TeamHandlers struct {
	db *sql.DB
}

func NewTeamHandlers(db *sql.DB) *TeamHandlers {
	return &TeamHandlers{db: db}
}

type TeamMember struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	PhotoURL string    `json:"photo_url"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsOwner  bool      `json:"is_owner"`
}

// GetTeamMembers lista todos los miembros del equipo: dueño + miembros invitados.
func (h *TeamHandlers) GetTeamMembers(c *gin.Context) {
	claims, _ := c.Get("claims")
	jwtClaims := claims.(*jwt.Claims)

	producerID, err := uuid.Parse(jwtClaims.ProducerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No producer associated with this account"})
		return
	}

	ctx := c.Request.Context()

	rows, err := h.db.QueryContext(ctx, `
		SELECT u.id, u.email, COALESCE(u.name, ''), COALESCE(u.photo_url, ''), 'owner' AS role, p.created_at, TRUE AS is_owner
		FROM producers p
		JOIN users u ON u.id = p.user_id
		WHERE p.id = $1

		UNION ALL

		SELECT u.id, u.email, COALESCE(u.name, ''), COALESCE(u.photo_url, ''), pm.role, pm.joined_at, FALSE AS is_owner
		FROM producer_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.producer_id = $1

		ORDER BY is_owner DESC, joined_at ASC
	`, producerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team members"})
		return
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.UserID, &m.Email, &m.Name, &m.PhotoURL, &m.Role, &m.JoinedAt, &m.IsOwner); err == nil {
			members = append(members, m)
		}
	}
	if members == nil {
		members = []TeamMember{}
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// RemoveMember elimina un miembro invitado del equipo del tenant.
// No se puede eliminar al dueño ni a uno mismo.
func (h *TeamHandlers) RemoveMember(c *gin.Context) {
	claims, _ := c.Get("claims")
	jwtClaims := claims.(*jwt.Claims)

	producerID, err := uuid.Parse(jwtClaims.ProducerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No producer associated with this account"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// No puedes eliminarte a ti mismo
	requestingUserID, _ := jwtClaims.GetUserID()
	if requestingUserID == targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove yourself from the team"})
		return
	}

	ctx := c.Request.Context()

	// Protección: no se puede eliminar al dueño de la productora
	var ownerID string
	if err := h.db.QueryRowContext(ctx,
		`SELECT user_id FROM producers WHERE id = $1`, producerID,
	).Scan(&ownerID); err == nil && ownerID == targetUserID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove the producer owner"})
		return
	}

	result, err := h.db.ExecContext(ctx,
		`DELETE FROM producer_members WHERE producer_id = $1 AND user_id = $2`,
		producerID, targetUserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found in this team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}
