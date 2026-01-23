#!/bin/bash

# Script de despliegue para Qenti API
# Uso: ./deploy.sh [dev|prod]

set -e

ENV=${1:-dev}

echo "🚀 Iniciando despliegue en modo: $ENV"

# Verificar que existe .env
if [ ! -f .env ]; then
    echo "❌ Error: Archivo .env no encontrado"
    echo "📝 Crea un archivo .env basado en .env.example"
    exit 1
fi

# Verificar que existe firebase-credentials.json
if [ ! -f firebase-credentials.json ]; then
    echo "⚠️  Advertencia: firebase-credentials.json no encontrado"
    echo "📝 Asegúrate de tener el archivo de credenciales de Firebase"
fi

# Cargar variables de entorno
export $(cat .env | grep -v '^#' | xargs)

# Verificar variables críticas
if [ -z "$JWT_SECRET" ]; then
    echo "❌ Error: JWT_SECRET no está configurado en .env"
    exit 1
fi

if [ -z "$DB_PASSWORD" ]; then
    echo "❌ Error: DB_PASSWORD no está configurado en .env"
    exit 1
fi

echo "✅ Variables de entorno verificadas"

# Construir imagen Docker
echo "🔨 Construyendo imagen Docker..."
docker build -t qenti-api:latest .

if [ "$ENV" = "prod" ]; then
    echo "🏭 Desplegando en modo PRODUCCIÓN..."
    docker-compose -f docker-compose.prod.yml up -d
    
    echo "⏳ Esperando a que los servicios estén listos..."
    sleep 10
    
    echo "🔍 Verificando health check..."
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ API desplegada correctamente!"
    else
        echo "❌ Error: Health check falló"
        echo "📋 Revisa los logs con: docker-compose -f docker-compose.prod.yml logs -f api"
        exit 1
    fi
else
    echo "💻 Desplegando en modo DESARROLLO..."
    docker-compose up -d postgres
    
    echo "⏳ Esperando a que PostgreSQL esté listo..."
    sleep 5
    
    echo "✅ PostgreSQL listo"
    echo "📝 Para ejecutar la API localmente:"
    echo "   go run cmd/server/main.go"
    echo ""
    echo "   O con Docker:"
    echo "   docker-compose up api"
fi

echo ""
echo "🎉 Despliegue completado!"
echo ""
echo "📊 Comandos útiles:"
echo "   Ver logs: docker-compose logs -f"
echo "   Detener: docker-compose down"
echo "   Estado: docker-compose ps"

