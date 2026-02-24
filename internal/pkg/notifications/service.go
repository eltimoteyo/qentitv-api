package notifications

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
)

// Service envuelve el cliente FCM de Firebase y el repositorio de tokens.
// Si Firebase no está configurado (app == nil), los envíos se ignoran silenciosamente.
type Service struct {
	messaging *messaging.Client
	repo      *Repository
}

// NewService crea un Service. Si app es nil (Firebase no configurado), el
// servicio funciona en modo "no-op" — las notificaciones se descartan con log.
func NewService(app *firebase.App, repo *Repository) *Service {
	svc := &Service{repo: repo}
	if app == nil {
		log.Println("⚠️  notifications: Firebase not configured, push notifications disabled")
		return svc
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("⚠️  notifications: failed to init FCM client: %v — push disabled", err)
		return svc
	}
	svc.messaging = client
	log.Println("✅ notifications: FCM messaging client ready")
	return svc
}

// RegisterToken guarda el token FCM de un dispositivo.
func (s *Service) RegisterToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	return s.repo.SaveToken(ctx, userID, token, platform)
}

// DeleteToken elimina un token FCM.
func (s *Service) DeleteToken(ctx context.Context, token string) error {
	return s.repo.DeleteToken(ctx, token)
}

// NotifyNewEpisode envía una notificación push a todos los usuarios que tienen
// la serie en favoritos, informando del nuevo episodio.
// Se ejecuta de forma best-effort: los errores se loguean pero no se propagan.
func (s *Service) NotifyNewEpisode(
	ctx context.Context,
	seriesID uuid.UUID,
	seriesTitle string,
	episodeNumber int,
	episodeTitle string,
) {
	if s.messaging == nil {
		return
	}

	tokens, err := s.repo.GetTokensForFavoriters(ctx, seriesID)
	if err != nil {
		log.Printf("notifications: failed to get tokens for series %s: %v", seriesID, err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	title := seriesTitle
	body := episodeTitle
	if episodeNumber > 0 {
		body = "Episodio " + itoa(episodeNumber) + ": " + episodeTitle
	}

	// FCM tiene un límite de 500 tokens por llamada MulticastMessage
	for i := 0; i < len(tokens); i += 500 {
		end := i + 500
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]

		msg := &messaging.MulticastMessage{
			Tokens: batch,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: map[string]string{
				"type":      "new_episode",
				"series_id": seriesID.String(),
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "new_episodes",
					Sound:     "default",
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound: "default",
						Badge: func() *int { v := 1; return &v }(),
					},
				},
			},
		}

		resp, err := s.messaging.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("notifications: FCM multicast error: %v", err)
			continue
		}
		if resp.FailureCount > 0 {
			log.Printf("notifications: FCM sent=%d failed=%d for series %s ep %d",
				resp.SuccessCount, resp.FailureCount, seriesID, episodeNumber)
		}
	}
}

// NotifyProducerApproved envía una notificación push al productor cuando su cuenta
// es aprobada por el super_admin. Ejecuta best-effort (errores solo se loguean).
func (s *Service) NotifyProducerApproved(ctx context.Context, producerUserID uuid.UUID, producerName string) {
	if s.messaging == nil {
		return
	}

	tokens, err := s.repo.GetTokensForUser(ctx, producerUserID)
	if err != nil {
		log.Printf("notifications: failed to get tokens for producer user %s: %v", producerUserID, err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	for i := 0; i < len(tokens); i += 500 {
		end := i + 500
		if end > len(tokens) {
			end = len(tokens)
		}
		batch := tokens[i:end]

		msg := &messaging.MulticastMessage{
			Tokens: batch,
			Notification: &messaging.Notification{
				Title: "¡Productora aprobada! 🎉",
				Body:  "Ya puedes publicar tu contenido en QentiTV.",
			},
			Data: map[string]string{
				"type":          "producer_approved",
				"producer_name": producerName,
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					ChannelID: "account_updates",
					Sound:     "default",
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound: "default",
						Badge: func() *int { v := 1; return &v }(),
					},
				},
			},
		}

		resp, err := s.messaging.SendEachForMulticast(ctx, msg)
		if err != nil {
			log.Printf("notifications: FCM error for producer_approved %s: %v", producerUserID, err)
			continue
		}
		if resp.FailureCount > 0 {
			log.Printf("notifications: FCM sent=%d failed=%d for producer_approved %s",
				resp.SuccessCount, resp.FailureCount, producerUserID)
		}
	}
}

// itoa convierte un int a string sin importar strconv para evitar imports innecesarios.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
