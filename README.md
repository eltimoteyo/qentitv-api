# 🐦 Qenti API

API backend para la plataforma de streaming de micro-dramas **Qenti** (Colibrí en Quechua).

## 📋 Características

- ✅ Autenticación con Firebase y JWT
- ✅ Gestión de series y episodios
- ✅ Sistema de desbloqueo (coins, ads, premium)
- ✅ Streaming de video con Bunny.net
- ✅ Sistema de pagos con RevenueCat
- ✅ Analytics y métricas
- ✅ Feed inteligente con trending y recomendaciones
- ✅ Validación de anuncios
- ✅ Rate limiting
- ✅ Admin panel

## 🚀 Inicio Rápido

### Requisitos

- Go 1.21+
- PostgreSQL 15+
- Docker (opcional)

### Instalación Local

```bash
# 1. Clonar repositorio
git clone <tu-repo>
cd QENTITV-API

# 2. Instalar dependencias
go mod download

# 3. Configurar variables de entorno
cp .env.example .env
# Editar .env con tus valores

# 4. Iniciar PostgreSQL (con Docker)
docker-compose up -d postgres

# 5. Ejecutar la API
go run cmd/server/main.go
```

### Con Docker

```bash
# Desarrollo
docker-compose up -d

# Producción
docker-compose -f docker-compose.prod.yml up -d
```

## 📚 Documentación

- **[DEPLOY.md](./DEPLOY.md)** - Guía completa de despliegue
- **[docs/API.md](./docs/API.md)** - Documentación de endpoints
- **[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - Arquitectura del sistema

## 🔧 Configuración

### Variables de Entorno

Ver [DEPLOY.md](./DEPLOY.md#variables-de-entorno) para la lista completa de variables.

**Variables críticas:**
- `JWT_SECRET` - Clave secreta para JWT
- `DB_*` - Configuración de PostgreSQL
- `FIREBASE_PROJECT_ID` - ID del proyecto Firebase
- `BUNNY_*` - Credenciales de Bunny.net

## 🏗️ Estructura del Proyecto

```
QENTITV-API/
├── api/              # Handlers HTTP
│   └── v1/
│       ├── app/      # Endpoints públicos/autenticados
│       ├── admin/    # Endpoints de administración
│       └── auth/     # Autenticación
├── cmd/
│   └── server/       # Punto de entrada
├── internal/
│   ├── config/       # Configuración
│   ├── database/     # Migraciones y conexión DB
│   ├── middleware/   # Middlewares HTTP
│   ├── pkg/          # Paquetes internos
│   │   ├── auth/     # Autenticación
│   │   ├── bunny/    # Integración Bunny.net
│   │   ├── jwt/      # JWT service
│   │   └── ...
│   └── router/       # Configuración de rutas
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## 📡 Endpoints Principales

### Públicos
- `GET /health` - Health check
- `GET /api/v1/app/feed` - Feed de contenido
- `GET /api/v1/app/series` - Lista de series

### Autenticados
- `POST /api/v1/auth/login` - Login con Firebase
- `GET /api/v1/app/episodes/:id/stream` - Stream de episodio
- `POST /api/v1/app/episodes/:id/unlock` - Desbloquear episodio

### Admin
- `GET /api/v1/admin/dashboard` - Dashboard de analytics
- `POST /api/v1/admin/series` - Crear serie
- `POST /api/v1/admin/episodes` - Crear episodio

Ver [docs/API.md](./docs/API.md) para documentación completa.

## 🧪 Testing

```bash
# Ejecutar tests
make test

# O directamente
go test -v ./...
```

## 🛠️ Comandos Útiles

```bash
# Desarrollo
make run          # Ejecutar API
make dev           # Ejecutar con hot reload (requiere air)
make build         # Compilar binario

# Base de datos
make migrate       # Ejecutar migraciones

# Calidad de código
make fmt           # Formatear código
make lint          # Ejecutar linter
```

## 🔒 Seguridad

- ✅ JWT con expiración
- ✅ Rate limiting
- ✅ Validación de anuncios
- ✅ URLs firmadas para video
- ✅ Autenticación Firebase
- ✅ Roles y permisos

## 📊 Monitoreo

- Health check: `GET /health`
- Logs estructurados
- Métricas de uso (preparado para integración)

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📝 Licencia

Este proyecto es privado y propietario.

## 🆘 Soporte

Para problemas o preguntas:
1. Revisa [DEPLOY.md](./DEPLOY.md) para troubleshooting
2. Abre un issue en el repositorio
3. Contacta al equipo de desarrollo

---

**Desarrollado con ❤️ para Qenti**
