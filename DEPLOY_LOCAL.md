# 🐳 Despliegue Local con Docker

Guía rápida para desplegar Qenti API localmente con Docker.

## 📋 Prerequisitos

1. **Docker Desktop** instalado y corriendo
   - Descargar: https://www.docker.com/products/docker-desktop
   - Verificar: `docker --version`

2. **Archivo .env** configurado (ya creado automáticamente)

## 🚀 Pasos de Despliegue

### 1. Verificar Docker está corriendo

```powershell
# Verificar Docker
docker ps

# Si da error, inicia Docker Desktop y espera a que esté listo
```

### 2. Iniciar servicios

```powershell
# Solo PostgreSQL (recomendado para desarrollo)
docker-compose up -d postgres

# O todos los servicios (PostgreSQL + API)
docker-compose up -d
```

### 3. Verificar que funciona

```powershell
# Ver logs
docker-compose logs -f

# Verificar health check
curl http://localhost:8080/health

# O en PowerShell
Invoke-WebRequest -Uri http://localhost:8080/health
```

### 4. Ejecutar la API localmente (sin Docker)

Si prefieres ejecutar la API fuera de Docker pero usar PostgreSQL en Docker:

```powershell
# 1. Iniciar solo PostgreSQL
docker-compose up -d postgres

# 2. Ejecutar API localmente
go run cmd/server/main.go
```

## 📊 Comandos Útiles

```powershell
# Ver estado de contenedores
docker-compose ps

# Ver logs
docker-compose logs -f api
docker-compose logs -f postgres

# Detener servicios
docker-compose down

# Detener y eliminar volúmenes (⚠️ borra datos)
docker-compose down -v

# Reconstruir imagen
docker-compose build --no-cache api

# Reiniciar un servicio
docker-compose restart postgres
```

## 🔍 Verificar Despliegue

### Health Check

```powershell
Invoke-WebRequest -Uri http://localhost:8080/health
```

**Respuesta esperada:**
```json
{
  "status": "ok",
  "service": "qenti-api"
}
```

### Probar Endpoints

```powershell
# Feed público
Invoke-WebRequest -Uri http://localhost:8080/api/v1/app/feed

# Series
Invoke-WebRequest -Uri http://localhost:8080/api/v1/app/series
```

## 🗄️ Acceder a PostgreSQL

### Desde Docker

```powershell
# Conectar al contenedor
docker exec -it qenti-postgres psql -U postgres -d qenti

# Ver tablas
\dt

# Salir
\q
```

### Desde pgAdmin (puerto 5050)

1. Abre http://localhost:5050
2. Login:
   - Email: `admin@qenti.com`
   - Password: `admin`
3. Agregar servidor:
   - Host: `postgres` (nombre del servicio)
   - Port: `5432`
   - User: `postgres`
   - Password: `postgres`

## ⚠️ Troubleshooting

### Error: "Docker Desktop no está corriendo"

**Solución:**
1. Abre Docker Desktop
2. Espera a que el ícono en la bandeja esté verde
3. Intenta de nuevo

### Error: "Port already in use"

**Solución:**
```powershell
# Ver qué usa el puerto 8080
netstat -ano | findstr :8080

# Matar proceso (reemplaza <PID> con el número)
taskkill /PID <PID> /F

# O cambiar puerto en docker-compose.yml
```

### Error: "Cannot connect to database"

**Solución:**
```powershell
# Verificar que PostgreSQL esté corriendo
docker-compose ps

# Ver logs de PostgreSQL
docker-compose logs postgres

# Reiniciar PostgreSQL
docker-compose restart postgres
```

### Error: "Migration failed"

**Solución:**
```powershell
# Limpiar y reiniciar
docker-compose down -v
docker-compose up -d postgres

# Esperar 5 segundos y ejecutar API
go run cmd/server/main.go
```

## 🔧 Configuración Adicional

### Variables de Entorno

Edita `.env` para configurar:
- `JWT_SECRET` - Generar uno seguro
- `BUNNY_*` - Credenciales de Bunny.net
- `FIREBASE_*` - Credenciales de Firebase

### Firebase Credentials

Si tienes `firebase-credentials.json`, colócalo en la raíz del proyecto.

## ✅ Checklist

- [ ] Docker Desktop corriendo
- [ ] Archivo `.env` configurado
- [ ] PostgreSQL iniciado (`docker-compose ps`)
- [ ] Health check responde OK
- [ ] Logs sin errores críticos

---

**¡Listo para desarrollar! 🚀**

