# 🚀 Referencia Rápida - Endpoints Admin

## Base URL
```
http://localhost:8080/api/v1
```

## Autenticación
```
Authorization: Bearer <JWT_TOKEN>
```

---

## 📊 Dashboard

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/dashboard` | Métricas y gráficas del dashboard |

**Response:**
```json
{
  "metrics": {
    "total_series": 15,
    "total_episodes": 120,
    "total_users": 1250,
    "active_users_7d": 450,
    "active_users_30d": 980,
    "premium_users": 85
  },
  "charts": {
    "retention_by_episode": [...],
    "top_dramas": [...]
  }
}
```

---

## 🎬 Series

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/series` | Listar todas las series |
| GET | `/admin/series/:id` | Obtener serie por ID |
| POST | `/admin/series` | Crear nueva serie |
| PUT | `/admin/series/:id` | Actualizar serie |
| DELETE | `/admin/series/:id` | Eliminar serie (soft delete) |

**POST /admin/series Body:**
```json
{
  "title": "Serie Title",
  "description": "Description",
  "horizontal_poster": "https://...",
  "vertical_poster": "https://...",
  "is_active": true
}
```

---

## 📺 Episodes

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/episodes?series_id=uuid` | Listar episodios (opcional: filtrar por serie) |
| GET | `/admin/episodes/:id` | Obtener episodio por ID |
| POST | `/admin/episodes` | Crear nuevo episodio |
| PUT | `/admin/episodes/:id` | Actualizar episodio |
| DELETE | `/admin/episodes/:id` | Eliminar episodio |
| POST | `/admin/episodes/:id/upload-url` | Obtener URL para subir video |
| POST | `/admin/episodes/:id/complete` | Completar upload de video |

**POST /admin/episodes Body:**
```json
{
  "series_id": "uuid",
  "episode_number": 1,
  "title": "Episode Title",
  "duration": 180,
  "is_free": true,
  "price_coins": 10
}
```

**POST /admin/episodes/:id/complete Body:**
```json
{
  "video_id_bunny": "bunny-video-id"
}
```

---

## 👥 Users

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/admin/users?page=1&limit=20` | Listar usuarios (paginado) |
| GET | `/admin/users/:id` | Obtener detalle de usuario |
| PUT | `/admin/users/:id/ban` | Banear usuario |
| PUT | `/admin/users/:id/coins` | Regalar monedas |

**PUT /admin/users/:id/ban Body:**
```json
{
  "reason": "Violación de términos",
  "expires_at": "2024-02-15T00:00:00Z"
}
```

**PUT /admin/users/:id/coins Body:**
```json
{
  "amount": 100
}
```

---

## 📋 Schemas Comunes

### Series
```typescript
{
  id: string (UUID)
  title: string
  description: string
  horizontal_poster: string (URL)
  vertical_poster: string (URL)
  is_active: boolean
  created_at: string (ISO 8601)
  updated_at: string (ISO 8601)
}
```

### Episode
```typescript
{
  id: string (UUID)
  series_id: string (UUID)
  episode_number: number
  title: string
  video_id_bunny: string
  duration: number (segundos)
  is_free: boolean
  price_coins: number
  created_at: string (ISO 8601)
  updated_at: string (ISO 8601)
}
```

### User
```typescript
{
  id: string (UUID)
  email: string
  firebase_uid: string
  coin_balance: number
  is_premium: boolean
  created_at: string (ISO 8601)
  updated_at: string (ISO 8601)
}
```

---

## 🔧 Para Generadores de Interfaces

### Importar OpenAPI Spec
```bash
# Archivo disponible en:
docs/admin-api-openapi.json
```

### Herramientas compatibles:
- **Swagger UI**: Importar `admin-api-openapi.json`
- **Postman**: Importar OpenAPI spec
- **Insomnia**: Importar OpenAPI spec
- **React Admin**: Usar endpoints directamente
- **Retool**: Conectar con API REST
- **Appsmith**: Usar endpoints REST

### Ejemplo de uso en generador:

```javascript
// Configuración base
const API_BASE = 'http://localhost:8080/api/v1';
const TOKEN = 'tu-jwt-token';

// Headers comunes
const headers = {
  'Authorization': `Bearer ${TOKEN}`,
  'Content-Type': 'application/json'
};

// Ejemplo: Obtener dashboard
fetch(`${API_BASE}/admin/dashboard`, { headers });

// Ejemplo: Crear serie
fetch(`${API_BASE}/admin/series`, {
  method: 'POST',
  headers,
  body: JSON.stringify({
    title: 'Nueva Serie',
    description: 'Descripción',
    is_active: true
  })
});
```

---

## 📝 Notas para UI

### Formularios Recomendados:

1. **Crear Serie:**
   - Campo requerido: `title`
   - Campos opcionales: `description`, `horizontal_poster`, `vertical_poster`, `is_active` (checkbox)

2. **Crear Episodio:**
   - Campos requeridos: `series_id` (select), `episode_number` (number), `title`
   - Campos opcionales: `duration`, `is_free` (checkbox), `price_coins` (number)

3. **Banear Usuario:**
   - Campo requerido: `reason` (textarea)
   - Campo opcional: `expires_at` (datetime picker, null = permanente)

4. **Regalar Monedas:**
   - Campo requerido: `amount` (number, min: 1)

### Tablas Recomendadas:

1. **Lista de Series:**
   - Columnas: ID, Título, Estado (is_active), Fecha creación, Acciones (Editar/Eliminar)

2. **Lista de Episodios:**
   - Columnas: ID, Serie, Número, Título, Duración, Gratis, Precio, Acciones

3. **Lista de Usuarios:**
   - Columnas: ID, Email, Monedas, Premium, Fecha registro, Acciones

### Componentes de Dashboard:

1. **Métricas (Cards):**
   - Total Series
   - Total Episodios
   - Total Usuarios
   - Usuarios Activos (7d)
   - Usuarios Premium

2. **Gráficas:**
   - Retención por Episodio (line chart)
   - Top Dramas (bar chart)
   - Revenue (line chart - placeholder)

---

**Para documentación completa, ver:** `ADMIN_API_SPEC.md`

