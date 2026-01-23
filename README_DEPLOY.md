# 🚀 Guía Rápida de Despliegue

## Para VPS Hostinger

### 1. Subir a GitHub

```bash
git add .
git commit -m "API lista para desplegar"
git push origin main
```

### 2. En el Servidor VPS

```bash
# Conectarse por SSH
ssh root@TU_IP_VPS

# Instalar Docker (si no está instalado)
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Instalar Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Clonar repositorio
mkdir -p /opt/qentitv
cd /opt/qentitv
git clone https://github.com/TU_USUARIO/qentitv-api.git .
cd qentitv-api

# Crear .env.production
nano .env.production
# (Pegar configuración - ver DEPLOY_HOSTINGER.md)

# Desplegar
chmod +x deploy-server.sh
./deploy-server.sh
```

### 3. Verificar

```bash
curl http://localhost:8080/health
```

### 4. Configurar Firewall

```bash
sudo ufw allow 8080/tcp
sudo ufw reload
```

### 5. Actualizar App Flutter

Editar `qentitv_mobile/lib/core/config/app_config.dart`:

```dart
static const String baseUrl = 'http://TU_IP_VPS:8080/api/v1';
```

---

## Documentación Completa

- **`DEPLOY_HOSTINGER.md`** - Guía completa paso a paso
- **`GUIA_DESPLIEGUE_RAPIDO.md`** - Guía general de despliegue
- **`DEPLOY.md`** - Documentación completa del API

---

## Actualizar Código

```bash
# En el servidor
cd /opt/qentitv/qentitv-api
git pull origin main
./deploy-server.sh
```
