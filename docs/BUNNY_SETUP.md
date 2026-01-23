# 🐰 Configuración de Bunny.net para Upload de Videos

## ✅ Validación de Conexión

### Método 1: Script de Validación

Ejecuta el script de validación:

```bash
go run scripts/validate_bunny.go
```

### Método 2: Endpoint de API

```bash
curl -X GET http://localhost:8080/api/v1/admin/validate/bunny \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## 🔧 Configuración Requerida

### Variables de Entorno

Asegúrate de tener estas variables configuradas:

```env
# Bunny Stream API
BUNNY_STREAM_API_KEY=tu-api-key-aqui
BUNNY_STREAM_LIBRARY_ID=tu-library-id-aqui

# Bunny CDN (para URLs de reproducción)
BUNNY_CDN_HOSTNAME=tu-hostname.b-cdn.net

# Bunny Security Key (opcional, para URLs firmadas)
BUNNY_SECURITY_KEY=tu-security-key-aqui
```

### Obtener Credenciales

1. **API Key y Library ID:**
   - Ve a https://bunny.net
   - Crea una cuenta o inicia sesión
   - Ve a "Stream" → "Libraries"
   - Crea una nueva librería o usa una existente
   - Copia el "API Key" y "Library ID"

2. **CDN Hostname:**
   - En la misma página de la librería
   - Busca "CDN Hostname" o "Pull Zone"
   - Copia el hostname (ej: `abc123.b-cdn.net`)

3. **Security Key (Opcional):**
   - Ve a "Stream" → "Settings"
   - Genera o copia el "Security Key"
   - Se usa para URLs firmadas con expiración

## 📤 Flujo de Upload Mejorado

### 1. Obtener URL de Upload

```bash
POST /api/v1/admin/episodes/:id/upload-url
Authorization: Bearer <admin-token>

Response:
{
  "upload_url": "https://video.bunnycdn.com/library/123/videos/abc-def",
  "video_id": "abc-def-ghi",
  "episode_id": "episode-uuid"
}
```

### 2. Subir Video Directamente

El admin sube el video directamente a `upload_url` usando **PUT**:

```javascript
const xhr = new XMLHttpRequest();
xhr.open('PUT', upload_url);
xhr.setRequestHeader('Content-Type', file.type);
xhr.send(file);
```

**Ventajas:**
- ✅ Upload directo (no pasa por el servidor Go)
- ✅ Más rápido y escalable
- ✅ Bunny maneja el ancho de banda
- ✅ Soporta archivos grandes sin problemas

### 3. Completar Registro

```bash
POST /api/v1/admin/episodes/:id/complete
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "video_id_bunny": "abc-def-ghi"
}

Response:
{
  "message": "Upload completed successfully"
}
```

## 🚀 Optimizaciones Implementadas

### 1. Upload Directo
- El video se sube directamente a Bunny.net
- No consume recursos del servidor Go
- Escalable para múltiples uploads simultáneos

### 2. Timeout Extendido
- Timeout de 30 minutos para videos grandes
- Manejo de errores mejorado

### 3. Validación de Estado
- Verifica el estado del video antes de completar
- Manejo de errores no críticos

### 4. Progress Tracking
- Barra de progreso en tiempo real
- Feedback visual durante el upload

## 🐛 Solución de Problemas

### Error: "bunny API error: 401"
- **Causa:** API Key inválido o expirado
- **Solución:** Verifica `BUNNY_STREAM_API_KEY` en las variables de entorno

### Error: "bunny API error: 404"
- **Causa:** Library ID incorrecto
- **Solución:** Verifica `BUNNY_STREAM_LIBRARY_ID`

### Error: "Upload failed: timeout"
- **Causa:** Video muy grande o conexión lenta
- **Solución:** 
  - Aumenta el timeout en el admin
  - Considera comprimir el video antes de subir
  - Verifica tu conexión a internet

### Error: "Failed to verify video status"
- **Causa:** El video aún se está procesando en Bunny
- **Solución:** Esto es normal, el video se procesará en segundo plano

## 📊 Monitoreo

### Verificar Estado de Video

```bash
GET https://video.bunnycdn.com/library/{library_id}/videos/{video_id}
Headers:
  AccessKey: {api_key}
```

### Estados del Video:
- `0` = Created
- `1` = Uploading
- `2` = Processing
- `3` = Queued
- `4` = Finished
- `5` = Error

## 🔐 Seguridad

- ✅ Las URLs de upload son temporales y específicas por video
- ✅ Solo el admin puede generar URLs de upload
- ✅ Las URLs de reproducción pueden ser firmadas con expiración
- ✅ No se almacenan videos en el servidor

---

**Nota:** El upload directo es la forma más eficiente y rápida de subir videos a Bunny.net. No requiere que el servidor Go maneje archivos grandes.
