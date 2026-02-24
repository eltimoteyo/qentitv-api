package storage

import (
	"fmt"

	"github.com/qenti/qenti/internal/config"
)

// NewProvider construye el VideoProvider correcto según cfg.CDNProvider.
// Para agregar un nuevo proveedor:
//  1. Crear <nombre>_provider.go implementando VideoProvider
//  2. Agregar el case aquí
//  3. Setear CDN_PROVIDER=<nombre> en .env
func NewProvider(cfg *config.Config) (VideoProvider, error) {
	switch cfg.CDNProvider {
	case "bunny", "":
		// Default: Bunny.net
		return NewBunnyProvider(cfg.Bunny), nil
	case "cloudflare":
		return NewCloudflareProvider(
			cfg.Cloudflare.AccountID,
			cfg.Cloudflare.APIToken,
		), nil
	default:
		return nil, fmt.Errorf("storage: unknown CDN_PROVIDER=%q (valid: bunny, cloudflare)", cfg.CDNProvider)
	}
}
