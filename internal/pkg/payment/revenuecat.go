package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/qenti/qenti/internal/config"
)

type Service struct {
	config config.RevenueCatConfig
	client *http.Client
}

func NewService(cfg config.RevenueCatConfig) *Service {
	return &Service{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WebhookEvent representa un evento de webhook de RevenueCat
type WebhookEvent struct {
	Event struct {
		ID        string    `json:"id"`
		Type      string    `json:"type"`
		AppUserID string    `json:"app_user_id"`
		ProductID string    `json:"product_id"`
		PeriodType string   `json:"period_type"`
		PurchasedAt time.Time `json:"purchased_at_ms"`
	} `json:"event"`
}

// VerifyWebhookSignature verifica la firma del webhook de RevenueCat
func (s *Service) VerifyWebhookSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(s.config.WebhookSecret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ProcessWebhook procesa un webhook de RevenueCat y retorna información del evento
func (s *Service) ProcessWebhook(body []byte, signature string) (*WebhookEvent, error) {
	// Verificar firma
	if !s.VerifyWebhookSignature(body, signature) {
		return nil, fmt.Errorf("invalid webhook signature")
	}
	
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook: %w", err)
	}
	
	return &event, nil
}

// GetUserSubscriptionStatus obtiene el estado de suscripción de un usuario desde RevenueCat
func (s *Service) GetUserSubscriptionStatus(appUserID string) (bool, error) {
	url := fmt.Sprintf("https://api.revenuecat.com/v1/subscribers/%s", appUserID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("revenuecat API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		Subscriber struct {
			Entitlements map[string]struct {
				ExpiresDate string `json:"expires_date"`
				ProductIdentifier string `json:"product_identifier"`
			} `json:"entitlements"`
		} `json:"subscriber"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Verificar si tiene alguna suscripción activa
	for _, entitlement := range result.Subscriber.Entitlements {
		if entitlement.ExpiresDate != "" {
			expiresAt, err := time.Parse(time.RFC3339, entitlement.ExpiresDate)
			if err == nil && expiresAt.After(time.Now()) {
				return true, nil
			}
		}
	}
	
	return false, nil
}

