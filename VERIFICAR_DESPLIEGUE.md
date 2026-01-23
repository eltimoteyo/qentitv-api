# ✅ Verificar Despliegue Exitoso

## 🎉 Despliegue Completado en Puerto 8081

Si ya desplegaste exitosamente, verifica lo siguiente:

---

## ✅ Checklist Post-Despliegue

### 1. Verificar Health Check

```bash
# En el servidor
curl http://localhost:8081/health

# Desde tu PC (reemplaza con la IP de tu VPS)
curl http://TU_IP_VPS:8081/health
```

**Respuesta esperada:**
```json
{"status":"ok","service":"qenti-api"}
```

### 2. Verificar Contenedores

```bash
docker ps
```

Deberías ver:
- `qenti-postgres-prod` (PostgreSQL)
- `qenti-api-prod` (API)

### 3. Ver Logs

```bash
# Ver logs del API
docker-compose -f docker-compose.prod.yml logs -f api

# Ver logs de PostgreSQL
docker-compose -f docker-compose.prod.yml logs -f postgres
```

### 4. Probar Endpoints

```bash
# Feed público
curl http://TU_IP_VPS:8081/api/v1/app/feed

# Series
curl http://TU_IP_VPS:8081/api/v1/app/series
```

---

## 📱 Actualizar App Flutter

**IMPORTANTE:** Actualiza la configuración de la app Flutter:

### Editar `qentitv_mobile/lib/core/config/app_config.dart`:

```dart
class AppConfig {
  // ⚠️ IMPORTANTE: Usa el puerto 8081 (el que configuraste)
  static const String baseUrl = 'http://TU_IP_VPS:8081/api/v1';
  //                                                      ^^^^
  //                                                      Puerto 8081
}
```

**Reemplaza `TU_IP_VPS` con la IP de tu VPS Hostinger.**

### Ejemplo:

```dart
static const String baseUrl = 'http://185.123.45.67:8081/api/v1';
```

---

## 🔒 Verificar Firewall

Asegúrate de que el puerto 8081 esté abierto:

```bash
# Verificar reglas de firewall
sudo ufw status

# Si no está abierto, abrirlo
sudo ufw allow 8081/tcp
sudo ufw reload
```

**También verifica en el panel de Hostinger:**
1. Ve a **Firewall** o **Security**
2. Agrega regla para puerto **8081** (TCP)

---

## 🧪 Probar desde la App

1. **Actualiza `app_config.dart`** con la IP y puerto correctos
2. **Recompila la app:**
   ```bash
   cd qentitv_mobile
   flutter run
   ```
3. **Prueba las funciones:**
   - Ver catálogo de series
   - Ver anuncio por monedas (requiere registro)
   - Ver episodios

---

## 📊 Comandos Útiles

```bash
# Ver estado de contenedores
docker-compose -f docker-compose.prod.yml ps

# Ver logs en tiempo real
docker-compose -f docker-compose.prod.yml logs -f api

# Reiniciar API
docker-compose -f docker-compose.prod.yml restart api

# Detener todo
docker-compose -f docker-compose.prod.yml down

# Ver uso de recursos
docker stats
```

---

## 🐛 Si Algo No Funciona

### "Connection refused" desde la app

**Causa:** Puerto incorrecto o firewall bloqueado

**Solución:**
1. Verifica que el puerto en `app_config.dart` sea **8081**
2. Verifica que el firewall permita el puerto 8081
3. Verifica que el API esté corriendo: `docker ps`

### "Timeout" desde la app

**Causa:** IP incorrecta o API no accesible

**Solución:**
1. Verifica la IP del VPS
2. Prueba desde el navegador: `http://TU_IP_VPS:8081/health`
3. Verifica logs: `docker-compose -f docker-compose.prod.yml logs api`

---

## ✅ Todo Listo

Si el health check responde OK, entonces:

1. ✅ API desplegada correctamente
2. ✅ Puerto 8081 configurado
3. ⚠️ **Falta:** Actualizar app Flutter con puerto 8081
4. ⚠️ **Falta:** Configurar firewall (si no lo hiciste)

---

**¡Felicitaciones por el despliegue exitoso! 🎉**
