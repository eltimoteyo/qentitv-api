package main

import (
	"fmt"
	"log"
	"os"

	"github.com/qenti/qenti/internal/config"
	"github.com/qenti/qenti/internal/pkg/bunny"
)

func main() {
	cfg := config.Load()
	
	bunnyService := bunny.NewService(cfg.Bunny)
	
	fmt.Println("🔍 Validando conexión con Bunny.net...")
	fmt.Println("")
	
	if err := bunnyService.ValidateConnection(); err != nil {
		fmt.Printf("❌ Error: %s\n", err)
		fmt.Println("")
		fmt.Println("Verifica:")
		fmt.Println("1. BUNNY_STREAM_API_KEY está configurado correctamente")
		fmt.Println("2. BUNNY_STREAM_LIBRARY_ID es válido")
		fmt.Println("3. Tienes conexión a internet")
		os.Exit(1)
	}
	
	fmt.Println("✅ Conexión exitosa con Bunny.net")
	fmt.Println("")
	fmt.Printf("Library ID: %s\n", cfg.Bunny.StreamLibraryID)
	fmt.Printf("CDN Hostname: %s\n", cfg.Bunny.CDNHostname)
	
	// Probar crear un video de prueba
	fmt.Println("")
	fmt.Println("🧪 Probando creación de video...")
	
	result, err := bunnyService.PresignedUploadURL("Test Video")
	if err != nil {
		log.Printf("⚠️  Error al crear video de prueba: %v", err)
		fmt.Println("   (Esto puede ser normal si falta configuración)")
	} else {
		fmt.Printf("✅ Video de prueba creado exitosamente\n")
		fmt.Printf("   Video ID: %s\n", result.VideoID)
		fmt.Printf("   Upload URL: %s\n", result.UploadURL)
	}
	
	fmt.Println("")
	fmt.Println("✨ Validación completada")
}
