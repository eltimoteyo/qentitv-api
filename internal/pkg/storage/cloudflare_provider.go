package storage

import (
	"fmt"
	"io"
)

// CloudflareProvider es un stub para Cloudflare Stream.
// Implementar reemplazando los métodos con las llamadas reales a la API de Cloudflare.
// Documentación: https://developers.cloudflare.com/stream/
type CloudflareProvider struct {
	AccountID string
	APIToken  string
}

func NewCloudflareProvider(accountID, apiToken string) *CloudflareProvider {
	return &CloudflareProvider{AccountID: accountID, APIToken: apiToken}
}

func (p *CloudflareProvider) ProviderName() string { return "cloudflare" }

func (p *CloudflareProvider) CreateVideo(title string) (*UploadResult, error) {
	// TODO: POST https://api.cloudflare.com/client/v4/accounts/{account_id}/stream
	// Retorna UID del video + URL de upload TUS
	return nil, fmt.Errorf("cloudflare provider: not yet implemented")
}

func (p *CloudflareProvider) UploadVideo(externalID string, data io.Reader, contentType string, contentLength int64) error {
	// TODO: subida TUS o multipart a Cloudflare Stream
	return fmt.Errorf("cloudflare provider: not yet implemented")
}

func (p *CloudflareProvider) GetPlaybackURL(externalID string, expirationMinutes int) (string, error) {
	// TODO: Return HLS URL con token firmado
	// https://customer-{code}.cloudflarestream.com/{uid}/manifest/video.m3u8
	return fmt.Sprintf("https://customer-REPLACE.cloudflarestream.com/%s/manifest/video.m3u8", externalID), nil
}

func (p *CloudflareProvider) DeleteVideo(externalID string) error {
	// TODO: DELETE https://api.cloudflare.com/client/v4/accounts/{account_id}/stream/{identifier}
	return fmt.Errorf("cloudflare provider: not yet implemented")
}

func (p *CloudflareProvider) CompleteUpload(externalID string) error {
	// Cloudflare no requiere un paso explícito de complete
	return nil
}

func (p *CloudflareProvider) ValidateConnection() error {
	// TODO: GET https://api.cloudflare.com/client/v4/accounts/{account_id}/stream
	return fmt.Errorf("cloudflare provider: not yet implemented")
}
