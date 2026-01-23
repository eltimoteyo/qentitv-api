# Qenti API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Autenticación

La mayoría de los endpoints requieren autenticación mediante Firebase JWT tokens.

**Header requerido:**
```
Authorization: Bearer <firebase-jwt-token>
```

---

## App Endpoints

### GET /app/series
Lista todas las series disponibles (público).

**Response:**
```json
{
  "series": [
    {
      "id": "uuid",
      "title": "Serie Title",
      "description": "Description",
      "horizontal_poster": "url",
      "vertical_poster": "url",
      "is_active": true,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

---

### GET /app/series/:id/playlist
Obtiene la playlist de una serie con URLs firmadas según acceso del usuario.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "playlist": [
    {
      "id": "uuid",
      "episode_number": 1,
      "title": "Episode Title",
      "duration": 120,
      "is_free": false,
      "price_coins": 10,
      "locked": false,
      "video_url": "https://..."
    }
  ]
}
```

**Lógica de desbloqueo:**
- `is_free=true` → Siempre desbloqueado
- Usuario Premium → Todos los episodios desbloqueados
- Episodio comprado → Desbloqueado
- Otros casos → `locked=true` y sin `video_url`

---

### POST /app/episodes/:id/unlock
Desbloquea un episodio usando monedas.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "message": "Episode unlocked successfully",
  "remaining_coins": 90
}
```

**Errores:**
- `400`: Episodio ya es gratis o monedas insuficientes
- `401`: No autenticado
- `404`: Episodio no encontrado

---

### GET /app/user/profile
Obtiene el perfil del usuario autenticado.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "firebase_uid": "firebase-uid",
    "coin_balance": 100,
    "is_premium": false,
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

---

## Admin Endpoints

Todos los endpoints de admin requieren rol de administrador.

### Series CRUD

#### GET /admin/series
Lista todas las series (incluye inactivas).

#### GET /admin/series/:id
Obtiene una serie por ID.

#### POST /admin/series
Crea una nueva serie.

**Body:**
```json
{
  "title": "Serie Title",
  "description": "Description",
  "horizontal_poster": "url",
  "vertical_poster": "url",
  "is_active": true
}
```

#### PUT /admin/series/:id
Actualiza una serie existente.

#### DELETE /admin/series/:id
Elimina una serie (soft delete, marca `is_active=false`).

---

### Episodes CRUD

#### POST /admin/episodes
Crea un nuevo episodio.

**Body:**
```json
{
  "series_id": "uuid",
  "episode_number": 1,
  "title": "Episode Title",
  "duration": 120,
  "is_free": false,
  "price_coins": 10
}
```

#### PUT /admin/episodes/:id
Actualiza un episodio existente.

---

### Video Upload Flow

#### POST /admin/episodes/:id/upload-url
Genera una URL presignada para subir video directamente a Bunny.net.

**Response:**
```json
{
  "upload_url": "https://video.bunnycdn.com/...",
  "episode_id": "uuid"
}
```

**Flujo:**
1. Admin llama este endpoint
2. Recibe `upload_url`
3. Sube video directamente a Bunny.net usando `upload_url`
4. Llama `/admin/episodes/:id/complete-upload` con `video_id_bunny`

#### POST /admin/episodes/:id/complete-upload
Marca el upload como completado y guarda el `video_id_bunny`.

**Body:**
```json
{
  "video_id_bunny": "bunny-video-id"
}
```

---

### Analytics

#### GET /admin/analytics
Retorna métricas básicas (placeholder, pendiente de implementación).

---

## Webhooks

### POST /webhooks/revenuecat
Endpoint para recibir webhooks de RevenueCat.

**Headers:** `Authorization: <webhook-signature>`

**Eventos procesados:**
- `INITIAL_PURCHASE`: Usuario compró suscripción → `is_premium=true`
- `RENEWAL`: Usuario renovó suscripción → `is_premium=true`
- `CANCELLATION`: Usuario canceló → `is_premium=false`
- `EXPIRATION`: Suscripción expiró → `is_premium=false`

---

## Health Check

### GET /health
Verifica el estado del servidor.

**Response:**
```json
{
  "status": "ok",
  "service": "qenti-api"
}
```

