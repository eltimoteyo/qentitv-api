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
		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("jti", claims.JTI)

		c.Next()
	}
}

// RequireAdmin verifica que el usuario sea administrador
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

		// Extraer userID desde el campo 'sub'
		userID, err := claims.GetUserID()
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Verificar si el usuario es admin desde el token
		if !claims.IsAdmin() {
			// Verificar también en DB por si acaso el rol cambió después del token
			// Necesitamos obtener el usuario para tener firebase_uid
			// Por ahora confiamos en el token, pero podemos mejorar esto después
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("jti", claims.JTI)

		c.Next()
	}
}

