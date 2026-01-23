# 🚀 Guía Rápida de Despliegue

## 📋 Pasos Rápidos

### 1. Crear archivo `.env.production`

Crea un archivo `.env.production` en la raíz del proyecto con:

```env
# Base de datos
DB_USER=postgres
DB_PASSWORD=tu_password_seguro_aqui
DB_NAME=qenti
DB_PORT=5432

# JWT
JWT_SECRET=tu_jwt_secret_muy_seguro_aqui

# Bunny.net (ya tienes estas credenciales)
BUNNY_STREAM_LIBRARY_ID=585077
BUNNY_STREAM_API_KEY=b5d6fea7-1f28-4c2f-b33b36e581d4-0e61-4d28
BUNNY_CDN_HOSTNAME=vz-e8e1ad01-079.b-cdn.net
BUNNY_SECURITY_KEY=10f4f6f9-d7be-4f87-9451-da11aeeab667

# Firebase (opcional - dejar vacío para modo mock)
FIREBASE_PROJECT_ID=
FIREBASE_CREDENTIALS_PATH=

# RevenueCat (opcional)
REVENUECAT_API_KEY=
REVENUECAT_WEBHOOK_SECRET=

# Puerto del API (opcional, default: 8080)
API_PORT=8080
```

### 2. Generar JWT_SECRET

**PowerShell:**
```powershell
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
```

**O manualmente:** Usa cualquier string largo y seguro (mínimo 32 caracteres)

### 3. Desplegar

**Opción A: Script PowerShell (Recomendado)**
```powershell
.\deploy-prod.ps1
```

**Opción B: Manual**
```powershell
# Construir imagen
docker build -t qenti-api:latest .

# Desplegar
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
```

### 4. Verificar

```powershell
# Health check
curl http://localhost:8080/health

# Ver logs
docker-compose -f docker-compose.prod.yml logs -f api

# Ver estado
docker-compose -f docker-compose.prod.yml ps
```

---

## 🌐 Configurar para Acceso Externo

### Opción 1: IP Pública (Servidor/VPS)

Si desplegaste en un servidor con IP pública:

1. **Obtén la IP pública** del servidor
2. **Configura Firewall** para permitir puerto 8080
3. **Actualiza la app Flutter:**

```dart
// lib/core/config/app_config.dart
static const String baseUrl = 'http://TU_IP_PUBLICA:8080/api/v1';
```

### Opción 2: Dominio + Nginx (Recomendado para Producción)

1. **Configura Nginx como reverse proxy:**

```nginx
server {
    listen 80;
    server_name api.tudominio.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

2. **Configura SSL con Let's Encrypt:**

```bash
sudo certbot --nginx -d api.tudominio.com
```

3. **Actualiza la app Flutter:**

```dart
static const String baseUrl = 'https://api.tudominio.com/api/v1';
```

### Opción 3: ngrok (Para Pruebas Rápidas)

```powershell
# Instalar ngrok: https://ngrok.com/download
ngrok http 8080
```

Copia la URL (ej: `https://abc123.ngrok.io`) y actualiza la app:

```dart
static const String baseUrl = 'https://abc123.ngrok.io/api/v1';
```

---

## 📱 Actualizar App Flutter

Después de desplegar, actualiza `qentitv_mobile/lib/core/config/app_config.dart`:

```dart
class AppConfig {
  // Para API desplegada
  static const String baseUrl = 'https://api.tudominio.com/api/v1';
  
  // O si usas IP pública:
  // static const String baseUrl = 'http://TU_IP:8080/api/v1';
  
  // O si usas ngrok:
  // static const String baseUrl = 'https://abc123.ngrok.io/api/v1';
}
```

---

## 🔒 Seguridad

### Firewall

**Windows (si desplegas en Windows Server):**
```powershell
New-NetFirewallRule -DisplayName "QENTITV API" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
```

**Linux (Ubuntu/Debian):**
```bash
sudo ufw allow 8080/tcp
```

### HTTPS (Recomendado)

- Usa Nginx + Let's Encrypt (gratis)
- O Cloudflare (gratis, con proxy)
- O certificado SSL de tu proveedor

---

## 🐛 Troubleshooting

### "Connection refused"
- Verifica que Docker esté corriendo
- Verifica que el contenedor esté activo: `docker ps`
- Verifica el puerto: `netstat -an | findstr 8080`

### "Health check failed"
- Revisa logs: `docker-compose -f docker-compose.prod.yml logs api`
- Verifica variables de entorno
- Verifica que PostgreSQL esté corriendo

### "Cannot connect to database"
- Verifica credenciales en `.env.production`
- Verifica que PostgreSQL esté corriendo: `docker ps`
- Verifica logs de PostgreSQL: `docker-compose -f docker-compose.prod.yml logs postgres`

---

## ✅ Checklist Post-Despliegue

- [ ] API responde en `/health`
- [ ] Endpoints públicos funcionan (`/api/v1/app/feed`)
- [ ] Variables de entorno configuradas
- [ ] Firewall configurado (si es necesario)
- [ ] App Flutter actualizada con nueva URL
- [ ] HTTPS configurado (si es producción)
- [ ] Logs monitoreados
- [ ] Backups configurados (base de datos)

---

## 🎯 Próximos Pasos

1. **Desplegar API** → ✅ (este script)
2. **Actualizar app Flutter** con la URL del API
3. **Probar en dispositivo** físico
4. **Configurar dominio** (opcional, para producción)
5. **Configurar HTTPS** (recomendado para producción)

---

**¡Listo para desplegar! 🚀**
