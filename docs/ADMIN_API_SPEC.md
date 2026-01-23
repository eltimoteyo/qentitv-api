# 📋 Especificación de API Admin - Qenti

Documentación completa de endpoints para el panel de administración.

## 🔐 Autenticación

Todos los endpoints requieren:
- **Header:** `Authorization: Bearer <JWT_TOKEN>`
- **Rol:** `admin` (verificado en token JWT y base de datos)

---

## 📊 Dashboard & Analytics

### GET /api/v1/admin/dashboard

Obtiene métricas y datos para el dashboard principal.

**Response 200:**
```json
{
  "metrics": {
    "total_series": 15,
    "total_episodes": 120,
    "total_users": 1250,
    "active_users_7d": 450,
    "active_users_30d": 980,
    "premium_users": 85,
    "total_revenue_30d": 0.0
  },
  "charts": {
    "retention_by_episode": [
      {
        "episode_number": 1,
        "completion_rate": 0.85
      },
      {
        "episode_number": 2,
        "completion_rate": 0.72
      }
    ],
    "top_dramas": [
      {
        "episode_id": "uuid",
        "episode_title": "Episode Title",
        "series_id": "uuid",
        "series_title": "Series Title",
        "view_count": 1250
      }
    ],
    "revenue": []
  }
}
```

**Campos:**
- `metrics.total_series`: Series activas
- `metrics.total_episodes`: Total de episodios
- `metrics.total_users`: Total de usuarios registrados
- `metrics.active_users_7d`: Usuarios activos últimos 7 días
- `metrics.active_users_30d`: Usuarios activos últimos 30 días
- `metrics.premium_users`: Usuarios con suscripción premium
- `metrics.total_revenue_30d`: Ingresos últimos 30 días (placeholder)
- `charts.retention_by_episode`: Tasa de completación por episodio
- `charts.top_dramas`: Top dramas más vistos (últimos 30 días)

---

## 🎬 Series Management

### GET /api/v1/admin/series

Lista todas las series (incluye inactivas).

**Query Parameters:**
- Ninguno

**Response 200:**
```json
{
  "series": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Serie Title",
      "description": "Description",
      "horizontal_poster": "https://...",
      "vertical_poster": "https://...",
      "is_active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-20T14:20:00Z"
    }
  ]
}
```

---

### GET /api/v1/admin/series/:id

Obtiene una serie específica por ID.

**Path Parameters:**
- `id` (UUID): ID de la serie

**Response 200:**
```json
{
  "series": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Serie Title",
    "description": "Description",
    "horizontal_poster": "https://...",
    "vertical_poster": "https://...",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T14:20:00Z"
  }
}
```

**Errors:**
- `400`: Invalid series ID
- `404`: Series not found

---

### POST /api/v1/admin/series

Crea una nueva serie.

**Request Body:**
```json
{
  "title": "Serie Title",
  "description": "Description of the series",
  "horizontal_poster": "https://example.com/poster-h.jpg",
  "vertical_poster": "https://example.com/poster-v.jpg",
  "is_active": true
}
```

**Campos requeridos:**
- `title` (string, required): Título de la serie

**Campos opcionales:**
- `description` (string): Descripción
- `horizontal_poster` (string): URL del poster horizontal
- `vertical_poster` (string): URL del poster vertical
- `is_active` (boolean, default: true): Si la serie está activa

**Response 201:**
```json
{
  "series": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Serie Title",
    "description": "Description",
    "horizontal_poster": "https://...",
    "vertical_poster": "https://...",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors:**
- `400`: Invalid request body
- `500`: Failed to create series

---

### PUT /api/v1/admin/series/:id

Actualiza una serie existente.

**Path Parameters:**
- `id` (UUID): ID de la serie

**Request Body:**
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "horizontal_poster": "https://...",
  "vertical_poster": "https://...",
  "is_active": false
}
```

**Nota:** Todos los campos son opcionales. Solo se actualizan los campos enviados.

**Response 200:**
```json
{
  "series": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Updated Title",
    "description": "Updated description",
    "horizontal_poster": "https://...",
    "vertical_poster": "https://...",
    "is_active": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T14:20:00Z"
  }
}
```

**Errors:**
- `400`: Invalid series ID or request body
- `404`: Series not found
- `500`: Failed to update series

---

### DELETE /api/v1/admin/series/:id

Elimina una serie (soft delete - marca `is_active=false`).

**Path Parameters:**
- `id` (UUID): ID de la serie

**Response 200:**
```json
{
  "message": "Series deleted successfully"
}
```

**Errors:**
- `400`: Invalid series ID
- `500`: Failed to delete series

---

## 📺 Episodes Management

### GET /api/v1/admin/episodes

Lista todos los episodios.

**Query Parameters:**
- `series_id` (UUID, optional): Filtrar por serie específica

**Ejemplo:**
```
GET /api/v1/admin/episodes?series_id=550e8400-e29b-41d4-a716-446655440000
```

**Response 200:**
```json
{
  "episodes": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "series_id": "550e8400-e29b-41d4-a716-446655440000",
      "episode_number": 1,
      "title": "Episode 1: Beginning",
      "video_id_bunny": "bunny-video-id-123",
      "duration": 180,
      "is_free": true,
      "price_coins": 0,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-20T14:20:00Z"
    }
  ]
}
```

---

### GET /api/v1/admin/episodes/:id

Obtiene un episodio específico por ID.

**Path Parameters:**
- `id` (UUID): ID del episodio

**Response 200:**
```json
{
  "episode": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "series_id": "550e8400-e29b-41d4-a716-446655440000",
    "episode_number": 1,
    "title": "Episode 1: Beginning",
    "video_id_bunny": "bunny-video-id-123",
    "duration": 180,
    "is_free": true,
    "price_coins": 0,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T14:20:00Z"
  }
}
```

**Errors:**
- `400`: Invalid episode ID
- `404`: Episode not found

---

### POST /api/v1/admin/episodes

Crea un nuevo episodio.

**Request Body:**
```json
{
  "series_id": "550e8400-e29b-41d4-a716-446655440000",
  "episode_number": 1,
  "title": "Episode 1: Beginning",
  "duration": 180,
  "is_free": true,
  "price_coins": 10
}
```

**Campos requeridos:**
- `series_id` (UUID, required): ID de la serie
- `episode_number` (integer, required): Número del episodio
- `title` (string, required): Título del episodio

**Campos opcionales:**
- `duration` (integer): Duración en segundos
- `is_free` (boolean, default: false): Si el episodio es gratis
- `price_coins` (integer, default: 0): Precio en monedas si no es gratis

**Response 201:**
```json
{
  "episode": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "series_id": "550e8400-e29b-41d4-a716-446655440000",
    "episode_number": 1,
    "title": "Episode 1: Beginning",
    "video_id_bunny": "",
    "duration": 180,
    "is_free": true,
    "price_coins": 0,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors:**
- `400`: Invalid request body
- `500`: Failed to create episode

---

### PUT /api/v1/admin/episodes/:id

Actualiza un episodio existente.

**Path Parameters:**
- `id` (UUID): ID del episodio

**Request Body:**
```json
{
  "title": "Updated Episode Title",
  "duration": 200,
  "is_free": false,
  "price_coins": 15
}
```

**Nota:** Todos los campos son opcionales. Solo se actualizan los campos enviados.

**Response 200:**
```json
{
  "episode": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "series_id": "550e8400-e29b-41d4-a716-446655440000",
    "episode_number": 1,
    "title": "Updated Episode Title",
    "video_id_bunny": "bunny-video-id-123",
    "duration": 200,
    "is_free": false,
    "price_coins": 15,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T14:20:00Z"
  }
}
```

**Errors:**
- `400`: Invalid episode ID or request body
- `404`: Episode not found
- `500`: Failed to update episode

---

### DELETE /api/v1/admin/episodes/:id

Elimina un episodio.

**Path Parameters:**
- `id` (UUID): ID del episodio

**Response 200:**
```json
{
  "message": "Episode deleted successfully"
}
```

**Errors:**
- `400`: Invalid episode ID
- `500`: Failed to delete episode

---

## 🎥 Video Upload Flow

### POST /api/v1/admin/episodes/:id/upload-url

Genera una URL presignada para subir video directamente a Bunny.net.

**Path Parameters:**
- `id` (UUID): ID del episodio

**Response 200:**
```json
{
  "upload_url": "https://video.bunnycdn.com/library/12345/videos/abc-def-ghi",
  "episode_id": "660e8400-e29b-41d4-a716-446655440000"
}
```

**Flujo de uso:**
1. Llamar este endpoint para obtener `upload_url`
2. Subir el video directamente a `upload_url` usando PUT request
3. Obtener el `video_id` de la respuesta de Bunny.net
4. Llamar `/admin/episodes/:id/complete` con el `video_id_bunny`

**Errors:**
- `400`: Invalid episode ID
- `404`: Episode not found
- `500`: Failed to generate upload URL

---

### POST /api/v1/admin/episodes/:id/complete

Marca el upload como completado y guarda el `video_id_bunny`.

**Path Parameters:**
- `id` (UUID): ID del episodio

**Request Body:**
```json
{
  "video_id_bunny": "bunny-video-id-123"
}
```

**Campos requeridos:**
- `video_id_bunny` (string, required): ID del video en Bunny.net

**Response 200:**
```json
{
  "message": "Upload completed successfully"
}
```

**Response 200 (con warning):**
```json
{
  "message": "Episode updated successfully, but re-encoding may have failed",
  "warning": "error message"
}
```

**Errors:**
- `400`: Invalid episode ID or request body
- `404`: Episode not found
- `500`: Failed to update episode video ID

---

## 👥 Users Management

### GET /api/v1/admin/users

Lista usuarios con paginación.

**Query Parameters:**
- `page` (integer, default: 1): Número de página
- `limit` (integer, default: 20, max: 100): Elementos por página

**Ejemplo:**
```
GET /api/v1/admin/users?page=1&limit=20
```

**Response 200:**
```json
{
  "users": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "coin_balance": 150,
      "is_premium": false,
      "created_at": "2024-01-10T08:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1250,
    "pages": 63
  }
}
```

---

### GET /api/v1/admin/users/:id

Obtiene el detalle completo de un usuario con historial.

**Path Parameters:**
- `id` (UUID): ID del usuario

**Response 200:**
```json
{
  "user": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "firebase_uid": "firebase-uid-123",
    "coin_balance": 150,
    "is_premium": false,
    "created_at": "2024-01-10T08:00:00Z",
    "updated_at": "2024-01-15T12:00:00Z"
  },
  "history": {
    "unlocked_episodes": 25,
    "completed_episodes": 18,
    "total_watch_time": 5400,
    "recent_transactions": [
      {
        "id": "880e8400-e29b-41d4-a716-446655440000",
        "user_id": "770e8400-e29b-41d4-a716-446655440000",
        "type": "coin_purchase",
        "amount": 100,
        "currency": "coins",
        "episode_id": null,
        "created_at": "2024-01-15T10:00:00Z"
      }
    ]
  },
  "bans": {
    "is_banned": false,
    "active_ban": null,
    "all_bans": []
  }
}
```

**Campos:**
- `history.unlocked_episodes`: Número de episodios desbloqueados
- `history.completed_episodes`: Número de episodios completados
- `history.total_watch_time`: Tiempo total de visionado en segundos
- `history.recent_transactions`: Últimas 10 transacciones
- `bans.is_banned`: Si el usuario está actualmente baneado
- `bans.active_ban`: Información del ban activo (si existe)
- `bans.all_bans`: Historial completo de bans

**Errors:**
- `400`: Invalid user ID
- `404`: User not found

---

### PUT /api/v1/admin/users/:id/ban

Banea un usuario.

**Path Parameters:**
- `id` (UUID): ID del usuario

**Request Body:**
```json
{
  "reason": "Violación de términos de servicio",
  "expires_at": "2024-02-15T00:00:00Z"
}
```

**Campos requeridos:**
- `reason` (string, required): Razón del ban

**Campos opcionales:**
- `expires_at` (ISO 8601 datetime, optional): Fecha de expiración del ban. Si es `null` o no se envía, el ban es permanente.

**Response 200:**
```json
{
  "message": "User banned successfully",
  "user_id": "770e8400-e29b-41d4-a716-446655440000",
  "ban_id": "990e8400-e29b-41d4-a716-446655440000",
  "expires_at": "2024-02-15T00:00:00Z"
}
```

**Errors:**
- `400`: Invalid user ID, user already banned, or invalid request body
- `404`: User not found
- `500`: Failed to ban user

---

### PUT /api/v1/admin/users/:id/coins

Regala monedas a un usuario manualmente.

**Path Parameters:**
- `id` (UUID): ID del usuario

**Request Body:**
```json
{
  "amount": 100
}
```

**Campos requeridos:**
- `amount` (integer, required, min: 1): Cantidad de monedas a regalar

**Response 200:**
```json
{
  "message": "Coins gifted successfully",
  "user_id": "770e8400-e29b-41d4-a716-446655440000",
  "amount": 100,
  "new_balance": 250,
  "gifted_by": "aa0e8400-e29b-41d4-a716-446655440000"
}
```

**Campos:**
- `new_balance`: Nuevo balance del usuario
- `gifted_by`: ID del administrador que realizó el regalo

**Errors:**
- `400`: Invalid user ID or request body
- `404`: User not found
- `500`: Failed to update coin balance

---

## 🔔 Webhooks

### POST /api/v1/webhooks/revenuecat

Endpoint para recibir webhooks de RevenueCat (no requiere autenticación JWT, usa firma propia).

**Headers:**
- `Authorization`: Firma del webhook de RevenueCat

**Request Body:**
```json
{
  "event": {
    "type": "INITIAL_PURCHASE",
    "app_user_id": "firebase-uid-123",
    "product_id": "premium_monthly"
  }
}
```

**Eventos procesados:**
- `INITIAL_PURCHASE`: Usuario compró suscripción → `is_premium=true`
- `RENEWAL`: Usuario renovó suscripción → `is_premium=true`
- `CANCELLATION`: Usuario canceló → `is_premium=false`
- `EXPIRATION`: Suscripción expiró → `is_premium=false`

**Response 200:**
```json
{
  "message": "Webhook processed successfully"
}
```

**Errors:**
- `400`: Failed to process webhook
- `401`: Missing or invalid authorization signature
- `500`: Failed to update premium status

---

## 📝 Tipos de Datos

### UUID
Formato: `550e8400-e29b-41d4-a716-446655440000`

### Timestamp
Formato ISO 8601: `2024-01-15T10:30:00Z`

### Boolean
`true` o `false`

### Integer
Números enteros

### String
Cadenas de texto

---

## ⚠️ Códigos de Error Comunes

- `400 Bad Request`: Request inválido (campos faltantes, formato incorrecto)
- `401 Unauthorized`: No autenticado o token inválido
- `403 Forbidden`: No tiene permisos de administrador
- `404 Not Found`: Recurso no encontrado
- `429 Too Many Requests`: Rate limit excedido
- `500 Internal Server Error`: Error del servidor

---

## 🔄 Rate Limiting

Los endpoints de admin tienen rate limiting más generoso:
- **Límite:** 10 requests por segundo
- **Burst:** 20 requests

---

## 📌 Notas Importantes

1. **Autenticación:** Todos los endpoints (excepto webhooks) requieren JWT con rol `admin`
2. **UUIDs:** Todos los IDs son UUIDs v4
3. **Timestamps:** Todos los timestamps están en formato ISO 8601 UTC
4. **Paginación:** Los endpoints de listado soportan paginación
5. **Soft Delete:** Las series se eliminan con soft delete (marcan `is_active=false`)
6. **Video Upload:** El flujo de subida de video requiere 2 pasos (upload-url → complete)

---

**Última actualización:** 2024-01-21

