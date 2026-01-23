# Arquitectura de Qenti

## Visión General

Qenti es un monolito modular construido en Go que sirve como backend unificado para:
- **App Móvil**: Consumo de contenido por usuarios finales
- **Panel Admin**: Gestión de contenido y métricas

## Stack Tecnológico

- **Lenguaje**: Go 1.21+
- **Framework Web**: Gin
- **Base de Datos**: PostgreSQL 14+
- **Video Streaming**: Bunny.net
- **Autenticación**: Firebase Auth (JWT)
- **Pagos**: RevenueCat

## Estructura del Proyecto

```
qenti/
├── cmd/
│   └── server/          # Punto de entrada
├── internal/
│   ├── config/          # Configuración
│   ├── database/        # DB connection & migrations
│   ├── middleware/      # Middleware (auth, CORS, logging)
│   ├── router/          # Setup de rutas
│   └── pkg/             # Paquetes por feature (modular monolith)
│       ├── auth/        # Autenticación Firebase
│       ├── bunny/       # Integración Bunny.net
│       ├── episodes/    # Lógica de episodios
│       ├── models/      # Modelos de datos
│       ├── payment/     # Integración RevenueCat
│       ├── series/      # Lógica de series
│       ├── unlocks/    # Lógica de desbloqueos
│       └── users/       # Lógica de usuarios
├── api/
│   └── v1/
│       ├── app/         # Handlers para App móvil
│       └── admin/       # Handlers para Panel Admin
└── migrations/          # Migraciones SQL (futuro)
```

## Patrón: Monolito Modular

El código está organizado por **features** (series, episodes, users) en lugar de por capas técnicas (controllers, services, repositories). Cada feature tiene su propio paquete con:

- **Repository**: Acceso a datos (PostgreSQL)
- **Service**: Lógica de negocio (opcional, si es necesario)
- **Handlers**: Endpoints HTTP (en `api/v1/app` o `api/v1/admin`)

## Flujo de Datos

### 1. Gestión de Video (Direct Upload)

```
Admin Panel → POST /admin/episodes/:id/upload-url
           ← { upload_url }
           
Admin Panel → Upload directo a Bunny.net usando upload_url
           
Admin Panel → POST /admin/episodes/:id/complete-upload
           ← { message: "success" }
```

**Ventajas:**
- El servidor Go nunca recibe el archivo de video
- Escalable (Bunny maneja el upload)
- Menor latencia

### 2. Reproducción Segura

```
App → GET /app/series/:id/playlist
   ← { playlist: [...] }
   
Backend verifica:
  - is_free?
  - Usuario premium?
  - Episodio comprado?
  
Si cumple → Genera URL firmada con expiración
Si no → locked: true, sin video_url
```

### 3. Desbloqueo de Episodios

**Métodos:**
- **COIN**: Usuario paga con monedas
- **AD**: Usuario ve anuncio (futuro)
- **SUB**: Usuario tiene suscripción premium

**Reglas:**
- Episodios `is_free=true` → Acceso libre
- Usuarios `is_premium=true` → Acceso a todo
- Otros → Requieren desbloqueo

## Base de Datos

### Esquema Principal

```
users
├── id (UUID)
├── email
├── firebase_uid (UNIQUE)
├── coin_balance
└── is_premium

series
├── id (UUID)
├── title
├── description
├── horizontal_poster
├── vertical_poster
└── is_active

episodes
├── id (UUID)
├── series_id (FK)
├── episode_number
├── title
├── video_id_bunny
├── duration
├── is_free
└── price_coins

unlocks
├── id (UUID)
├── user_id (FK)
├── episode_id (FK)
├── method (COIN/AD/SUB)
└── unlocked_at
```

## Seguridad

### Autenticación
- Firebase Auth JWT tokens
- Middleware `RequireAuth` valida tokens
- Middleware `RequireAdmin` valida rol admin

### Streaming de Video
- URLs firmadas con expiración (Bunny Token Authentication)
- Evita hotlinking y acceso no autorizado

### Webhooks
- RevenueCat webhooks verificados con HMAC signature
- Previene webhooks falsos

## Performance

### Optimizaciones
- Conexión pool de PostgreSQL
- Índices en campos frecuentemente consultados
- URLs firmadas con expiración corta (1 hora)
- Goroutines para tareas pesadas en segundo plano (futuro)

### Escalabilidad Futura
- Estructura preparada para separar en microservicios
- Cada feature puede convertirse en servicio independiente
- Base de datos puede particionarse por feature si es necesario

## Integraciones Externas

### Bunny.net
- **Stream API**: Gestión de videos
- **Storage API**: Almacenamiento (futuro)
- **CDN**: Distribución de contenido

### Firebase Auth
- Verificación de tokens JWT
- Gestión de usuarios (email, UID)

### RevenueCat
- Webhooks para actualizar `is_premium`
- API para verificar estado de suscripción

## Próximos Pasos

1. Implementar Firebase Admin SDK para verificación real de tokens
2. Implementar sistema de roles/admin en DB o Firebase Custom Claims
3. Agregar analytics reales
4. Implementar sistema de anuncios para desbloqueo
5. Agregar tests unitarios e integración
6. Implementar rate limiting
7. Agregar logging estructurado (Zap)
8. Implementar métricas (Prometheus)

