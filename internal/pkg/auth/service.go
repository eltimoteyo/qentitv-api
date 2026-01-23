package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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
	
	// Si no hay Firebase service configurado, usar mock para desarrollo
	if s.firebaseService == nil {
		log.Println("⚠️  Firebase Auth not configured - using mock")
		return &UserInfo{
			ID:          uuid.New(),
			FirebaseUID: "mock-firebase-uid",
			Email:       "test@example.com",
			IsPremium:   false,
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

// IsAdmin verifica si un usuario es administrador
func (s *Service) IsAdmin(firebaseUID string) (bool, error) {
	ctx := context.Background()
	
	// Primero verificar en la tabla de roles de la DB
	var role string
	query := `SELECT role FROM user_roles WHERE user_id = (SELECT id FROM users WHERE firebase_uid = $1) AND role = 'admin' LIMIT 1`
	err := s.db.QueryRowContext(ctx, query, firebaseUID).Scan(&role)
	if err == nil && role == "admin" {
		return true, nil
	}
	
	// Si no está en DB, verificar custom claims de Firebase (si está configurado)
	if s.firebaseService != nil {
		// Obtener usuario de Firebase para verificar claims
		// Esto se hace mejor durante VerifyToken, pero por compatibilidad lo dejamos aquí
		// En la práctica, el rol debería venir del JWT generado después de login
	}
	
	return false, nil
}

// GrantAdminRole otorga rol de admin a un usuario
func (s *Service) GrantAdminRole(ctx context.Context, userID uuid.UUID, grantedBy uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role, granted_by) VALUES ($1, 'admin', $2) 
	          ON CONFLICT (user_id, role) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, userID, grantedBy)
	return err
}

