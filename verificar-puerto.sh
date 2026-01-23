#!/bin/bash

# Script para verificar qué puertos están disponibles
# Uso: ./verificar-puerto.sh [puerto_inicial]

PORT=${1:-8080}
MAX_PORT=8090

echo "🔍 Verificando puertos disponibles..."
echo ""

# Función para verificar si un puerto está en uso
check_port() {
    if command -v netstat &> /dev/null; then
        netstat -tuln | grep -q ":$1 "
    elif command -v ss &> /dev/null; then
        ss -tuln | grep -q ":$1 "
    elif command -v lsof &> /dev/null; then
        lsof -i :$1 &> /dev/null
    else
        # Fallback: intentar conectar
        timeout 1 bash -c "echo > /dev/tcp/localhost/$1" 2>/dev/null
    fi
}

# Buscar puerto disponible
FOUND_PORT=""
for ((p=$PORT; p<=$MAX_PORT; p++)); do
    if ! check_port $p; then
        FOUND_PORT=$p
        break
    else
        echo "⚠️  Puerto $p está en uso"
    fi
done

if [ -z "$FOUND_PORT" ]; then
    echo ""
    echo "❌ No se encontró un puerto disponible entre $PORT y $MAX_PORT"
    echo "   Prueba con un rango diferente o libera un puerto"
    exit 1
fi

echo ""
echo "✅ Puerto disponible encontrado: $FOUND_PORT"
echo ""
echo "📝 Para usar este puerto:"
echo "   1. Edita .env.production y agrega:"
echo "      API_PORT=$FOUND_PORT"
echo ""
echo "   2. O exporta la variable antes de desplegar:"
echo "      export API_PORT=$FOUND_PORT"
echo "      ./deploy-server.sh"
echo ""
