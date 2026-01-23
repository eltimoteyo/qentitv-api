# 🚀 Despliegue en VPS Hostinger

## 📋 Flujo de Despliegue

1. **Subir código a GitHub**
2. **Conectarse al VPS**
3. **Clonar repositorio**
4. **Configurar variables de entorno**
5. **Desplegar con Docker Compose**

---

## 🔧 Paso 1: Preparar Repositorio GitHub

### 1.1 Crear .gitignore (si no existe)

Asegúrate de que `.gitignore` incluya:

```gitignore
# Variables de entorno
.env
.env.production
.env.local

# Credenciales
firebase-credentials.json
*.pem
*.key

# Binarios
*.exe
main.exe

# Logs
*.log

# Docker
.docker/
```

### 1.2 Subir a GitHub

```bash
# Inicializar repositorio (si no está inicializado)
git init

# Agregar archivos
git add .

# Commit
git commit -m "Initial commit - QENTITV API"

# Agregar remote
git remote add origin https://github.com/TU_USUARIO/qentitv-api.git

# Push
git push -u origin main
```

---

## 🖥️ Paso 2: Configurar VPS Hostinger

### 2.1 Conectarse al VPS

**SSH:**
```bash
ssh root@TU_IP_VPS
# O si usas usuario específico:
ssh usuario@TU_IP_VPS
```

**Puerto:** Generalmente 22 (verifica en el panel de Hostinger)

### 2.2 Instalar Dependencias

```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Instalar Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Agregar usuario al grupo docker (si no eres root)
sudo usermod -aG docker $USER

# Verificar instalación
docker --version
docker-compose --version
```

### 2.3 Instalar Git (si no está instalado)

```bash
sudo apt install git -y
```

---

## 📥 Paso 3: Clonar y Configurar

### 3.1 Clonar Repositorio

```bash
# Crear directorio para la aplicación
mkdir -p /opt/qentitv
cd /opt/qentitv

# Clonar repositorio
git clone https://github.com/TU_USUARIO/qentitv-api.git .
# O si es privado:
git clone https://TU_TOKEN@github.com/TU_USUARIO/qentitv-api.git .
```

### 3.2 Verificar Puerto Disponible

```bash
cd /opt/qentitv/qentitv-api
# O si clonaste directamente en /opt/qentitv:
cd /opt/qentitv

# Verificar qué puertos están disponibles
chmod +x verificar-puerto.sh  # ⚠️ IMPORTANTE: Dar permisos de ejecución
./verificar-puerto.sh

# O si aún da error:
bash verificar-puerto.sh

# O verificar manualmente
netstat -tuln | grep :8080
# Si muestra algo, el puerto está en uso
```

**Si el puerto 8080 está en uso**, el script te sugerirá otro puerto (ej: 8081, 8082, etc.)

### 3.3 Crear Archivo de Configuración

```bash
# Crear .env.production
nano .env.production
```

**Contenido de `.env.production`:**

```env
# Base de datos
DB_USER=postgres
DB_PASSWORD=TU_PASSWORD_SEGURO_AQUI
DB_NAME=qenti
DB_PORT=5432

# JWT
JWT_SECRET=TU_JWT_SECRET_MUY_SEGURO_AQUI

# Bunny.net
BUNNY_STREAM_LIBRARY_ID=585077
BUNNY_STREAM_API_KEY=b5d6fea7-1f28-4c2f-b33b36e581d4-0e61-4d28
BUNNY_CDN_HOSTNAME=vz-e8e1ad01-079.b-cdn.net
BUNNY_SECURITY_KEY=10f4f6f9-d7be-4f87-9451-da11aeeab667

# Firebase (opcional)
FIREBASE_PROJECT_ID=
FIREBASE_CREDENTIALS_PATH=

# RevenueCat (opcional)
REVENUECAT_API_KEY=
REVENUECAT_WEBHOOK_SECRET=

# Puerto del API (verifica que esté disponible con ./verificar-puerto.sh)
API_PORT=8080
# Si 8080 está ocupado, usa otro puerto (ej: 8081, 8082, 3000, etc.)
```

**Generar JWT_SECRET:**
```bash
openssl rand -base64 32
```

**Guardar:** `Ctrl+O`, `Enter`, `Ctrl+X`

### 3.3 Subir firebase-credentials.json (si usas Firebase)

```bash
# Desde tu PC local
scp firebase-credentials.json root@TU_IP_VPS:/opt/qentitv/qentitv-api/

# O crear el archivo directamente en el servidor
nano firebase-credentials.json
# Pegar el contenido del JSON
```

---

## 🚀 Paso 4: Desplegar

### 4.1 Usar Script de Despliegue

```bash
cd /opt/qentitv/qentitv-api

# Dar permisos de ejecución a todos los scripts
chmod +x *.sh

# Ejecutar despliegue
./deploy-server.sh

# O si da error de permisos:
bash deploy-server.sh
```

### 4.2 O Manualmente

```bash
cd /opt/qentitv/qentitv-api

# Construir y desplegar
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d --build

# Ver logs
docker-compose -f docker-compose.prod.yml logs -f api
```

---

## 🔍 Paso 5: Verificar Despliegue

### 5.1 Health Check

```bash
# Desde el servidor
curl http://localhost:8080/health

# Desde tu PC (reemplaza con la IP de tu VPS)
curl http://TU_IP_VPS:8080/health
```

**Respuesta esperada:**
```json
{"status":"ok","service":"qenti-api"}
```

### 5.2 Verificar Contenedores

```bash
docker ps
```

Deberías ver:
- `qenti-postgres-prod`
- `qenti-api-prod`

### 5.3 Ver Logs

```bash
# Logs del API
docker-compose -f docker-compose.prod.yml logs -f api

# Logs de PostgreSQL
docker-compose -f docker-compose.prod.yml logs -f postgres
```

---

## 🔒 Paso 6: Configurar Firewall

### 6.1 Abrir Puerto en Firewall

```bash
# Obtener el puerto configurado
API_PORT=$(grep "^API_PORT" .env.production | cut -d '=' -f2 | tr -d ' ' | tr -d '"' | tr -d "'")
API_PORT=${API_PORT:-8080}

echo "Abriendo puerto: $API_PORT"

# UFW (Ubuntu/Debian)
sudo ufw allow $API_PORT/tcp
sudo ufw reload

# O iptables
sudo iptables -A INPUT -p tcp --dport $API_PORT -j ACCEPT
sudo iptables-save
```

### 6.2 Verificar en Panel de Hostinger

1. Accede al panel de Hostinger
2. Ve a **Firewall** o **Security**
3. Agrega regla para el puerto configurado (ej: **8081**, **8082**, etc.) (TCP)
4. **IMPORTANTE:** Usa el mismo puerto que configuraste en `API_PORT` del `.env.production`

---

## 🌐 Paso 7: Configurar Dominio (Opcional)

### 7.1 Nginx como Reverse Proxy

```bash
# Instalar Nginx
sudo apt install nginx -y

# Crear configuración
sudo nano /etc/nginx/sites-available/qentitv-api
```

**Contenido:**

```nginx
server {
    listen 80;
    server_name api.tudominio.com;

    location / {
        # Reemplaza 8080 con el puerto que configuraste
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts para requests largos
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

**Activar:**
```bash
sudo ln -s /etc/nginx/sites-available/qentitv-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 7.2 SSL con Let's Encrypt

```bash
# Instalar Certbot
sudo apt install certbot python3-certbot-nginx -y

# Obtener certificado
sudo certbot --nginx -d api.tudominio.com

# Renovación automática (ya está configurada)
sudo certbot renew --dry-run
```

---

## 📱 Paso 8: Actualizar App Flutter

Edita `qentitv_mobile/lib/core/config/app_config.dart`:

```dart
class AppConfig {
  // Para API desplegada en VPS Hostinger
  // ⚠️ IMPORTANTE: Reemplaza 8080 con el puerto que configuraste
  static const String baseUrl = 'http://TU_IP_VPS:8080/api/v1';
  
  // Ejemplo si usaste puerto 8081:
  // static const String baseUrl = 'http://TU_IP_VPS:8081/api/v1';
  
  // O si configuraste dominio:
  // static const String baseUrl = 'https://api.tudominio.com/api/v1';
}
```

---

## 🔄 Actualizaciones Futuras

### Actualizar Código

```bash
# En el servidor
cd /opt/qentitv/qentitv-api

# Configurar estrategia de merge (si no lo has hecho)
git config pull.rebase false

# Obtener últimos cambios
git pull origin main

# Si hay conflictos, resolverlos primero
# Luego reconstruir y redesplegar
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d --build
```

### ⚠️ Si Sale Error de "Divergent Branches"

```bash
# Opción 1: Merge (recomendado)
git pull --no-rebase origin main

# Opción 2: Descartar cambios locales (si no son importantes)
git reset --hard origin/main
git pull origin main
```

**Ver:** `SOLUCION_GIT_PULL.md` para más detalles

### Script de Actualización Automática

```bash
# Crear script
nano /opt/qentitv/update-api.sh
```

**Contenido:**

```bash
#!/bin/bash
cd /opt/qentitv/qentitv-api
git pull origin main
docker-compose -f docker-compose.prod.yml --env-file .env.production up -d --build
docker-compose -f docker-compose.prod.yml logs -f api
```

**Dar permisos:**
```bash
chmod +x /opt/qentitv/update-api.sh
```

---

## 🐛 Troubleshooting

### Error: "Cannot connect to database"

```bash
# Verificar que PostgreSQL esté corriendo
docker ps

# Ver logs
docker-compose -f docker-compose.prod.yml logs postgres

# Reiniciar
docker-compose -f docker-compose.prod.yml restart postgres
```

### Error: "Port already in use" o "bind: address already in use"

**Solución:** Usa un puerto diferente

```bash
# Opción 1: Script automático (Recomendado)
./verificar-puerto.sh
# Te dirá qué puerto está disponible (ej: 8082)

# Actualizar .env.production
nano .env.production
# Cambiar API_PORT=8080 a API_PORT=8082

# Redesplegar
./deploy-server.sh
```

**O manualmente:**

```bash
# Ver qué puertos están en uso
sudo netstat -tulpn | grep LISTEN

# Ver qué usa el puerto 8080
sudo netstat -tulpn | grep :8080

# Actualizar .env.production con puerto disponible
nano .env.production
# Cambiar API_PORT=8080 a API_PORT=8081 (o el que esté disponible)

# Actualizar firewall
sudo ufw allow 8081/tcp
sudo ufw reload

# Redesplegar
./deploy-server.sh

# IMPORTANTE: Actualizar app Flutter con el nuevo puerto
```

**Ver:** `SOLUCION_PUERTO_OCUPADO.md` para guía completa

### Error: "Permission denied" en Docker

```bash
# Agregar usuario al grupo docker
sudo usermod -aG docker $USER

# Cerrar sesión y volver a entrar
exit
# Luego reconectar por SSH
```

### Error: "Out of memory"

```bash
# Ver uso de memoria
free -h

# Verificar límites de Docker
docker stats

# Si es necesario, aumentar swap o recursos del VPS
```

---

## 📊 Monitoreo

### Ver Estado de Contenedores

```bash
docker-compose -f docker-compose.prod.yml ps
```

### Ver Uso de Recursos

```bash
docker stats
```

### Ver Logs en Tiempo Real

```bash
docker-compose -f docker-compose.prod.yml logs -f
```

---

## ✅ Checklist de Despliegue

- [ ] Código subido a GitHub
- [ ] Docker instalado en VPS
- [ ] Repositorio clonado
- [ ] `.env.production` configurado
- [ ] `firebase-credentials.json` subido (si aplica)
- [ ] Docker Compose ejecutado
- [ ] Health check responde OK
- [ ] Firewall configurado (puerto 8080)
- [ ] Dominio configurado (opcional)
- [ ] SSL configurado (opcional)
- [ ] App Flutter actualizada con nueva URL

---

## 🆘 Comandos Útiles

```bash
# Ver estado
docker-compose -f docker-compose.prod.yml ps

# Ver logs
docker-compose -f docker-compose.prod.yml logs -f api

# Detener
docker-compose -f docker-compose.prod.yml down

# Reiniciar
docker-compose -f docker-compose.prod.yml restart api

# Reconstruir
docker-compose -f docker-compose.prod.yml up -d --build

# Limpiar (cuidado: elimina volúmenes)
docker-compose -f docker-compose.prod.yml down -v
```

---

**¡Listo para desplegar en Hostinger! 🚀**
