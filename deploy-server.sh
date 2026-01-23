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

# Verificar puerto antes de desplegar
echo ""
echo "🔍 Verificando puerto configurado..."

API_PORT=$(grep "^API_PORT" .env.production | cut -d '=' -f2 | tr -d ' ' | tr -d '"' | tr -d "'")
API_PORT=${API_PORT:-8080}

echo "   Puerto configurado: $API_PORT"

# Verificar si el puerto está en uso
if command -v netstat &> /dev/null; then
    if netstat -tuln 2>/dev/null | grep -q ":$API_PORT "; then
        echo ""
        echo "⚠️  ADVERTENCIA: El puerto $API_PORT está en uso!"
        echo ""
        echo "📋 Opciones:"
        echo "   1. Usar otro puerto disponible"
        echo "   2. Detener el servicio que usa el puerto $API_PORT"
        echo ""
        
        # Intentar encontrar puerto disponible
        if [ -f "verificar-puerto.sh" ]; then
            echo "🔍 Buscando puerto disponible..."
            chmod +x verificar-puerto.sh
            ./verificar-puerto.sh $API_PORT
            echo ""
            read -p "¿Deseas continuar de todas formas? (s/n): " continue_anyway
            if [ "$continue_anyway" != "s" ]; then
                echo "❌ Despliegue cancelado. Actualiza API_PORT en .env.production y vuelve a intentar."
                exit 1
            fi
        else
            read -p "¿Deseas continuar de todas formas? (s/n): " continue_anyway
            if [ "$continue_anyway" != "s" ]; then
                echo "❌ Despliegue cancelado. Actualiza API_PORT en .env.production y vuelve a intentar."
                exit 1
            fi
        fi
    else
        echo "✅ Puerto $API_PORT disponible"
    fi
elif command -v ss &> /dev/null; then
    if ss -tuln 2>/dev/null | grep -q ":$API_PORT "; then
        echo ""
        echo "⚠️  ADVERTENCIA: El puerto $API_PORT está en uso!"
        echo "   Actualiza API_PORT en .env.production y vuelve a intentar."
        exit 1
    else
        echo "✅ Puerto $API_PORT disponible"
    fi
fi

# Construir y desplegar
echo ""
echo "🔨 Construyendo imágenes..."
docker-compose -f docker-compose.prod.yml --env-file .env.production build --no-cache

if [ $? -ne 0 ]; then
    echo "❌ Error al construir las imágenes"
    exit 1
fi

echo ""
echo "🚀 Iniciando servicios..."
if ! docker-compose -f docker-compose.prod.yml --env-file .env.production up -d 2>&1 | tee /tmp/docker-up.log; then
    echo ""
    echo "❌ Error al iniciar servicios"
    echo ""
    
    # Verificar si el error es por puerto ocupado
    if grep -q "port is already allocated\|bind: address already in use\|port.*already in use" /tmp/docker-up.log; then
        echo "🔴 ERROR: Puerto $API_PORT está en uso!"
        echo ""
        echo "📋 Solución:"
        echo ""
        
        # Intentar encontrar puerto disponible
        if [ -f "verificar-puerto.sh" ]; then
            echo "🔍 Buscando puerto disponible..."
            chmod +x verificar-puerto.sh
            AVAILABLE_PORT=$(./verificar-puerto.sh $API_PORT 2>&1 | grep "Puerto disponible encontrado" | grep -oE '[0-9]+' | head -1)
            
            if [ -n "$AVAILABLE_PORT" ]; then
                echo ""
                echo "✅ Puerto disponible encontrado: $AVAILABLE_PORT"
                echo ""
                echo "📝 Pasos para solucionar:"
                echo "   1. Edita .env.production:"
                echo "      nano .env.production"
                echo ""
                echo "   2. Cambia API_PORT=$API_PORT a API_PORT=$AVAILABLE_PORT"
                echo ""
                echo "   3. Actualiza firewall:"
                echo "      sudo ufw allow $AVAILABLE_PORT/tcp"
                echo "      sudo ufw reload"
                echo ""
                echo "   4. Vuelve a ejecutar:"
                echo "      ./deploy-server.sh"
                echo ""
                echo "   5. Actualiza app Flutter con puerto $AVAILABLE_PORT"
            else
                echo "   No se pudo encontrar puerto automáticamente"
                echo ""
                echo "   Pasos manuales:"
                echo "   1. Ejecuta: ./verificar-puerto.sh"
                echo "   2. Edita .env.production con el puerto disponible"
                echo "   3. Vuelve a ejecutar: ./deploy-server.sh"
            fi
        else
            echo "   Pasos para solucionar:"
            echo "   1. Ver qué usa el puerto:"
            echo "      sudo netstat -tulpn | grep :$API_PORT"
            echo ""
            echo "   2. Encuentra puerto disponible:"
            echo "      sudo netstat -tulpn | grep LISTEN"
            echo ""
            echo "   3. Edita .env.production:"
            echo "      nano .env.production"
            echo "      Cambia API_PORT=$API_PORT a otro puerto (ej: 8081, 8082)"
            echo ""
            echo "   4. Vuelve a ejecutar:"
            echo "      ./deploy-server.sh"
        fi
    else
        echo "🔍 Posibles causas:"
        echo "   1. Error en la configuración"
        echo "   2. Problema con Docker"
        echo "   3. Error en .env.production"
        echo ""
        echo "📋 Soluciones:"
        echo "   1. Ver logs: docker-compose -f docker-compose.prod.yml logs api"
        echo "   2. Verificar .env.production"
        echo "   3. Verificar Docker: docker ps"
    fi
    
    echo ""
    echo "📄 Ver logs completos:"
    echo "   cat /tmp/docker-up.log"
    echo ""
    exit 1
fi

echo ""
echo "⏳ Esperando a que los servicios estén listos..."
sleep 15

# Verificar health check
echo ""
echo "🔍 Verificando health check..."

# Obtener puerto configurado
API_PORT=$(grep "^API_PORT" .env.production | cut -d '=' -f2 | tr -d ' ' | tr -d '"' | tr -d "'")
API_PORT=${API_PORT:-8080}

echo "   Usando puerto: $API_PORT"

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
