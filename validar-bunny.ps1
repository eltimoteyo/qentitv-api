# Script para validar conexión con Bunny.net
# Configura las variables de entorno y ejecuta la validación

Write-Host "🐰 Validación de Conexión con Bunny.net" -ForegroundColor Cyan
Write-Host ""

# Verificar si las variables ya están configuradas
$bunnyApiKey = $env:BUNNY_STREAM_API_KEY
$bunnyLibraryId = $env:BUNNY_STREAM_LIBRARY_ID

if ([string]::IsNullOrEmpty($bunnyApiKey) -or [string]::IsNullOrEmpty($bunnyLibraryId)) {
    Write-Host "⚠️  Variables de entorno no configuradas" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Por favor, proporciona las credenciales de Bunny.net:" -ForegroundColor Yellow
    Write-Host ""
    
    # Solicitar API Key
    if ([string]::IsNullOrEmpty($bunnyApiKey)) {
        $bunnyApiKey = Read-Host "BUNNY_STREAM_API_KEY"
        $env:BUNNY_STREAM_API_KEY = $bunnyApiKey
    }
    
    # Solicitar Library ID
    if ([string]::IsNullOrEmpty($bunnyLibraryId)) {
        $bunnyLibraryId = Read-Host "BUNNY_STREAM_LIBRARY_ID"
        $env:BUNNY_STREAM_LIBRARY_ID = $bunnyLibraryId
    }
    
    # Solicitar CDN Hostname (opcional pero recomendado)
    $bunnyCdnHostname = $env:BUNNY_CDN_HOSTNAME
    if ([string]::IsNullOrEmpty($bunnyCdnHostname)) {
        $bunnyCdnHostname = Read-Host "BUNNY_CDN_HOSTNAME (opcional, presiona Enter para omitir)"
        if (-not [string]::IsNullOrEmpty($bunnyCdnHostname)) {
            $env:BUNNY_CDN_HOSTNAME = $bunnyCdnHostname
        }
    }
    
    # Solicitar Security Key (opcional)
    $bunnySecurityKey = $env:BUNNY_SECURITY_KEY
    if ([string]::IsNullOrEmpty($bunnySecurityKey)) {
        $bunnySecurityKey = Read-Host "BUNNY_SECURITY_KEY (opcional, presiona Enter para omitir)"
        if (-not [string]::IsNullOrEmpty($bunnySecurityKey)) {
            $env:BUNNY_SECURITY_KEY = $bunnySecurityKey
        }
    }
    
    Write-Host ""
} else {
    Write-Host "✅ Variables de entorno encontradas" -ForegroundColor Green
    Write-Host "   API Key: $($bunnyApiKey.Substring(0, [Math]::Min(10, $bunnyApiKey.Length)))..." -ForegroundColor Gray
    Write-Host "   Library ID: $bunnyLibraryId" -ForegroundColor Gray
    Write-Host ""
}

# Ejecutar validación
Write-Host "🔍 Ejecutando validación..." -ForegroundColor Cyan
Write-Host ""

go run scripts/validate_bunny.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✨ Validación exitosa!" -ForegroundColor Green
    Write-Host ""
    Write-Host "💡 Para hacer estas variables permanentes, agrégalas al sistema:" -ForegroundColor Yellow
    Write-Host "   [System.Environment]::SetEnvironmentVariable('BUNNY_STREAM_API_KEY', 'TU_API_KEY', 'User')" -ForegroundColor Gray
    Write-Host "   [System.Environment]::SetEnvironmentVariable('BUNNY_STREAM_LIBRARY_ID', 'TU_LIBRARY_ID', 'User')" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "❌ Validación falló. Revisa las credenciales." -ForegroundColor Red
    Write-Host ""
    Write-Host "📖 Consulta docs/BUNNY_SETUP.md para más información" -ForegroundColor Yellow
}
