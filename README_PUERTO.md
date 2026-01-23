# 🔌 Configurar Puerto del API

## ⚠️ Problema Común

Si el puerto **8080** ya está en uso por otra aplicación en el VPS, necesitas usar un puerto diferente.

---

## 🔍 Verificar Puertos Disponibles

### Opción 1: Script Automático (Recomendado)

```bash
cd /opt/qentitv/qentitv-api
chmod +x verificar-puerto.sh
./verificar-puerto.sh
```

El script te mostrará el primer puerto disponible entre 8080-8090.

### Opción 2: Manual

```bash
# Ver qué puertos están en uso
sudo netstat -tulpn | grep LISTEN

# O con ss (más moderno)
sudo ss -tulpn | grep LISTEN

# Verificar puerto específico
sudo netstat -tulpn | grep :8080
# Si muestra algo, el puerto está en uso
```

---

## 📝 Configurar Puerto Diferente

### 1. Editar .env.production

```bash
nano .env.production
```

Cambiar:
```env
API_PORT=8080
```

Por:
```env
API_PORT=8081
# O cualquier puerto disponible (8082, 3000, 5000, etc.)
```

### 2. Redesplegar

```bash
./deploy-server.sh
```

---

## 🔒 Configurar Firewall

Después de cambiar el puerto, actualiza el firewall:

```bash
# Obtener el puerto configurado
API_PORT=$(grep API_PORT .env.production | cut -d '=' -f2 | tr -d ' ')

# Abrir puerto en firewall
sudo ufw allow $API_PORT/tcp
sudo ufw reload
```

---

## 📱 Actualizar App Flutter

**IMPORTANTE:** Actualiza la URL en la app Flutter con el puerto correcto:

```dart
// lib/core/config/app_config.dart
class AppConfig {
  // ⚠️ Usa el mismo puerto que configuraste en .env.production
  static const String baseUrl = 'http://TU_IP_VPS:8081/api/v1';
  //                                                      ^^^^
  //                                                      Cambiar aquí
}
```

---

## 🌐 Si Usas Nginx

Si configuraste Nginx como reverse proxy, actualiza la configuración:

```nginx
server {
    listen 80;
    server_name api.tudominio.com;

    location / {
        # Cambiar 8080 por el puerto que configuraste
        proxy_pass http://localhost:8081;
        # ...
    }
}
```

Luego reinicia Nginx:
```bash
sudo nginx -t
sudo systemctl restart nginx
```

---

## ✅ Verificar

```bash
# Obtener puerto configurado
API_PORT=$(grep API_PORT .env.production | cut -d '=' -f2 | tr -d ' ')

# Health check
curl http://localhost:$API_PORT/health

# Desde tu PC
curl http://TU_IP_VPS:$API_PORT/health
```

---

## 📋 Puertos Comunes Alternativos

Si 8080 está ocupado, puedes usar:
- **8081** - Común para APIs alternativas
- **8082** - Otra opción común
- **3000** - Popular para Node.js
- **5000** - Popular para Flask
- **9000** - Alternativa común
- **Cualquier puerto > 1024** - Evita puertos del sistema (< 1024)

**Recomendación:** Usa **8081** o **8082** para mantener consistencia.

---

## 🐛 Troubleshooting

### "Bind: address already in use"

**Causa:** El puerto está ocupado

**Solución:**
1. Usa `./verificar-puerto.sh` para encontrar puerto disponible
2. Actualiza `API_PORT` en `.env.production`
3. Redesplega con `./deploy-server.sh`

### "Connection refused" desde la app

**Causa:** Puerto incorrecto en la app Flutter

**Solución:**
1. Verifica el puerto en `.env.production` del servidor
2. Actualiza `app_config.dart` con el mismo puerto
3. Recompila la app Flutter

---

**¡Listo! Ahora puedes usar cualquier puerto disponible** 🚀
