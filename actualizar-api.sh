#!/bin/bash

# Script para actualizar el API desde GitHub
# Uso: ./actualizar-api.sh

set -e

echo "🔄 Actualizando API desde GitHub..."
echo ""

# Verificar que estamos en el directorio correcto
if [ ! -f "docker-compose.prod.yml" ]; then
    echo "❌ Error: No estás en el directorio del proyecto"
    exit 1
fi

# Configurar estrategia de merge
echo "⚙️  Configurando Git..."
git config pull.rebase false

# Verificar estado
echo ""
echo "📋 Estado actual:"
git status --short

# Hacer pull
echo ""
echo "⬇️  Descargando cambios desde GitHub..."
if git pull origin main; then
    echo "✅ Cambios descargados correctamente"
else
    echo ""
    echo "⚠️  Hay conflictos o ramas divergentes"
    echo ""
    echo "Opciones:"
    echo "  1. Merge (recomendado): git pull --no-rebase origin main"
    echo "  2. Descartar cambios locales: git reset --hard origin/main && git pull origin main"
    echo ""
    read -p "¿Deseas descartar cambios locales y usar solo GitHub? (s/n): " discard
    
    if [ "$discard" = "s" ]; then
        echo "🔄 Descartando cambios locales..."
        git reset --hard origin/main
        git pull origin main
        echo "✅ Cambios descartados y actualizado"
    else
        echo "❌ Actualización cancelada. Resuelve los conflictos manualmente."
        exit 1
    fi
fi

# Reconstruir y redesplegar
echo ""
echo "🔨 Reconstruyendo y redesplegando..."
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d --build

echo ""
echo "✅ Actualización completada!"
echo ""
echo "📊 Ver logs:"
echo "   docker-compose -f docker-compose.prod.yml logs -f api"
echo ""
