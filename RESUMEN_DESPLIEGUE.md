# 🚀 Resumen Rápido - Despliegue en VPS Hostinger

## ⚠️ IMPORTANTE: Puerto

Si ya tienes otras APIs, el puerto **8080** probablemente esté ocupado. 
**Solución:** Usa otro puerto (8081, 8082, etc.)

---

## 📋 Pasos Rápidos

### 1. En tu PC: Subir a GitHub

```powershell
cd D:\PROJECTS\QENTITV\QENTITV-API
.\subir-github.ps1
```

### 2. En el VPS: Conectarse y Clonar

```bash
ssh root@TU_IP_VPS
mkdir -p /opt/qentitv
cd /opt/qentitv
git clone https://github.com/TU_USUARIO/qentitv-api.git .
cd qentitv-api
```

### 3. Verificar Puerto Disponible

```bash
chmod +x verificar-puerto.sh
./verificar-puerto.sh
```

**Ejemplo de salida:**
```
✅ Puerto disponible encontrado: 8081
```

### 4. Crear .env.production

```bash
nano .env.production
```

**Contenido (usa el puerto que encontraste):**

```env
DB_USER=postgres
DB_PASSWORD=tu_password_seguro
DB_NAME=qenti
DB_PORT=5432

JWT_SECRET=$(openssl rand -base64 32)

BUNNY_STREAM_LIBRARY_ID=585077
BUNNY_STREAM_API_KEY=b5d6fea7-1f28-4c2f-b33b36e581d4-0e61-4d28
BUNNY_CDN_HOSTNAME=vz-e8e1ad01-079.b-cdn.net
BUNNY_SECURITY_KEY=10f4f6f9-d7be-4f87-9451-da11aeeab667

FIREBASE_PROJECT_ID=
REVENUECAT_API_KEY=
REVENUECAT_WEBHOOK_SECRET=

# ⚠️ IMPORTANTE: Usa el puerto que encontraste (ej: 8081)
API_PORT=8081
```

### 5. Desplegar

```bash
chmod +x deploy-server.sh
./deploy-server.sh
```

### 6. Configurar Firewall

```bash
# Obtener puerto configurado
API_PORT=$(grep "^API_PORT" .env.production | cut -d '=' -f2 | tr -d ' ')

# Abrir puerto
sudo ufw allow $API_PORT/tcp
sudo ufw reload
```

### 7. Verificar

```bash
# Obtener puerto
API_PORT=$(grep "^API_PORT" .env.production | cut -d '=' -f2 | tr -d ' ')

# Health check
curl http://localhost:$API_PORT/health
```

### 8. Actualizar App Flutter

Edita `qentitv_mobile/lib/core/config/app_config.dart`:

```dart
class AppConfig {
  // ⚠️ Usa el MISMO puerto que configuraste en .env.production
  static const String baseUrl = 'http://TU_IP_VPS:8081/api/v1';
  //                                                      ^^^^
  //                                                      Cambiar aquí
}
```

---

## 🔍 Verificar Puerto en Uso

```bash
# Ver todos los puertos en uso
sudo netstat -tulpn | grep LISTEN

# Verificar puerto específico
sudo netstat -tulpn | grep :8080
# Si muestra algo, está ocupado
```

---

## 📚 Documentación Completa

- **`DEPLOY_HOSTINGER.md`** - Guía completa paso a paso
- **`README_PUERTO.md`** - Guía específica para configurar puertos
- **`verificar-puerto.sh`** - Script para encontrar puerto disponible

---

## ✅ Checklist

- [ ] Código subido a GitHub
- [ ] VPS conectado por SSH
- [ ] Docker instalado
- [ ] Repositorio clonado
- [ ] Puerto verificado (ej: 8081)
- [ ] `.env.production` creado con puerto correcto
- [ ] Desplegado con `./deploy-server.sh`
- [ ] Firewall configurado
- [ ] Health check OK
- [ ] App Flutter actualizada con puerto correcto

---

**¡Listo! 🚀**
