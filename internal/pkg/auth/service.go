package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/qenti/qenti/internal/pkg/models"
)

type Service struct {
	db              *sql.DB
	firebaseService *FirebaseService
}

func NewService(db *sql.DB, firebaseService *FirebaseService) *Service {
	return &Service{
		db:              db,
		firebaseService: firebaseService,
	}
}

// UserInfo contiene información del usuario autenticado
type UserInfo struct {
	ID          uuid.UUID
	FirebaseUID string
	Email       string
	IsPremium   bool
}

// VerifyToken verifica un token JWT de Firebase y retorna información del usuario
func (s *Service) VerifyToken(token string) (*UserInfo, error) {
	ctx := context.Background()

	// Si no hay Firebase Admin SDK configurado, decodificar el JWT sin verificar firma.
	// Esto permite desarrollo local con Google Auth real en el frontend.
	// NUNCA usar en producción sin Firebase Admin SDK configurado.
	if s.firebaseService == nil {
		log.Println("⚠️  Firebase Admin SDK no configurado — decodificando token sin verificar (solo dev)")
		firebaseUID, email, err := decodeFirebaseTokenUnsafe(token)
		if err != nil {
			// Fallback último recurso: usuario mock genérico para tests sin frontend
			log.Printf("⚠️  No se pudo decodificar el token: %v — usando mock genérico", err)
			firebaseUID = "mock-firebase-uid-" + uuid.New().String()[:8]
			email = "dev-" + firebaseUID[:16] + "@mock.local"
		}
		user, err := s.GetOrCreateUser(ctx, firebaseUID, email)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create user: %w", err)
		}
		return &UserInfo{
			ID:          user.ID,
			FirebaseUID: user.FirebaseUID,
			Email:       user.Email,
			IsPremium:   user.IsPremium,
		}, nil
	}
	
	// Verificar token con Firebase
	firebaseUser, err := s.firebaseService.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Firebase token: %w", err)
	}
	
	// Obtener o crear usuario en DB
	user, err := s.GetOrCreateUser(ctx, firebaseUser.FirebaseUID, firebaseUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create user: %w", err)
	}
	
	return &UserInfo{
		ID:          user.ID,
		FirebaseUID: user.FirebaseUID,
		Email:       user.Email,
		IsPremium:   user.IsPremium,
	}, nil
}

// GetOrCreateUser obtiene un usuario por Firebase UID o lo crea si no existe
func (s *Service) GetOrCreateUser(ctx context.Context, firebaseUID, email string) (*models.User, error) {
	var user models.User
	
	query := `SELECT id, email, firebase_uid, coin_balance, is_premium, created_at, updated_at 
	          FROM users WHERE firebase_uid = $1`
	
	err := s.db.QueryRowContext(ctx, query, firebaseUID).Scan(
		&user.ID, &user.Email, &user.FirebaseUID, &user.CoinBalance,
		&user.IsPremium, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Crear nuevo usuario con bono de bienvenida
		user.ID = uuid.New()
		user.Email = email
		user.FirebaseUID = firebaseUID
		user.CoinBalance = 50 // Bono de bienvenida: 50 monedas
		user.IsPremium = false
		
		insertQuery := `INSERT INTO users (id, email, firebase_uid, coin_balance, is_premium) 
		                VALUES ($1, $2, $3, $4, $5) RETURNING created_at, updated_at`
		
		err = s.db.QueryRowContext(ctx, insertQuery,
			user.ID, user.Email, user.FirebaseUID, user.CoinBalance, user.IsPremium,
		).Scan(&user.CreatedAt, &user.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		
		// Registrar transacción de bono de bienvenida
		// TODO: Agregar registro de transacción si existe tabla de transacciones
		log.Printf("✅ Nuevo usuario creado: %s con bono de bienvenida de 50 monedas", email)
		
		return &user, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// IsAdmin verifica si un usuario es administrador (retrocompat)
func (s *Service) IsAdmin(firebaseUID string) (bool, error) {
	role, _, _ := s.GetUserRole(firebaseUID)
	return role == "admin" || role == "super_admin", nil
}

// RoleInfo contiene el rol y el producer_id (si aplica) de un usuario.
type RoleInfo struct {
	Role       string // "user", "admin", "super_admin", "producer"
	ProducerID string // UUID string, vacío si el rol no es "producer"
}

// GetUserRole retorna el rol del usuario y, si es producer, su producer_id.
// Prioridad: super_admin > admin > producer > user
// Para producers: busca primero si es dueño de una productora, luego si es miembro invitado.
func (s *Service) GetUserRole(firebaseUID string) (role, producerID string, err error) {
	ctx := context.Background()

	// Busca el rol más prioritario del usuario.
	// El producer_id se toma de producers (dueño) o producer_members (invitado).
	query := `
		SELECT ur.role, COALESCE(
		    (SELECT p.id::text  FROM producers p       WHERE p.user_id  = u.id AND ur.role = 'producer' LIMIT 1),
		    (SELECT pm.producer_id::text FROM producer_members pm WHERE pm.user_id = u.id AND pm.role = ur.role LIMIT 1),
		    ''
		)
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		WHERE u.firebase_uid = $1
		ORDER BY CASE ur.role
			WHEN 'super_admin' THEN 1
			WHEN 'admin'       THEN 2
			WHEN 'producer'    THEN 3
			WHEN 'moderator'   THEN 4
			ELSE 5
		END
		LIMIT 1`

	err = s.db.QueryRowContext(ctx, query, firebaseUID).Scan(&role, &producerID)
	if err == sql.ErrNoRows {
		return "user", "", nil
	}
	if err != nil {
		return "user", "", fmt.Errorf("GetUserRole: %w", err)
	}
	return role, producerID, nil
}

// GrantAdminRole otorga rol de admin a un usuario
func (s *Service) GrantAdminRole(ctx context.Context, userID uuid.UUID, grantedBy uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role, granted_by) VALUES ($1, 'admin', $2) 
	          ON CONFLICT (user_id, role) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, userID, grantedBy)
	return err
}

// GrantRole otorga cualquier rol a un usuario (usado para producer, super_admin)
func (s *Service) GrantRole(ctx context.Context, userID uuid.UUID, roleName string, grantedBy uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role, granted_by) VALUES ($1, $2, $3) 
	          ON CONFLICT (user_id, role) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, userID, roleName, grantedBy)
	return err
}

// AddMemberToProducer agrega un usuario como miembro de un tenant.
// Se usa cuando un usuario acepta una invitación de link.
// También registra el rol en user_roles si no existe.
func (s *Service) AddMemberToProducer(ctx context.Context, producerID, userID uuid.UUID, role string) error {
	// 1. Insertar en producer_members
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO producer_members (producer_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (producer_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		producerID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("AddMemberToProducer members: %w", err)
	}

	// 2. Asegurar que tiene el rol en user_roles (usando el producerID como granted_by)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role) DO NOTHING`,
		userID, role, producerID,
	)
	if err != nil {
		return fmt.Errorf("AddMemberToProducer roles: %w", err)
	}
	return nil
}

// decodeFirebaseTokenUnsafe extrae email y uid de un Firebase ID token (JWT)
// SIN verificar la firma. Solo para modo desarrollo cuando Firebase Admin SDK
// no está configurado. Nunca usar en producción.
func decodeFirebaseTokenUnsafe(token string) (uid, email string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("token no tiene formato JWT (partes: %d)", len(parts))
	}

	// El payload es la segunda parte, codificado en base64url sin padding
	payload := parts[1]
	// Agregar padding si falta
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Intentar con RawURLEncoding (sin padding)
		decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", "", fmt.Errorf("base64 decode: %w", err)
		}
	}

	var claims struct {
		Sub   string `json:"sub"`   // Firebase UID
		Email string `json:"email"` // email del usuario
		UID   string `json:"uid"`   // alternativo en algunos tokens
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", "", fmt.Errorf("json unmarshal: %w", err)
	}

	uid = claims.Sub
	if uid == "" {
		uid = claims.UID
	}
	if uid == "" {
		return "", "", fmt.Errorf("token no contiene 'sub' ni 'uid'")
	}
	if claims.Email == "" {
		return "", "", fmt.Errorf("token no contiene 'email'")
	}

	return uid, claims.Email, nil
}


