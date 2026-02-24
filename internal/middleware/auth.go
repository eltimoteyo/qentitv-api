package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qenti/qenti/internal/pkg/auth"
	"github.com/qenti/qenti/internal/pkg/jwt"
)

// RequireAuth verifica que el usuario esté autenticado con JWT
func RequireAuth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extraer token (Bearer <token>)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extraer userID desde el campo 'sub'
		userID, err := claims.GetUserID()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Verificar si el usuario está baneado (si tenemos acceso a DB)
		// Nota: Esto requiere pasar el bansRepo, por ahora lo omitimos para no cambiar toda la estructura
		// Se puede agregar después si es necesario

		// Guardar información del usuario en el contexto
		c.Set("claims", claims)
		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("jti", claims.JTI)

		c.Next()
	}
}

// RequireAdmin verifica que el usuario tenga acceso al panel:
// acepta roles "admin", "super_admin" y "producer".
// Establece producer_id en el contexto cuando el rol es "producer".
func RequireAdmin(jwtService *jwt.Service, authService *auth.Service, usersRepo interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		userID, err := claims.GetUserID()
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Verificar que el rol permite acceso al panel
		if !claims.IsProducerOrAdmin() {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Panel access required",
			})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("jti", claims.JTI)
		c.Set("producer_id", claims.ProducerID) // vacío si no es producer

		c.Next()
	}
}

// RequireSuperAdmin verifica que el usuario sea super_admin o admin.
// Usado para rutas de gestión de productores y otras acciones exclusivas de plataforma.
func RequireSuperAdmin(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if !claims.IsSuperAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}

		userID, err := claims.GetUserID()
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("jti", claims.JTI)

		c.Next()
	}
}

