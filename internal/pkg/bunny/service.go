package bunny

import (
	"bytes"
	"crypto/hmac"
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

type Service struct {
	config config.BunnyConfig
	client *http.Client
}

func NewService(cfg config.BunnyConfig) *Service {
	return &Service{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Minute, // Timeout largo para uploads grandes
		},
	}
}

// UploadResult contiene la URL de upload y el video ID
type UploadResult struct {
	UploadURL string
	VideoID   string
}

// PresignedUploadURL genera una URL presignada para subir un video directamente a Bunny.net
// Retorna tanto la URL de upload como el video_id para completar el registro después
func (s *Service) PresignedUploadURL(videoTitle string) (*UploadResult, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos", s.config.StreamLibraryID)
	
	payload := map[string]interface{}{
		"title": videoTitle,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("AccessKey", s.config.StreamAPIKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bunny API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var result struct {
		VideoID        string `json:"guid"`
		VideoLibraryID int    `json:"videoLibraryId"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// URL de upload: PUT directamente a esta URL
	uploadURL := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", s.config.StreamLibraryID, result.VideoID)
	
	return &UploadResult{
		UploadURL: uploadURL,
		VideoID:   result.VideoID,
	}, nil
}

// GetSignedPlaybackURL genera una URL firmada con token de expiración para reproducir un video
func (s *Service) GetSignedPlaybackURL(videoID string, expirationMinutes int) (string, error) {
	// Validar que el videoID no esté vacío
	if videoID == "" {
		return "", fmt.Errorf("video ID is empty")
	}
	
	// Validar que el CDN hostname esté configurado
	if s.config.CDNHostname == "" {
		return "", fmt.Errorf("BUNNY_CDN_HOSTNAME is not configured")
	}
	
	expirationTime := time.Now().Add(time.Duration(expirationMinutes) * time.Minute).Unix()
	
	baseURL := fmt.Sprintf("https://%s/%s.mp4", s.config.CDNHostname, videoID)
	
	// Si no hay Security Key configurado, retornar URL sin firma (solo para desarrollo)
	if s.config.SecurityKey == "" {
		return baseURL, nil
	}
	
	// Generar token usando HMAC SHA256 según documentación de Bunny.net
	// Formato: base64(hmac_sha256(path + expiration_time, security_key))
	path := fmt.Sprintf("/%s.mp4", videoID)
	message := path + strconv.FormatInt(expirationTime, 10)
	
	mac := hmac.New(sha256.New, []byte(s.config.SecurityKey))
	mac.Write([]byte(message))
	token := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	
	signedURL := fmt.Sprintf("%s?token=%s&expires=%d", baseURL, token, expirationTime)
	
	return signedURL, nil
}

// CompleteUpload verifica el estado del video y opcionalmente inicia re-encoding
func (s *Service) CompleteUpload(videoID string) error {
	// Primero verificar que el video existe y está procesado
	status, err := s.GetVideoStatus(videoID)
	if err != nil {
		return fmt.Errorf("failed to verify video status: %w", err)
	}
	
	// Verificar estado del video (opcional, para logging)
	if statusState, ok := status["status"].(float64); ok {
		// Estado 4 = procesado completamente
		if statusState != 4 {
			// El video aún se está procesando, pero esto es normal
			// No es un error, solo informativo
		}
	}
	
	// Opcional: Iniciar re-encoding para optimización
	// Esto es opcional y puede omitirse si el video ya está en buen formato
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s/re-encode", s.config.StreamLibraryID, videoID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("AccessKey", s.config.StreamAPIKey)
	
	resp, err := s.client.Do(req)
	if err != nil {
		// No fallar si el re-encoding falla, el video ya está subido
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// No es crítico, el video ya está subido
		return nil
	}
	
	return nil
}

// UploadVideo sube un archivo de video directamente a Bunny.net
// Recibe un io.Reader con el contenido del video y el videoID de Bunny
func (s *Service) UploadVideo(videoID string, videoData io.Reader, contentType string, contentLength int64) error {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", s.config.StreamLibraryID, videoID)
	
	req, err := http.NewRequest("PUT", url, videoData)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("AccessKey", s.config.StreamAPIKey)
	req.Header.Set("Content-Type", contentType)
	if contentLength > 0 {
		req.ContentLength = contentLength
	}
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bunny API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

