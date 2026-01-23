# 🔌 Solución: Puerto Ya en Uso

## 🔴 Error Común

```
Error: bind: address already in use
Error: port is already allocated
```

Esto significa que el puerto configurado (ej: 8080) ya está siendo usado por otra aplicación.

---

## ✅ Solución Rápida

### Paso 1: Verificar Puerto Disponible

```bash
cd /opt/qentitv/qentitv-api

# Usar script automático
chmod +x verificar-puerto.sh
./verificar-puerto.sh
```

**Ejemplo de salida:**
```
⚠️  Puerto 8080 está en uso
⚠️  Puerto 8081 está en uso
✅ Puerto disponible encontrado: 8082
```

### Paso 2: Actualizar .env.production

```bash
nano .env.production
```

Cambiar:
```env
API_PORT=8080
```

Por el puerto disponible:
```env
API_PORT=8082
```

### Paso 3: Redesplegar

```bash
./deploy-server.sh
```

---

## 🔍 Verificar Manualmente

### Ver Qué Usa el Puerto

```bash
# Ver qué proceso usa el puerto 8080
sudo netstat -tulpn | grep :8080

# O con ss (más moderno)
sudo ss -tulpn | grep :8080

# O con lsof
sudo lsof -i :8080
```

**Ejemplo de salida:**
```
tcp  0  0  0.0.0.0:8080  0.0.0.0:*  LISTEN  1234/docker-proxy
```

Esto te dice qué proceso (PID 1234) está usando el puerto.

### Ver Todos los Puertos en Uso

```bash
# Ver todos los puertos
sudo netstat -tulpn | grep LISTEN

# O con ss
sudo ss -tulpn | grep LISTEN
```

---

## 🎯 Opciones

### Opción 1: Usar Otro Puerto (Recomendado)

Si el puerto está ocupado por otra API importante:

1. **Encontrar puerto disponible:**
   ```bash
   ./verificar-puerto.sh
   ```

2. **Actualizar .env.production:**
   ```bash
   nano .env.production
   # Cambiar API_PORT=8080 a API_PORT=8082
   ```

3. **Actualizar app Flutter:**
   ```dart
   // lib/core/config/app_config.dart
   static const String baseUrl = 'http://TU_IP_VPS:8082/api/v1';
   ```

4. **Redesplegar:**
   ```bash
   ./deploy-server.sh
   ```

### Opción 2: Detener el Servicio que Usa el Puerto

Si el puerto está ocupado por un servicio que puedes detener:

```bash
# Ver qué proceso usa el puerto
sudo lsof -i :8080
# O
sudo netstat -tulpn | grep :8080

# Detener el proceso (reemplaza PID con el número que aparezca)
sudo kill -9 PID

# O si es un contenedor Docker
docker ps
docker stop NOMBRE_CONTENEDOR
```

**⚠️ Cuidado:** Asegúrate de que el servicio que detienes no sea crítico.

### Opción 3: Cambiar Puerto del Otro Servicio

Si puedes modificar la otra API:

1. Detener la otra API
2. Cambiar su puerto
3. Reiniciarla
4. Usar 8080 para QENTITV

---

## 🚀 Script Automático

El script `deploy-server.sh` ahora verifica automáticamente si el puerto está en uso:

```bash
./deploy-server.sh
```

Si detecta que el puerto está ocupado:
- Te avisa
- Ejecuta `verificar-puerto.sh` para encontrar uno disponible
- Te pregunta si quieres continuar o cancelar

---

## 📋 Checklist

- [ ] Verificar puerto disponible con `./verificar-puerto.sh`
- [ ] Actualizar `API_PORT` en `.env.production`
- [ ] Actualizar `baseUrl` en app Flutter
- [ ] Configurar firewall para el nuevo puerto
- [ ] Redesplegar con `./deploy-server.sh`

---

## 🔒 Actualizar Firewall

Después de cambiar el puerto, actualiza el firewall:

```bash
# Obtener nuevo puerto
NEW_PORT=$(grep "^API_PORT" .env.production | cut -d '=' -f2 | tr -d ' ')

# Abrir nuevo puerto
sudo ufw allow $NEW_PORT/tcp

# Cerrar puerto viejo (opcional)
sudo ufw delete allow 8080/tcp

# Recargar
sudo ufw reload
```

---

## 🌐 Actualizar Nginx (Si Usas)

Si configuraste Nginx, actualiza el proxy_pass:

```bash
sudo nano /etc/nginx/sites-available/qentitv-api
```

Cambiar:
```nginx
proxy_pass http://localhost:8080;
```

Por:
```nginx
proxy_pass http://localhost:8082;  # Tu nuevo puerto
```

Reiniciar:
```bash
sudo nginx -t
sudo systemctl restart nginx
```

---

## 🐛 Troubleshooting

### "Error: bind: address already in use" Durante Despliegue

**Causa:** Puerto ocupado

**Solución:**
1. Ejecuta `./verificar-puerto.sh`
2. Actualiza `API_PORT` en `.env.production`
3. Redesplega

### "Connection refused" desde la App

**Causa:** Puerto incorrecto en la app Flutter

**Solución:**
1. Verifica el puerto en `.env.production` del servidor
2. Actualiza `app_config.dart` con el mismo puerto
3. Recompila la app

### Puerto Cambia Cada Vez

**Causa:** No guardaste el puerto en `.env.production`

**Solución:**
- Asegúrate de que `API_PORT` esté en `.env.production`
- No uses variables de entorno temporales

---

## 💡 Prevenir en el Futuro

1. **Siempre verifica el puerto antes de desplegar:**
   ```bash
   ./verificar-puerto.sh
   ```

2. **Documenta qué puertos usas:**
   - Crea un archivo `PUERTOS.md` con los puertos de cada servicio

3. **Usa puertos consistentes:**
   - API 1: 8080
   - API 2: 8081
   - API 3: 8082
   - etc.

---

**¡Listo para resolver el problema de puerto! 🚀**
