package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterDeviceTokenRequest payload para registrar un token FCM.
type RegisterDeviceTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required"` // "android" | "ios"
}

// RegisterDeviceToken guarda el token FCM del dispositivo del usuario autenticado.
// POST /api/v1/app/device-token
func (h *Handlers) RegisterDeviceToken(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req RegisterDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and platform are required"})
		return
	}

	if req.Platform != "android" && req.Platform != "ios" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform must be 'android' or 'ios'"})
		return
	}

	if err := h.notifService.RegisterToken(ctx, userID, req.Token, req.Platform); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token registered"})
}

// UnregisterDeviceToken elimina el token FCM (logout / permisos revocados).
// DELETE /api/v1/app/device-token
func (h *Handlers) UnregisterDeviceToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	if err := h.notifService.DeleteToken(ctx, req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unregister token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token unregistered"})
}
