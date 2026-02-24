package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	secretKey []byte
}

func NewService(secretKey string) *Service {
	return &Service{
		secretKey: []byte(secretKey),
	}
}

// Claims representa los claims del JWT según la estructura propuesta
type Claims struct {
	// Subject: identificador único del usuario con prefijo
	Sub string `json:"sub"` // Formato: "usr_<uuid>", "adm_<uuid>", "prd_<uuid>", "sad_<uuid>"

	// Role: define permisos ("user", "admin", "producer", "super_admin")
	Role string `json:"role"`

	// Email: opcional, útil para auditoría
	Email string `json:"email,omitempty"`

	// ProducerID: sólo presente cuando Role == "producer"
	ProducerID string `json:"producer_id,omitempty"`

	// JWT ID: único por token para revocación
	JTI string `json:"jti"`

	// Registered claims estándar
	jwt.RegisteredClaims
}

// UserInfo contiene información adicional del usuario (no va en el token)
type UserInfo struct {
	UserID      uuid.UUID
	FirebaseUID string
	IsPremium   bool
}

// GenerateJTI genera un JWT ID único
func GenerateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// subPrefixByRole devuelve el prefijo del campo 'sub' según el rol.
func subPrefixByRole(role string) string {
	switch role {
	case "admin":
		return "adm_"
	case "super_admin":
		return "sad_"
	case "producer":
		return "prd_"
	default:
		return "usr_"
	}
}

// GenerateToken genera un nuevo access token JWT.
// producerID puede ser una UUID string vacía si el usuario no es producer.
func (s *Service) GenerateToken(userID uuid.UUID, email, role, producerID string, expirationHours int) (string, string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	jti, err := GenerateJTI()
	if err != nil {
		return "", "", err
	}

	sub := subPrefixByRole(role) + userID.String()

	claims := &Claims{
		Sub:        sub,
		Role:       role,
		Email:      email,
		ProducerID: producerID,
		JTI:        jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "qenti-api",
			Subject:   sub,
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, jti, nil
}

// GenerateRefreshToken genera un refresh token (string aleatorio)
func GenerateRefreshToken() (string, error) {
	return GenerateJTI()
}

// ValidateToken valida y parsea un token JWT
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// IsAdmin verifica permisos de plataforma (alias legado de IsSuperAdmin)
func (c *Claims) IsAdmin() bool {
	return c.Role == "admin" || c.Role == "super_admin"
}

// IsSuperAdmin verifica si es dueño de la plataforma SaaS (gestiona tenants).
// El rol 'admin' se mantiene como alias legado de 'super_admin'.
func (c *Claims) IsSuperAdmin() bool {
	return c.Role == "admin" || c.Role == "super_admin"
}

// IsProducer verifica si es admin de un tenant/productora (acceso total a su propio contenido).
// En el futuro, los tenants podrán crear sub-usuarios con permisos granulares.
func (c *Claims) IsProducer() bool {
	return c.Role == "producer"
}

// IsProducerOrAdmin verifica si puede acceder al panel (tenant admin o super_admin)
func (c *Claims) IsProducerOrAdmin() bool {
	return c.IsProducer() || c.IsAdmin()
}

// GetUserID extrae el UUID del usuario desde el campo 'sub'
func (c *Claims) GetUserID() (uuid.UUID, error) {
	// Remover prefijo de 4 caracteres (usr_, adm_, prd_, sad_)
	sub := c.Sub
	if len(sub) > 4 {
		sub = sub[4:]
	}
	return uuid.Parse(sub)
}

// GetProducerID devuelve el producer UUID o nil si no aplica
func (c *Claims) GetProducerID() *uuid.UUID {
	if c.ProducerID == "" {
		return nil
	}
	id, err := uuid.Parse(c.ProducerID)
	if err != nil {
		return nil
	}
	return &id
}

