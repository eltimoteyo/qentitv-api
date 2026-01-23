# Script de despliegue local para Qenti API
# Uso: .\deploy-local.ps1

Write-Host "🚀 Desplegando Qenti API localmente..." -ForegroundColor Cyan

# Verificar Docker
Write-Host "`n📦 Verificando Docker..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
    Write-Host "✅ Docker está corriendo" -ForegroundColor Green
} catch {
    Write-Host "❌ Error: Docker Desktop no está corriendo" -ForegroundColor Red
    Write-Host "   Por favor, inicia Docker Desktop y vuelve a intentar" -ForegroundColor Yellow
    exit 1
}

# Verificar .env
Write-Host "`n📝 Verificando configuración..." -ForegroundColor Yellow
if (-not (Test-Path .env)) {
    Write-Host "⚠️  Archivo .env no encontrado. Creando uno básico..." -ForegroundColor Yellow
    @"
APP_ENV=development
PORT=8080

DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=qenti
DB_SSLMODE=disable

JWT_SECRET=dev-secret-key-change-in-production-min-32-chars

BUNNY_STREAM_LIBRARY_ID=
BUNNY_STREAM_API_KEY=
BUNNY_CDN_HOSTNAME=
BUNNY_SECURITY_KEY=

FIREBASE_PROJECT_ID=
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json

REVENUECAT_API_KEY=
REVENUECAT_WEBHOOK_SECRET=
"@ | Out-File -FilePath .env -Encoding utf8
    Write-Host "✅ Archivo .env creado" -ForegroundColor Green
} else {
    Write-Host "✅ Archivo .env existe" -ForegroundColor Green
}

# Detener contenedores existentes
Write-Host "`n🛑 Deteniendo contenedores existentes..." -ForegroundColor Yellow
docker-compose down 2>&1 | Out-Null

# Construir imagen
Write-Host "`n🔨 Construyendo imagen Docker..." -ForegroundColor Yellow
docker-compose build

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error al construir la imagen" -ForegroundColor Red
    exit 1
}

# Iniciar servicios
Write-Host "`n🚀 Iniciando servicios..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error al iniciar servicios" -ForegroundColor Red
    exit 1
}

# Esperar a que PostgreSQL esté listo
Write-Host "`n⏳ Esperando a que PostgreSQL esté listo..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    Start-Sleep -Seconds 2
    $attempt++
    try {
        $result = docker exec qenti-postgres pg_isready -U postgres 2>&1
        if ($result -match "accepting connections") {
            $ready = $true
            Write-Host "✅ PostgreSQL está listo" -ForegroundColor Green
        }
    } catch {
        # Continuar intentando
    }
    Write-Host "   Intento $attempt/$maxAttempts..." -ForegroundColor Gray
}

if (-not $ready) {
    Write-Host "⚠️  PostgreSQL puede no estar completamente listo, pero continuando..." -ForegroundColor Yellow
}

# Verificar health check
Write-Host "`n🔍 Verificando health check..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

try {
    $response = Invoke-WebRequest -Uri http://localhost:8080/health -TimeoutSec 5 -ErrorAction Stop
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ API está respondiendo correctamente!" -ForegroundColor Green
        Write-Host "   Health check: OK" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️  API aún no está respondiendo (puede estar iniciando)" -ForegroundColor Yellow
    Write-Host "   Revisa los logs con: docker-compose logs -f api" -ForegroundColor Cyan
}

# Mostrar estado
Write-Host "`n📊 Estado de los servicios:" -ForegroundColor Cyan
docker-compose ps

Write-Host "`n✅ Despliegue completado!" -ForegroundColor Green
Write-Host "`n📋 Comandos útiles:" -ForegroundColor Cyan
Write-Host "   Ver logs:        docker-compose logs -f" -ForegroundColor White
Write-Host "   Ver logs API:    docker-compose logs -f api" -ForegroundColor White
Write-Host "   Detener:         docker-compose down" -ForegroundColor White
Write-Host "   Health check:    http://localhost:8080/health" -ForegroundColor White
Write-Host "   pgAdmin:         http://localhost:5050" -ForegroundColor White
Write-Host "`n🎉 ¡Listo para usar!" -ForegroundColor Green

