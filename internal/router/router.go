package router

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/qenti/qenti/api/v1/admin"
	appHandlers "github.com/qenti/qenti/api/v1/app"
	authHandlers "github.com/qenti/qenti/api/v1/auth"
	"github.com/qenti/qenti/internal/config"
	"github.com/qenti/qenti/internal/middleware"
	"github.com/qenti/qenti/internal/pkg/auth"
	"github.com/qenti/qenti/internal/pkg/episodes"
	"github.com/qenti/qenti/internal/pkg/invitations"
	"github.com/qenti/qenti/internal/pkg/jwt"
	"github.com/qenti/qenti/internal/pkg/notifications"
	"github.com/qenti/qenti/internal/pkg/payment"
	"github.com/qenti/qenti/internal/pkg/producers"
	"github.com/qenti/qenti/internal/pkg/series"
	"github.com/qenti/qenti/internal/pkg/storage"
	"github.com/qenti/qenti/internal/pkg/unlocks"
	"github.com/qenti/qenti/internal/pkg/users"
)

func SetupRoutes(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Inicializar Firebase Service (puede ser nil si no está configurado)
	var firebaseService *auth.FirebaseService
	if cfg.Firebase.ProjectID != "" {
		var err error
		firebaseService, err = auth.NewFirebaseService(cfg.Firebase.CredentialsPath)
		if err != nil {
			// Log error pero continuar sin Firebase (modo desarrollo)
			// En producción esto debería ser fatal
		}
	}
	
	// Inicializar servicios
	authService := auth.NewService(db, firebaseService)
	jwtService := jwt.NewService(cfg.JWT.SecretKey)
	paymentService := payment.NewService(cfg.RevenueCat)

	// Inicializar servicio de notificaciones (FCM via Firebase Admin SDK)
	notifRepo := notifications.NewRepository(db)
	notifService := notifications.NewService(firebaseService.GetApp(), notifRepo)

	// Inicializar proveedor de video (intercambiable via CDN_PROVIDER env)
	videoProvider, err := storage.NewProvider(cfg)
	if err != nil {
		log.Fatalf("storage provider: %v", err)
	}
	log.Printf("CDN provider: %s", videoProvider.ProviderName())
	
	// Inicializar repositorios
	seriesRepo := series.NewRepository(db)
	episodesRepo := episodes.NewRepository(db)
	usersRepo := users.NewRepository(db)
	unlocksRepo := unlocks.NewRepository(db)
	producersRepo := producers.NewRepository(db)
	invitationsRepo := invitations.NewRepository(db)
	
	// Inicializar handlers de Auth
	authHandlers := authHandlers.NewHandlers(authService, jwtService, db, usersRepo, producersRepo, invitationsRepo, cfg.SuperAdminEmail)
	
	// Inicializar handlers de App
	appHandlers := appHandlers.NewHandlers(
		seriesRepo,
		episodesRepo,
		usersRepo,
		unlocksRepo,
		videoProvider,
		paymentService,
		notifService,
		db,
		cfg,
	)

	// Inicializar handlers de Admin
	adminHandlers := admin.NewHandlers(
		seriesRepo,
		episodesRepo,
		videoProvider,
		notifService,
		cfg.VideoUpload.MaxFileSizeMB,
		cfg.VideoUpload.WarnFileSizeMB,
		cfg.EpisodeCliff.CliffStart,
		cfg.EpisodeCliff.BasePrice,
		cfg.EpisodeCliff.CliffPrice,
	)
	
	// Inicializar handlers de Admin Users
	adminUsersHandlers := admin.NewUsersHandlers(usersRepo, db)
	
	// Inicializar handlers de Admin Dashboard
	adminDashboardHandlers := admin.NewDashboardHandlers(db)

	// Inicializar handlers de Producers (super_admin only)
	adminProducersHandlers := admin.NewProducersHandlers(producersRepo, notifService)
	// Inicializar handlers de MyProducer (el propio productor gestiona sus datos)
	adminMyProducerHandlers := admin.NewMyProducerHandlers(producersRepo)
	// Inicializar handlers de Invitations (tenant admin)
	adminInvitationsHandlers := admin.NewInvitationsHandlers(invitationsRepo)
	// Inicializar handlers de Team (gestión de equipo del tenant)
	adminTeamHandlers := admin.NewTeamHandlers(db)
	
	// Inicializar handlers de Webhook
	webhookHandlers := admin.NewWebhookHandlers(
		paymentService,
		usersRepo,
	)
	
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "qenti-api",
		})
	})
	
	// API v1 - Auth (unificado) con rate limiting
	v1Auth := r.Group("/api/v1/auth")
	v1Auth.Use(middleware.RateLimitMiddleware(5.0, 10)) // 5 requests por segundo, burst de 10
	{
		v1Auth.POST("/login", authHandlers.Login)
		v1Auth.POST("/refresh", authHandlers.Refresh)
		// Onboarding: primer usuario crea su productora (requiere JWT básico)
		v1Auth.POST("/onboarding", middleware.RequireAuth(jwtService), authHandlers.Onboarding)
		// Invitaciones: info pública (no requiere auth) + aceptar (requiere auth)
		v1Auth.GET("/invite/:token", authHandlers.GetInviteInfo)
		v1Auth.POST("/invite/accept", middleware.RequireAuth(jwtService), authHandlers.AcceptInvite)
	}
	
	// API v1 - App endpoints
	v1App := r.Group("/api/v1/app")
	{
		// Endpoints públicos
		v1App.GET("/feed", appHandlers.GetFeed)
		v1App.GET("/series", appHandlers.GetSeries)
		v1App.GET("/series/:id", appHandlers.GetSeriesByID)
		v1App.GET("/series/:id/episodes", appHandlers.GetSeriesEpisodes)
		v1App.GET("/trending", appHandlers.GetTrending)
		v1App.GET("/search", appHandlers.Search)
		v1App.GET("/most-viewed", appHandlers.GetMostViewed)
		v1App.GET("/new-releases", appHandlers.GetNewReleases)
		
		// Endpoints autenticados
		v1AppAuth := v1App.Group("")
		v1AppAuth.Use(middleware.RequireAuth(jwtService))
		{
			// Episodios
			v1AppAuth.GET("/episodes/:id/stream", appHandlers.GetEpisodeStream)
			v1AppAuth.POST("/episodes/:id/unlock", middleware.RateLimitMiddleware(2.0, 5), appHandlers.UnlockEpisode)
			v1AppAuth.POST("/episodes/:id/progress", appHandlers.UpdateWatchProgress)

			// Anuncios con rate limiting más estricto
			v1AppAuth.POST("/ads/unlock-episode", middleware.RateLimitMiddleware(1.0, 3), appHandlers.UnlockEpisodeWithAd)
			v1AppAuth.POST("/ads/reward-coins", middleware.RateLimitMiddleware(1.0, 3), appHandlers.RewardCoinsForAd)
			
			// Wallet
			v1AppAuth.GET("/wallet", appHandlers.GetWallet)
			v1AppAuth.GET("/wallet/history", appHandlers.GetWalletHistory)
			
			// Payment
			v1AppAuth.GET("/payment/subscription-status", appHandlers.GetSubscriptionStatus)
			v1AppAuth.GET("/payment/offer", appHandlers.GetOffer)

			// Push notifications (FCM device token)
			v1AppAuth.POST("/device-token", appHandlers.RegisterDeviceToken)
			v1AppAuth.DELETE("/device-token", appHandlers.UnregisterDeviceToken)

			// Usuario
			v1AppAuth.GET("/user/profile", appHandlers.GetUserProfile)

			// Continuar viendo
			v1AppAuth.GET("/continue-watching", appHandlers.GetContinueWatching)

			// Favoritos
			v1AppAuth.GET("/favorites", appHandlers.GetFavorites)
			v1AppAuth.POST("/favorites/:series_id", appHandlers.ToggleFavorite)
		}
	}
	
	// API v1 - Admin endpoints (requieren rol admin)
	v1Admin := r.Group("/api/v1/admin")
	v1Admin.Use(middleware.RequireAdmin(jwtService, authService, usersRepo))
	v1Admin.Use(middleware.RateLimitMiddleware(10.0, 20)) // Rate limit más generoso para admin
	{
		// Dashboard
		v1Admin.GET("/dashboard", adminDashboardHandlers.GetDashboard)
		// Estado del productor — permite polling desde StatusScreen sin re-login
		v1Admin.GET("/producer-status", adminDashboardHandlers.GetProducerStatus)
		v1Admin.GET("/my-producer", adminMyProducerHandlers.GetMyProducer)
		v1Admin.PUT("/my-producer", adminMyProducerHandlers.UpdateMyProducer)
		
		// Series CRUD
		v1Admin.GET("/series", adminHandlers.GetSeries)
		v1Admin.GET("/series/:id", adminHandlers.GetSeriesByID)
		v1Admin.POST("/series", adminHandlers.CreateSeries)
		v1Admin.PUT("/series/:id", adminHandlers.UpdateSeries)
		v1Admin.DELETE("/series/:id", adminHandlers.DeleteSeries)
		
		// Episodes CRUD
		v1Admin.GET("/episodes", adminHandlers.GetEpisodes)
		v1Admin.GET("/episodes/:id", adminHandlers.GetEpisodeByID)
		v1Admin.POST("/episodes", adminHandlers.CreateEpisode)
		v1Admin.PUT("/episodes/:id", adminHandlers.UpdateEpisode)
		v1Admin.DELETE("/episodes/:id", adminHandlers.DeleteEpisode)
		
		// Video upload flow (específico por episodio)
		v1Admin.POST("/episodes/:id/upload-url", adminHandlers.GetUploadURL)
		v1Admin.POST("/episodes/:id/upload", adminHandlers.UploadVideo)
		v1Admin.POST("/episodes/:id/complete", adminHandlers.CompleteUpload)
		
		// Validación de servicios
		v1Admin.GET("/validate/bunny", adminHandlers.ValidateBunnyConnection)
		
		// Users management
		v1Admin.GET("/users", adminUsersHandlers.GetUsers)
		v1Admin.GET("/users/:id", adminUsersHandlers.GetUserByID)
		v1Admin.PUT("/users/:id/ban", adminUsersHandlers.BanUser)
		v1Admin.PUT("/users/:id/coins", adminUsersHandlers.GiftCoins)

		// Invitaciones: el tenant admin genera links para su equipo
		v1Admin.POST("/invitations", adminInvitationsHandlers.CreateInvite)
		v1Admin.GET("/invitations", adminInvitationsHandlers.ListInvites)

		// Equipo: miembros actuales del tenant
		v1Admin.GET("/team", adminTeamHandlers.GetTeamMembers)
		v1Admin.DELETE("/team/:userId", adminTeamHandlers.RemoveMember)
	}

	// API v1 - Super Admin: gestión de productores (sólo super_admin/admin)
	v1SuperAdmin := r.Group("/api/v1/admin/producers")
	v1SuperAdmin.Use(middleware.RequireSuperAdmin(jwtService))
	{
		v1SuperAdmin.GET("", adminProducersHandlers.GetProducers)
		v1SuperAdmin.GET("/:id", adminProducersHandlers.GetProducerByID)
		v1SuperAdmin.POST("", adminProducersHandlers.CreateProducer)
		v1SuperAdmin.PUT("/:id", adminProducersHandlers.UpdateProducer)
		v1SuperAdmin.DELETE("/:id", adminProducersHandlers.DeleteProducer)
		// Flujo de aprobación
		v1SuperAdmin.PUT("/:id/approve", adminProducersHandlers.ApproveProducer)
		v1SuperAdmin.PUT("/:id/suspend", adminProducersHandlers.SuspendProducer)
	}
	
	// Webhooks (sin autenticación estándar, usan firma propia)
	webhooks := r.Group("/api/v1/webhooks")
	{
		webhooks.POST("/revenuecat", webhookHandlers.HandleRevenueCatWebhook)
	}
}

