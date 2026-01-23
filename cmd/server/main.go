package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/qenti/qenti/internal/config"
	"github.com/qenti/qenti/internal/database"
	"github.com/qenti/qenti/internal/middleware"
	"github.com/qenti/qenti/internal/router"
)

// validateCriticalConfig valida que las variables críticas estén configuradas
func validateCriticalConfig(cfg *config.Config) error {
	var missing []string

	// Validar configuración de base de datos
	if cfg.Database.Host == "" {
		missing = append(missing, "DB_HOST")
	}
	if cfg.Database.User == "" {
		missing = append(missing, "DB_USER")
	}
	if cfg.Database.Password == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if cfg.Database.Name == "" {
		missing = append(missing, "DB_NAME")
	}

	// Validar JWT Secret (crítico para producción)
	if cfg.JWT.SecretKey == "" || cfg.JWT.SecretKey == "change-this-secret-key-in-production" {
		if cfg.Environment == "production" {
			missing = append(missing, "JWT_SECRET (must be set in production)")
		} else {
			log.Println("⚠️  WARNING: JWT_SECRET is using default value. Change it in production!")
		}
	}

	// Validar Bunny.net (crítico para streaming)
	if cfg.Bunny.StreamAPIKey == "" {
		log.Println("⚠️  WARNING: BUNNY_STREAM_API_KEY not set. Video streaming will not work.")
	}
	if cfg.Bunny.StreamLibraryID == "" {
		log.Println("⚠️  WARNING: BUNNY_STREAM_LIBRARY_ID not set. Video streaming will not work.")
	}
	if cfg.Bunny.CDNHostname == "" {
		log.Println("⚠️  WARNING: BUNNY_CDN_HOSTNAME not set. Video URLs will not work.")
	}

	// Firebase es opcional (puede funcionar en modo desarrollo sin él)
	if cfg.Firebase.ProjectID == "" {
		log.Println("⚠️  INFO: FIREBASE_PROJECT_ID not set. Auth will use mock mode (development only).")
	}

	// RevenueCat es opcional (solo necesario si usas suscripciones)
	if cfg.RevenueCat.APIKey == "" {
		log.Println("⚠️  INFO: REVENUECAT_API_KEY not set. Payment webhooks will not work.")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Cargar configuración
	cfg := config.Load()

	// Validar variables críticas
	if err := validateCriticalConfig(cfg); err != nil {
		log.Fatalf("❌ Configuration validation failed: %v", err)
	}

	// Inicializar base de datos
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ejecutar migraciones
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Crear router
	r := gin.Default()

	// Middleware global
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Inicializar rutas
	router.SetupRoutes(r, db, cfg)

	// Iniciar servidor
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Qenti server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

