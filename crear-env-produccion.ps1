# Script para crear archivo .env.production
# Usa las credenciales de Bunny.net que ya tienes

Write-Host "📝 Creando archivo .env.production..." -ForegroundColor Cyan
Write-Host ""

# Generar JWT_SECRET automáticamente
$jwtSecret = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))

# Credenciales de Bunny.net (ya las tienes)
$bunnyLibraryId = "585077"
$bunnyApiKey = "b5d6fea7-1f28-4c2f-b33b36e581d4-0e61-4d28"
$bunnyCdnHostname = "vz-e8e1ad01-079.b-cdn.net"
$bunnySecurityKey = "10f4f6f9-d7be-4f87-9451-da11aeeab667"

# Solicitar información al usuario
Write-Host "Por favor, proporciona la siguiente información:" -ForegroundColor Yellow
Write-Host ""

$dbPassword = Read-Host "Contraseña de PostgreSQL (o presiona Enter para 'postgres')"
if ([string]::IsNullOrWhiteSpace($dbPassword)) {
    $dbPassword = "postgres"
}

$dbUser = Read-Host "Usuario de PostgreSQL (o presiona Enter para 'postgres')"
if ([string]::IsNullOrWhiteSpace($dbUser)) {
    $dbUser = "postgres"
}

$dbName = Read-Host "Nombre de la base de datos (o presiona Enter para 'qenti')"
if ([string]::IsNullOrWhiteSpace($dbName)) {
    $dbName = "qenti"
}

$apiPort = Read-Host "Puerto del API (o presiona Enter para '8080')"
if ([string]::IsNullOrWhiteSpace($apiPort)) {
    $apiPort = "8080"
}

Write-Host ""
Write-Host "Firebase (opcional - presiona Enter para omitir):" -ForegroundColor Yellow
$firebaseProjectId = Read-Host "Firebase Project ID"

Write-Host ""
Write-Host "RevenueCat (opcional - presiona Enter para omitir):" -ForegroundColor Yellow
$revenueCatApiKey = Read-Host "RevenueCat API Key"
$revenueCatWebhookSecret = Read-Host "RevenueCat Webhook Secret"

# Crear contenido del archivo
$envContent = @"
# ============================================
# BASE DE DATOS
# ============================================
DB_USER=$dbUser
DB_PASSWORD=$dbPassword
DB_NAME=$dbName
DB_PORT=5432

# ============================================
# JWT AUTHENTICATION
# ============================================
JWT_SECRET=$jwtSecret

# ============================================
# BUNNY.NET (Video Streaming)
# ============================================
BUNNY_STREAM_LIBRARY_ID=$bunnyLibraryId
BUNNY_STREAM_API_KEY=$bunnyApiKey
BUNNY_CDN_HOSTNAME=$bunnyCdnHostname
BUNNY_SECURITY_KEY=$bunnySecurityKey

# ============================================
# FIREBASE AUTHENTICATION (Opcional)
# ============================================
FIREBASE_PROJECT_ID=$firebaseProjectId
FIREBASE_CREDENTIALS_PATH=

# ============================================
# REVENUECAT (Payments - Opcional)
# ============================================
REVENUECAT_API_KEY=$revenueCatApiKey
REVENUECAT_WEBHOOK_SECRET=$revenueCatWebhookSecret

# ============================================
# API CONFIGURATION
# ============================================
API_PORT=$apiPort
"@

# Escribir archivo
$envContent | Out-File -FilePath ".env.production" -Encoding utf8 -NoNewline

Write-Host ""
Write-Host "✅ Archivo .env.production creado exitosamente!" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Resumen de configuración:" -ForegroundColor Cyan
Write-Host "   DB_USER: $dbUser" -ForegroundColor White
Write-Host "   DB_NAME: $dbName" -ForegroundColor White
Write-Host "   JWT_SECRET: $jwtSecret" -ForegroundColor White
Write-Host "   BUNNY_STREAM_LIBRARY_ID: $bunnyLibraryId" -ForegroundColor White
Write-Host "   API_PORT: $apiPort" -ForegroundColor White
Write-Host ""
Write-Host "🚀 Próximo paso: Ejecuta .\deploy-prod.ps1" -ForegroundColor Cyan
Write-Host ""
