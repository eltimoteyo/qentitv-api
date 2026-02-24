package auth

import (
	"context"
	"log"
	"os"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
)

type FirebaseService struct {
	app  *firebase.App
	auth *auth.Client
}

func NewFirebaseService(credentialsPath string) (*FirebaseService, error) {
	ctx := context.Background()
	
	// Configurar opciones de Firebase
	// En v4, las opciones se pasan directamente al NewApp
	var app *firebase.App
	var err error
	
	// Si existe el archivo de credenciales, configurar variable de entorno
	if credentialsPath != "" {
		if _, err := os.Stat(credentialsPath); err == nil {
			// Establecer variable de entorno para que Firebase la use automáticamente
			os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath)
		} else {
			log.Printf("⚠️  Firebase credentials file not found at %s, using default credentials", credentialsPath)
		}
	}
	
	// Firebase v4 usa GOOGLE_APPLICATION_CREDENTIALS automáticamente
	app, err = firebase.NewApp(ctx, nil)
	
	if err != nil {
		return nil, err
	}
	
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	
	return &FirebaseService{
		app:  app,
		auth: authClient,
	}, nil
}

// GetApp devuelve el *firebase.App subyacente para reutilizarlo en otros servicios
// (ej. messaging/FCM). Devuelve nil si el servicio no está inicializado.
func (f *FirebaseService) GetApp() *firebase.App {
	if f == nil {
		return nil
	}
	return f.app
}

// VerifyIDToken verifica un token de Firebase y retorna los claims
func (f *FirebaseService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := f.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// GetUserInfo extrae información del usuario del token de Firebase
func (f *FirebaseService) GetUserInfo(ctx context.Context, idToken string) (*UserInfo, error) {
	token, err := f.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	
	// Obtener email del token
	email := ""
	if emailClaim, ok := token.Claims["email"].(string); ok {
		email = emailClaim
	}
	
	// Verificar si es admin desde custom claims (variable no usada, se verifica en el servicio)
	_ = false
	if adminClaim, ok := token.Claims["admin"].(bool); ok && adminClaim {
		_ = true
	}
	if roleClaim, ok := token.Claims["role"].(string); ok && roleClaim == "admin" {
		_ = true
	}
	
	return &UserInfo{
		ID:          uuid.New(), // Se actualizará cuando se busque en DB
		FirebaseUID: token.UID,
		Email:       email,
		IsPremium:   false, // Se actualizará cuando se busque en DB
	}, nil
}

// SetCustomClaims establece custom claims para un usuario
func (f *FirebaseService) SetCustomClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	return f.auth.SetCustomUserClaims(ctx, uid, claims)
}

