// Package storage define la interfaz abstracta para proveedores de video.
// Swappear el CDN solo requiere una nueva implementación de VideoProvider.
package storage

import "io"

// UploadResult contiene el ID externo del video y la URL de upload directo (si aplica).
type UploadResult struct {
	// ExternalID es el identificador del video en el proveedor (e.g. GUID de Bunny.net)
	ExternalID string
	// UploadURL es la URL a la que el cliente puede hacer PUT directamente (upload directo).
	// Vacío si el proveedor no soporta upload directo.
	UploadURL string
}

// VideoProvider es la interfaz que debe implementar cualquier proveedor de hosting de video.
// Implementaciones actuales: BunnyProvider.  Futuros: CloudflareStreamProvider, MuxProvider, S3Provider.
type VideoProvider interface {
	// CreateVideo reserva un slot en el proveedor y devuelve metadatos de upload.
	CreateVideo(title string) (*UploadResult, error)

	// UploadVideo envía los bytes del video al proveedor.
	UploadVideo(externalID string, data io.Reader, contentType string, contentLength int64) error

	// GetPlaybackURL genera la URL de reproducción (firmada si aplica) con TTL en minutos.
	GetPlaybackURL(externalID string, expirationMinutes int) (string, error)

	// DeleteVideo elimina el video del proveedor.
	DeleteVideo(externalID string) error

	// CompleteUpload señala al proveedor que el upload finalizó (re-encoding, etc.).
	// Algunos proveedores no necesitan este paso; en ese caso retornar nil.
	CompleteUpload(externalID string) error

	// ProviderName retorna el nombre del proveedor para logging/config.
	ProviderName() string

	// ValidateConnection verifica que las credenciales son correctas y el servicio está accesible.
	ValidateConnection() error
}
