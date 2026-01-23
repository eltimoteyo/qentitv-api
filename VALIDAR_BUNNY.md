# ✅ Validar Conexión con Bunny.net

## 🔍 Validación Rápida

### Opción 1: Desde el Admin Panel

1. Abre el admin panel
2. Ve a cualquier página
3. Abre la consola del navegador (F12)
4. Ejecuta:

```javascript
fetch('/api/v1/admin/validate/bunny', {
  headers: {
    'Authorization': 'Bearer ' + localStorage.getItem('auth_token')
  }
})
.then(r => r.json())
.then(console.log)
```

### Opción 2: Desde la Terminal

```bash
# Desde la raíz del proyecto QENTITV-API
go run scripts/validate_bunny.go
```

### Opción 3: Con curl

```bash
curl -X GET http://localhost:8080/api/v1/admin/validate/bunny \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## ✅ Respuesta Esperada

**Si está conectado:**
```json
{
  "status": "ok",
  "message": "Conexión con Bunny.net exitosa"
}
```

**Si hay error:**
```json
{
  "status": "error",
  "error": "bunny API returned status 401 - check your API key",
  "message": "No se pudo conectar con Bunny.net. Verifica tus credenciales."
}
```

## 🔧 Configuración Requerida

Asegúrate de tener estas variables de entorno:

```env
BUNNY_STREAM_API_KEY=tu-api-key
BUNNY_STREAM_LIBRARY_ID=tu-library-id
BUNNY_CDN_HOSTNAME=tu-hostname.b-cdn.net
BUNNY_SECURITY_KEY=tu-security-key (opcional)
```

## 📤 Flujo de Upload Mejorado

1. **Admin solicita URL de upload** → Backend crea video en Bunny y retorna `upload_url` + `video_id`
2. **Admin sube video directamente** → PUT directo a Bunny.net (no pasa por el servidor)
3. **Admin completa registro** → Backend guarda `video_id` en la base de datos

**Ventajas:**
- ⚡ Upload más rápido (directo a Bunny)
- 📈 Escalable (Bunny maneja el ancho de banda)
- 💰 Menor costo (no consume recursos del servidor)
- 🔒 Seguro (URLs temporales y específicas)
