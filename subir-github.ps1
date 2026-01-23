# Script para preparar y subir a GitHub
# Uso: .\subir-github.ps1

Write-Host "📤 Preparando repositorio para GitHub..." -ForegroundColor Cyan
Write-Host ""

# Verificar si Git está inicializado
if (-not (Test-Path .git)) {
    Write-Host "🔧 Inicializando Git..." -ForegroundColor Yellow
    git init
    Write-Host "✅ Git inicializado" -ForegroundColor Green
} else {
    Write-Host "✅ Git ya está inicializado" -ForegroundColor Green
}

Write-Host ""

# Verificar .gitignore
Write-Host "🔍 Verificando .gitignore..." -ForegroundColor Cyan
if (-not (Test-Path .gitignore)) {
    Write-Host "⚠️  .gitignore no encontrado, creando uno básico..." -ForegroundColor Yellow
    @"
# Binarios
*.exe
main.exe
qenti-api

# Variables de entorno
.env
.env.local
.env.production
.env.*.local

# Credenciales
firebase-credentials.json
*.pem
*.key

# Logs
*.log
logs/
"@ | Out-File -FilePath .gitignore -Encoding utf8
    Write-Host "✅ .gitignore creado" -ForegroundColor Green
} else {
    Write-Host "✅ .gitignore existe" -ForegroundColor Green
}

Write-Host ""

# Verificar que archivos sensibles NO estén en staging
Write-Host "🔒 Verificando archivos sensibles..." -ForegroundColor Cyan
$sensitiveFiles = @(".env.production", "firebase-credentials.json", "*.exe")
$foundSensitive = $false

foreach ($pattern in $sensitiveFiles) {
    $files = Get-ChildItem -Path . -Filter $pattern -Recurse -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notmatch "\.git" }
    if ($files) {
        Write-Host "⚠️  Advertencia: Se encontraron archivos sensibles: $pattern" -ForegroundColor Yellow
        $foundSensitive = $true
    }
}

if ($foundSensitive) {
    Write-Host ""
    Write-Host "❌ NO hagas push hasta que estos archivos estén en .gitignore" -ForegroundColor Red
    Write-Host "   Verifica que .gitignore incluya estos patrones" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ No se encontraron archivos sensibles" -ForegroundColor Green
Write-Host ""

# Agregar archivos
Write-Host "📦 Agregando archivos..." -ForegroundColor Cyan
git add .

Write-Host "✅ Archivos agregados" -ForegroundColor Green
Write-Host ""

# Verificar estado
Write-Host "📋 Archivos que se van a subir:" -ForegroundColor Cyan
git status --short | Select-Object -First 20
Write-Host ""

# Preguntar si continuar
$continue = Read-Host "¿Continuar con commit? (s/n)"
if ($continue -ne "s") {
    Write-Host "❌ Cancelado" -ForegroundColor Red
    exit 0
}

# Commit
Write-Host ""
Write-Host "💾 Creando commit..." -ForegroundColor Cyan
$commitMessage = Read-Host "Mensaje de commit (o presiona Enter para usar mensaje por defecto)"
if ([string]::IsNullOrWhiteSpace($commitMessage)) {
    $commitMessage = "QENTITV API - Lista para desplegar"
}

git commit -m $commitMessage

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Error al crear commit" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Commit creado" -ForegroundColor Green
Write-Host ""

# Verificar si hay remote
$remoteUrl = git remote get-url origin 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  No hay remote configurado" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Para agregar remote, ejecuta:" -ForegroundColor Cyan
    Write-Host "   git remote add origin https://github.com/TU_USUARIO/qentitv-api.git" -ForegroundColor White
    Write-Host ""
    Write-Host "Luego ejecuta:" -ForegroundColor Cyan
    Write-Host "   git branch -M main" -ForegroundColor White
    Write-Host "   git push -u origin main" -ForegroundColor White
    Write-Host ""
    exit 0
}

Write-Host "✅ Remote configurado: $remoteUrl" -ForegroundColor Green
Write-Host ""

# Preguntar si hacer push
$push = Read-Host "¿Hacer push a GitHub? (s/n)"
if ($push -ne "s") {
    Write-Host "✅ Listo para hacer push manualmente" -ForegroundColor Green
    Write-Host "   Ejecuta: git push -u origin main" -ForegroundColor Cyan
    exit 0
}

# Push
Write-Host ""
Write-Host "🚀 Haciendo push a GitHub..." -ForegroundColor Cyan
git branch -M main 2>$null
git push -u origin main

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Código subido a GitHub exitosamente!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Próximos pasos:" -ForegroundColor Cyan
    Write-Host "   1. Conectarse al VPS: ssh root@TU_IP_VPS" -ForegroundColor White
    Write-Host "   2. Clonar repositorio: git clone https://github.com/TU_USUARIO/qentitv-api.git" -ForegroundColor White
    Write-Host "   3. Seguir instrucciones en DEPLOY_HOSTINGER.md" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "❌ Error al hacer push" -ForegroundColor Red
    Write-Host "   Verifica:" -ForegroundColor Yellow
    Write-Host "   - Que el repositorio exista en GitHub" -ForegroundColor Yellow
    Write-Host "   - Que tengas permisos" -ForegroundColor Yellow
    Write-Host "   - Que la autenticación esté configurada" -ForegroundColor Yellow
}
