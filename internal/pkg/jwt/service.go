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
	Sub string `json:"sub"` // Formato: "usr_<uuid>" o "adm_<uuid>"
	
	// Role: define permisos
	Role string `json:"role"` // "user" o "admin"
	
	// Email: opcional, útil para auditoría
	Email string `json:"email,omitempty"`
	
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

// GenerateToken genera un nuevo access token JWT según la estructura propuesta
func (s *Service) GenerateToken(userID uuid.UUID, email, role string, expirationHours int) (string, string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	
	// Generar JTI único
	jti, err := GenerateJTI()
	if err != nil {
		return "", "", err
	}
	
	// Determinar prefijo según rol
	subPrefix := "usr_"
	if role == "admin" {
		subPrefix = "adm_"
	}
	
	claims := &Claims{
		Sub:   subPrefix + userID.String(),
		Role:  role,
		Email: email,
		JTI:   jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "qenti-api",
			Subject:   subPrefix + userID.String(),
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

// IsAdmin verifica si el rol en los claims es admin
func (c *Claims) IsAdmin() bool {
	return c.Role == "admin"
}

// GetUserID extrae el UUID del usuario desde el campo 'sub'
func (c *Claims) GetUserID() (uuid.UUID, error) {
	// Remover prefijo "usr_" o "adm_"
	sub := c.Sub
	if len(sub) > 4 {
		sub = sub[4:] // Remover prefijo
	}
	return uuid.Parse(sub)
}

