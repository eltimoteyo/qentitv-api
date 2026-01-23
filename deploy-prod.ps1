# Script de despliegue para producción - Qenti API
# Uso: .\deploy-prod.ps1

param(
    [string]$EnvFile = ".env.production"
)

Write-Host "🚀 Iniciando despliegue en modo PRODUCCIÓN" -ForegroundColor Cyan
Write-Host ""

# Verificar que existe .env.production
if (-not (Test-Path $EnvFile)) {
    Write-Host "❌ Error: Archivo $EnvFile no encontrado" -ForegroundColor Red
    Write-Host "📝 Crea un archivo $EnvFile con las siguientes variables:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "DB_USER=postgres" -ForegroundColor White
    Write-Host "DB_PASSWORD=tu_password_seguro" -ForegroundColor White
    Write-Host "DB_NAME=qenti" -ForegroundColor White
    Write-Host "JWT_SECRET=tu_jwt_secret_muy_seguro" -ForegroundColor White
    Write-Host "BUNNY_STREAM_LIBRARY_ID=tu_library_id" -ForegroundColor White
    Write-Host "BUNNY_STREAM_API_KEY=tu_api_key" -ForegroundColor White
    Write-Host "BUNNY_CDN_HOSTNAME=tu_cdn_hostname" -ForegroundColor White
    Write-Host "BUNNY_SECURITY_KEY=tu_security_key" -ForegroundColor White
    Write-Host "FIREBASE_PROJECT_ID=tu_project_id (opcional)" -ForegroundColor White
    Write-Host "REVENUECAT_API_KEY=tu_api_key (opcional)" -ForegroundColor White
    Write-Host "REVENUECAT_WEBHOOK_SECRET=tu_secret (opcional)" -ForegroundColor White
    Write-Host ""
    exit 1
}

Write-Host "✅ Archivo $EnvFile encontrado" -ForegroundColor Green

# Verificar que existe firebase-credentials.json (opcional)
if (-not (Test-Path "firebase-credentials.json")) {
    Write-Host "⚠️  Advertencia: firebase-credentials.json no encontrado" -ForegroundColor Yellow
    Write-Host "   El API funcionará en modo mock (sin Firebase)" -ForegroundColor Yellow
    Write-Host ""
}

# Cargar variables de entorno desde .env.production
Write-Host "📋 Cargando variables de entorno..." -ForegroundColor Cyan
$envVars = @{}
Get-Content $EnvFile | ForEach-Object {
    if ($_ -match '^([^#][^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()
        $envVars[$key] = $value
        [Environment]::SetEnvironmentVariable($key, $value, "Process")
    }
}

# Verificar variables críticas
$requiredVars = @("DB_PASSWORD", "JWT_SECRET", "BUNNY_STREAM_LIBRARY_ID", "BUNNY_STREAM_API_KEY", "BUNNY_CDN_HOSTNAME", "BUNNY_SECURITY_KEY")
$missingVars = @()

foreach ($var in $requiredVars) {
    if (-not $envVars.ContainsKey($var) -or [string]::IsNullOrWhiteSpace($envVars[$var])) {
        $missingVars += $var
    }
}

if ($missingVars.Count -gt 0) {
    Write-Host "❌ Error: Variables faltantes en $EnvFile :" -ForegroundColor Red
    $missingVars | ForEach-Object { Write-Host "   - $_" -ForegroundColor Red }
    exit 1
}

Write-Host "✅ Variables de entorno verificadas" -ForegroundColor Green
Write-Host ""

# Verificar Docker
Write-Host "🔍 Verificando Docker..." -ForegroundColor Cyan
try {
    $dockerVersion = docker --version
    Write-Host "✅ Docker encontrado: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Error: Docker no está instalado o no está en PATH" -ForegroundColor Red
    exit 1
}

# Construir imagen Docker
Write-Host ""
Write-Host "🔨 Construyendo imagen Docker..." -ForegroundColor Cyan
docker build -t qenti-api:latest .

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error al construir la imagen Docker" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Imagen construida correctamente" -ForegroundColor Green
Write-Host ""

# Detener contenedores existentes
Write-Host "🛑 Deteniendo contenedores existentes..." -ForegroundColor Cyan
docker-compose -f docker-compose.prod.yml --env-file $EnvFile down

# Desplegar
Write-Host ""
Write-Host "🏭 Desplegando en modo PRODUCCIÓN..." -ForegroundColor Cyan
docker-compose -f docker-compose.prod.yml --env-file $EnvFile up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error al desplegar" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "⏳ Esperando a que los servicios estén listos..." -ForegroundColor Cyan
Start-Sleep -Seconds 10

# Verificar health check
Write-Host ""
Write-Host "🔍 Verificando health check..." -ForegroundColor Cyan
$apiPort = if ($envVars.ContainsKey("API_PORT")) { $envVars["API_PORT"] } else { "8080" }

try {
    $response = Invoke-WebRequest -Uri "http://localhost:$apiPort/health" -TimeoutSec 5 -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ API desplegada correctamente!" -ForegroundColor Green
        Write-Host ""
        Write-Host "🌐 API disponible en: http://localhost:$apiPort" -ForegroundColor Cyan
    } else {
        Write-Host "⚠️  Health check retornó código: $($response.StatusCode)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠️  No se pudo verificar el health check automáticamente" -ForegroundColor Yellow
    Write-Host "   Verifica manualmente: http://localhost:$apiPort/health" -ForegroundColor Yellow
    Write-Host "   O revisa los logs: docker-compose -f docker-compose.prod.yml logs -f api" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Despliegue completado!" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Comandos útiles:" -ForegroundColor Cyan
Write-Host "   Ver logs: docker-compose -f docker-compose.prod.yml logs -f api" -ForegroundColor White
Write-Host "   Detener: docker-compose -f docker-compose.prod.yml down" -ForegroundColor White
Write-Host "   Estado: docker-compose -f docker-compose.prod.yml ps" -ForegroundColor White
Write-Host "   Reiniciar: docker-compose -f docker-compose.prod.yml restart api" -ForegroundColor White
Write-Host ""
