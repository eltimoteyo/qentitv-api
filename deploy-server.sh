#!/bin/bash

# Script de despliegue para servidor VPS
# Uso: ./deploy-server.sh

set -e

echo "🚀 Iniciando despliegue en servidor..." 
echo ""

# Verificar que estamos en el directorio correcto
if [ ! -f "docker-compose.prod.yml" ]; then
    echo "❌ Error: docker-compose.prod.yml no encontrado"
    echo "   Asegúrate de estar en el directorio raíz del proyecto"
    exit 1
fi

# Verificar que existe .env.production
if [ ! -f ".env.production" ]; then
    echo "❌ Error: Archivo .env.production no encontrado"
    echo "   Crea el archivo .env.production con las variables necesarias"
    exit 1
fi

echo "✅ Archivo .env.production encontrado"

# Verificar Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker no está instalado"
    echo "   Instala Docker: curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: Docker Compose no está instalado"
    exit 1
fi

echo "✅ Docker y Docker Compose encontrados"
echo ""

# Detener contenedores existentes
echo "🛑 Deteniendo contenedores existentes..."
docker-compose -f docker-compose.prod.yml --env-file .env.production down 2>/dev/null || true

# Construir y desplegar
echo ""
echo "🔨 Construyendo imágenes..."
docker-compose -f docker-compose.prod.yml --env-file .env.production build --no-cache

echo ""
echo "🚀 Iniciando servicios..."
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d

echo ""
echo "⏳ Esperando a que los servicios estén listos..."
sleep 15

# Verificar health check
echo ""
echo "🔍 Verificando health check..."

API_PORT=$(grep API_PORT .env.production | cut -d '=' -f2 | tr -d ' ')
API_PORT=${API_PORT:-8080}

MAX_ATTEMPTS=10
ATTEMPT=0
SUCCESS=false

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    if curl -f http://localhost:$API_PORT/health > /dev/null 2>&1; then
        SUCCESS=true
        break
    fi
    ATTEMPT=$((ATTEMPT + 1))
    echo "   Intento $ATTEMPT/$MAX_ATTEMPTS..."
    sleep 3
done

if [ "$SUCCESS" = true ]; then
    echo "✅ API desplegada correctamente!"
    echo ""
    echo "🌐 API disponible en: http://localhost:$API_PORT"
    echo ""
else
    echo "⚠️  Health check no respondió, pero los servicios están iniciados"
    echo "   Revisa los logs: docker-compose -f docker-compose.prod.yml logs -f api"
    echo ""
fi

# Mostrar estado
echo "📊 Estado de los servicios:"
docker-compose -f docker-compose.prod.yml ps

echo ""
echo "🎉 Despliegue completado!"
echo ""
echo "📋 Comandos útiles:"
echo "   Ver logs:        docker-compose -f docker-compose.prod.yml logs -f api"
echo "   Detener:         docker-compose -f docker-compose.prod.yml down"
echo "   Reiniciar:       docker-compose -f docker-compose.prod.yml restart api"
echo "   Estado:          docker-compose -f docker-compose.prod.yml ps"
echo ""
