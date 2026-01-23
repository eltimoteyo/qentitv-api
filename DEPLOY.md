# 🚀 Guía de Despliegue - Qenti API

Esta guía te ayudará a desplegar la API de Qenti en diferentes entornos.

## 📋 Tabla de Contenidos

1. [Requisitos Previos](#requisitos-previos)
2. [Configuración de Servicios Externos](#configuración-de-servicios-externos)
3. [Variables de Entorno](#variables-de-entorno)
4. [Despliegue Local](#despliegue-local)
5. [Despliegue con Docker](#despliegue-con-docker)
6. [Despliegue en Cloud](#despliegue-en-cloud)
7. [Verificación Post-Despliegue](#verificación-post-despliegue)
8. [Troubleshooting](#troubleshooting)

---

## 🔧 Requisitos Previos

### Software Necesario

- **Go 1.21+** - [Descargar](https://golang.org/dl/)
- **PostgreSQL 15+** - [Descargar](https://www.postgresql.org/download/)
- **Docker & Docker Compose** (opcional, para despliegue con contenedores) - [Descargar](https://www.docker.com/get-started)
- **Git** - Para clonar el repositorio

### Cuentas de Servicios Externos

1. **Firebase** - Para autenticación
2. **Bunny.net** - Para streaming de video
3. **RevenueCat** - Para pagos y suscripciones (opcional para MVP)

---

## 🌐 Configuración de Servicios Externos

### 1. Firebase Authentication

1. Ve a [Firebase Console](https://console.firebase.google.com/)
2. Crea un nuevo proyecto o usa uno existente
3. Habilita **Authentication** → **Sign-in method** → **Email/Password**
4. Ve a **Project Settings** → **Service accounts**
5. Genera una nueva clave privada (JSON)
6. Guarda el archivo como `firebase-credentials.json` en la raíz del proyecto

**Variables necesarias:**
- `FIREBASE_PROJECT_ID`: ID del proyecto Firebase
- `FIREBASE_CREDENTIALS_PATH`: Ruta al archivo JSON (ej: `./firebase-credentials.json`)

### 2. Bunny.net

1. Crea una cuenta en [Bunny.net](https://bunny.net/)
2. Crea una **Stream Library**
3. Obtén las siguientes credenciales:
   - **Library ID**: ID de tu Stream Library
   - **API Key**: Clave API de tu cuenta
   - **CDN Hostname**: Hostname de tu CDN (ej: `qenti.b-cdn.net`)
   - **Security Key**: Clave de seguridad para firmar URLs (en configuración de la library)

**Variables necesarias:**
- `BUNNY_STREAM_LIBRARY_ID`
- `BUNNY_STREAM_API_KEY`
- `BUNNY_CDN_HOSTNAME`
- `BUNNY_SECURITY_KEY`

### 3. RevenueCat (Opcional)

1. Crea una cuenta en [RevenueCat](https://www.revenuecat.com/)
2. Crea un nuevo proyecto
3. Configura productos y suscripciones
4. Obtén la **API Key** y configura el **Webhook Secret**

**Variables necesarias:**
- `REVENUECAT_API_KEY`
- `REVENUECAT_WEBHOOK_SECRET`

---

## 🔐 Variables de Entorno

Crea un archivo `.env` en la raíz del proyecto con las siguientes variables:

```bash
# ============================================
# APP CONFIGURATION
# ============================================
APP_ENV=production
PORT=8080

# ============================================
# DATABASE (PostgreSQL)
# ============================================
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=tu_password_seguro
DB_NAME=qenti
DB_SSLMODE=disable

# ============================================
# JWT AUTHENTICATION
# ============================================
# Genera una clave secreta segura (ej: openssl rand -base64 32)
JWT_SECRET=tu_jwt_secret_muy_seguro_aqui

# ============================================
# BUNNY.NET (Video Streaming)
# ============================================
BUNNY_STREAM_LIBRARY_ID=tu_library_id
BUNNY_STREAM_API_KEY=tu_api_key
BUNNY_CDN_HOSTNAME=tu_cdn_hostname.b-cdn.net
BUNNY_SECURITY_KEY=tu_security_key

# ============================================
# FIREBASE AUTHENTICATION
# ============================================
FIREBASE_PROJECT_ID=tu-firebase-project-id
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json

# ============================================
# REVENUECAT (Payments)
# ============================================
REVENUECAT_API_KEY=tu_revenuecat_api_key
REVENUECAT_WEBHOOK_SECRET=tu_webhook_secret
```

### Generar JWT Secret

```bash
# Linux/Mac
openssl rand -base64 32

# Windows (PowerShell)
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
```

---

## 💻 Despliegue Local

### Opción 1: Sin Docker

1. **Clonar el repositorio:**
```bash
git clone <tu-repo>
cd QENTITV-API
```

2. **Instalar dependencias:**
```bash
go mod download
go mod tidy
```

3. **Configurar PostgreSQL:**
```bash
# Crear base de datos
createdb qenti

# O usando psql
psql -U postgres
CREATE DATABASE qenti;
\q
```

4. **Configurar variables de entorno:**
```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar .env con tus valores
nano .env  # o usar tu editor preferido
```

5. **Ejecutar migraciones:**
```bash
# Las migraciones se ejecutan automáticamente al iniciar la app
go run cmd/server/main.go
```

6. **Verificar que funciona:**
```bash
curl http://localhost:8080/health
```

### Opción 2: Con Docker Compose (Desarrollo)

1. **Iniciar PostgreSQL:**
```bash
docker-compose up -d postgres
```

2. **Configurar .env** (ver sección anterior)

3. **Ejecutar la API:**
```bash
go run cmd/server/main.go
```

---

## 🐳 Despliegue con Docker

### Desarrollo Local

1. **Construir y ejecutar:**
```bash
docker-compose up -d
```

2. **Ver logs:**
```bash
docker-compose logs -f api
```

3. **Detener:**
```bash
docker-compose down
```

### Producción

1. **Preparar archivo de producción:**
```bash
# Crear .env.production con tus valores
cp .env .env.production
```

2. **Construir imagen:**
```bash
docker build -t qenti-api:latest .
```

3. **Ejecutar con docker-compose.prod.yml:**
```bash
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
```

4. **Verificar:**
```bash
docker-compose -f docker-compose.prod.yml logs -f api
```

---

## ☁️ Despliegue en Cloud

### Opción 1: Railway

1. **Instalar Railway CLI:**
```bash
npm i -g @railway/cli
railway login
```

2. **Inicializar proyecto:**
```bash
railway init
railway link
```

3. **Configurar variables:**
```bash
railway variables set JWT_SECRET=tu_secret
railway variables set DB_HOST=tu_host
# ... (configurar todas las variables)
```

4. **Desplegar:**
```bash
railway up
```

### Opción 2: Heroku

1. **Instalar Heroku CLI:**
```bash
# Ver: https://devcenter.heroku.com/articles/heroku-cli
```

2. **Crear app:**
```bash
heroku create qenti-api
```

3. **Agregar PostgreSQL:**
```bash
heroku addons:create heroku-postgresql:hobby-dev
```

4. **Configurar variables:**
```bash
heroku config:set JWT_SECRET=tu_secret
heroku config:set FIREBASE_PROJECT_ID=tu_project_id
# ... (configurar todas las variables)
```

5. **Desplegar:**
```bash
git push heroku main
```

### Opción 3: Google Cloud Run

1. **Instalar gcloud CLI:**
```bash
# Ver: https://cloud.google.com/sdk/docs/install
```

2. **Configurar proyecto:**
```bash
gcloud config set project tu-project-id
```

3. **Construir y desplegar:**
```bash
# Construir imagen
gcloud builds submit --tag gcr.io/tu-project-id/qenti-api

# Desplegar
gcloud run deploy qenti-api \
  --image gcr.io/tu-project-id/qenti-api \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars="JWT_SECRET=tu_secret,DB_HOST=tu_host,..."
```

### Opción 4: AWS (EC2 + Docker)

1. **Conectar a EC2:**
```bash
ssh -i tu-key.pem ubuntu@tu-ec2-ip
```

2. **Instalar Docker:**
```bash
sudo apt update
sudo apt install docker.io docker-compose -y
sudo usermod -aG docker $USER
```

3. **Clonar y configurar:**
```bash
git clone <tu-repo>
cd QENTITV-API
# Configurar .env
nano .env
```

4. **Desplegar:**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

5. **Configurar Nginx (opcional, como reverse proxy):**
```bash
sudo apt install nginx
# Configurar /etc/nginx/sites-available/qenti-api
sudo systemctl restart nginx
```

### Opción 5: DigitalOcean App Platform

1. **Conectar repositorio** en [DigitalOcean](https://cloud.digitalocean.com/apps)

2. **Configurar build:**
   - Build Command: `go build -o qenti-api cmd/server/main.go`
   - Run Command: `./qenti-api`

3. **Agregar PostgreSQL** desde el panel

4. **Configurar variables de entorno** en el panel

5. **Desplegar**

---

## ✅ Verificación Post-Despliegue

### 1. Health Check

```bash
curl http://localhost:8080/health
```

**Respuesta esperada:**
```json
{
  "status": "ok",
  "service": "qenti-api"
}
```

### 2. Verificar Base de Datos

```bash
# Conectar a PostgreSQL
psql -h localhost -U postgres -d qenti

# Verificar tablas
\dt

# Deberías ver: users, series, episodes, unlocks, transactions, views, bans, etc.
```

### 3. Probar Endpoints

```bash
# Feed público
curl http://localhost:8080/api/v1/app/feed

# Series
curl http://localhost:8080/api/v1/app/series
```

### 4. Verificar Logs

```bash
# Docker
docker-compose logs -f api

# Local
# Los logs aparecen en la consola donde ejecutaste la app
```

---

## 🔍 Troubleshooting

### Error: "Failed to connect to database"

**Solución:**
- Verificar que PostgreSQL esté corriendo
- Verificar credenciales en `.env`
- Verificar que la base de datos exista

```bash
# Verificar conexión
psql -h localhost -U postgres -d qenti
```

### Error: "Firebase Admin SDK not initialized"

**Solución:**
- Verificar que `FIREBASE_CREDENTIALS_PATH` apunte al archivo correcto
- Verificar que el archivo JSON sea válido
- Verificar permisos del archivo

### Error: "Bunny.net API error"

**Solución:**
- Verificar que las credenciales de Bunny.net sean correctas
- Verificar que la Stream Library esté activa
- Verificar conectividad a internet

### Error: "JWT_SECRET is not set"

**Solución:**
- Generar un nuevo JWT_SECRET (ver sección de variables de entorno)
- Asegurarse de que esté en el archivo `.env`

### Error: "Port already in use"

**Solución:**
```bash
# Cambiar puerto en .env
PORT=8081

# O matar el proceso que usa el puerto
# Linux/Mac
lsof -ti:8080 | xargs kill -9

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Error: "Migration failed"

**Solución:**
- Verificar que PostgreSQL tenga permisos para crear tablas
- Verificar que la base de datos esté vacía o usar una nueva
- Revisar logs para ver el error específico

---

## 📊 Monitoreo Recomendado

### Logs

- **Desarrollo:** Logs en consola
- **Producción:** Considerar usar servicios como:
  - **Datadog**
  - **Sentry** (para errores)
  - **CloudWatch** (AWS)
  - **Stackdriver** (GCP)

### Métricas

- **Health checks:** `/health`
- **Rate limiting:** Monitorear respuestas 429
- **Database:** Monitorear conexiones y queries lentas
- **API:** Tiempo de respuesta y throughput

### Alertas

Configurar alertas para:
- Disponibilidad del servicio (< 99%)
- Errores 5xx > 1%
- Tiempo de respuesta > 1s
- Uso de base de datos > 80%

---

## 🔒 Seguridad en Producción

1. **HTTPS:** Usar certificados SSL/TLS (Let's Encrypt, Cloudflare)
2. **Secrets:** No commitear `.env` o `firebase-credentials.json`
3. **Rate Limiting:** Ya implementado, verificar límites
4. **CORS:** Configurar orígenes permitidos en producción
5. **Database:** Usar SSL para conexiones a PostgreSQL
6. **Firewall:** Restringir acceso a la base de datos
7. **Backups:** Configurar backups automáticos de PostgreSQL

---

## 📝 Checklist de Despliegue

- [ ] Variables de entorno configuradas
- [ ] Firebase credentials configuradas
- [ ] Bunny.net configurado
- [ ] PostgreSQL corriendo y accesible
- [ ] Migraciones ejecutadas correctamente
- [ ] Health check responde OK
- [ ] Endpoints públicos funcionando
- [ ] Autenticación funcionando
- [ ] Logs configurados
- [ ] Monitoreo configurado
- [ ] Backups configurados
- [ ] HTTPS configurado
- [ ] Rate limiting activo
- [ ] CORS configurado para producción

---

## 🆘 Soporte

Si encuentras problemas:

1. Revisar logs de la aplicación
2. Verificar variables de entorno
3. Verificar conectividad a servicios externos
4. Revisar esta guía de troubleshooting
5. Abrir un issue en el repositorio

---

**¡Listo para desplegar! 🚀**

