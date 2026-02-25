package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/qenti/qenti/internal/config"
)

// BunnyProvider implementa VideoProvider usando Bunny.net Stream.
type BunnyProvider struct {
	cfg    config.BunnyConfig
	client *http.Client
}

func NewBunnyProvider(cfg config.BunnyConfig) *BunnyProvider {
	return &BunnyProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Minute,
		},
	}
}

func (p *BunnyProvider) ProviderName() string { return "bunny" }

// CreateVideo reserva un slot en Bunny Stream y retorna el GUID + URL de upload.
func (p *BunnyProvider) CreateVideo(title string) (*UploadResult, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", p.cfg.StreamLibraryID)

	payload, _ := json.Marshal(map[string]interface{}{"title": title})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("bunny: create video request: %w", err)
	}
	req.Header.Set("AccessKey", p.cfg.StreamAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bunny: create video execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bunny: create video API %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		GUID string `json:"guid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("bunny: decode response: %w", err)
	}

	uploadURL := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", p.cfg.StreamLibraryID, result.GUID)
	return &UploadResult{ExternalID: result.GUID, UploadURL: uploadURL}, nil
}

// UploadVideo sube los bytes al endpoint de Bunny por PUT.
func (p *BunnyProvider) UploadVideo(externalID string, data io.Reader, contentType string, contentLength int64) error {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", p.cfg.StreamLibraryID, externalID)

	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return fmt.Errorf("bunny: upload request: %w", err)
	}
	req.Header.Set("AccessKey", p.cfg.StreamAPIKey)
	req.Header.Set("Content-Type", contentType)
	if contentLength > 0 {
		req.ContentLength = contentLength
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("bunny: upload execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bunny: upload API %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetPlaybackURL genera la URL de reproducción para Bunny Stream CDN.
// Si SecurityKey está configurado, genera token firmado (Advanced SHA256, query-style).
// Si está vacío, devuelve URL directa (cuando tokenAuthEnabled=false en la biblioteca Bunny).
func (p *BunnyProvider) GetPlaybackURL(externalID string, expirationMinutes int) (string, error) {
	if externalID == "" {
		return "", fmt.Errorf("bunny: empty video ID")
	}
	if p.cfg.CDNHostname == "" {
		return "", fmt.Errorf("bunny: BUNNY_CDN_HOSTNAME not configured")
	}

	baseURL := fmt.Sprintf("https://%s/%s/playlist.m3u8", p.cfg.CDNHostname, externalID)

	if p.cfg.SecurityKey == "" {
		// tokenAuthEnabled=false en Bunny → devolver URL directa sin token
		return baseURL, nil
	}

	// tokenAuthEnabled=true → generar token SHA256 (Bunny Advanced Token Auth)
	// Formula: token = Base64Url_NoPadding(SHA256(securityKey + path + expiry))
	expiry := time.Now().Add(time.Duration(expirationMinutes) * time.Minute).Unix()
	expiryStr := strconv.FormatInt(expiry, 10)
	path := fmt.Sprintf("/%s/playlist.m3u8", externalID)

	h := sha256.Sum256([]byte(p.cfg.SecurityKey + path + expiryStr))
	token := base64.RawURLEncoding.EncodeToString(h[:])

	return fmt.Sprintf("%s?token=%s&expires=%s", baseURL, token, expiryStr), nil
}

// DeleteVideo elimina el video de la biblioteca de Bunny Stream.
func (p *BunnyProvider) DeleteVideo(externalID string) error {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", p.cfg.StreamLibraryID, externalID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("bunny: delete request: %w", err)
	}
	req.Header.Set("AccessKey", p.cfg.StreamAPIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("bunny: delete execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bunny: delete API %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ValidateConnection verifica las credenciales listando la biblioteca.
func (p *BunnyProvider) ValidateConnection() error {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s", p.cfg.StreamLibraryID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("bunny: validate request: %w", err)
	}
	req.Header.Set("AccessKey", p.cfg.StreamAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("bunny: validate execute: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bunny: validate API %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// CompleteUpload solicita re-encoding a Bunny (no es crítico si falla).
func (p *BunnyProvider) CompleteUpload(externalID string) error {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s/re-encode", p.cfg.StreamLibraryID, externalID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil // no crítico
	}
	req.Header.Set("AccessKey", p.cfg.StreamAPIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil // no crítico
	}
	defer resp.Body.Close()
	return nil
}
