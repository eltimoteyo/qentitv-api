package bunny

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ValidateConnection verifica que la conexión con Bunny.net funcione correctamente
func (s *Service) ValidateConnection() error {
	// Intentar obtener información de la librería
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s", s.config.StreamLibraryID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("AccessKey", s.config.StreamAPIKey)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Bunny.net: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bunny API returned status %d - check your API key and library ID", resp.StatusCode)
	}
	
	return nil
}

// GetVideoStatus obtiene el estado de un video en Bunny.net
func (s *Service) GetVideoStatus(videoID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s", s.config.StreamLibraryID, videoID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("AccessKey", s.config.StreamAPIKey)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bunny API error: %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result, nil
}
