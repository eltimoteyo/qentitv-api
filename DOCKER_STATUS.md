# 🐳 Estado del Despliegue Docker

## ✅ Servicios Desplegados

### 1. PostgreSQL Database
- **Container**: `qenti-postgres`
- **Puerto**: `5432:5432`
- **Estado**: ✅ Running
- **Base de datos**: `qenti`
- **Usuario**: `postgres`
- **Password**: `postgres`

### 2. pgAdmin (Opcional)
- **Container**: `qenti-pgadmin`
- **Puerto**: `5050:80`
- **URL**: http://localhost:5050
- **Email**: `admin@qenti.com`
- **Password**: `admin`

### 3. API Backend
- **Container**: `qenti-api-dev`
- **Puerto**: `8080:8080`
- **URL**: http://localhost:8080
- **Estado**: ✅ Running
- **Modo**: Development (con mock de Firebase)

## 🔗 URLs Importantes

- **API Base**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/api/v1/health (si existe)
- **pgAdmin**: http://localhost:5050

## 🔧 Comandos Útiles

### Ver logs del API
```bash
docker-compose logs -f api
```

### Ver logs de PostgreSQL
```bash
docker-compose logs -f postgres
```

### Detener todos los servicios
```bash
docker-compose down
```

### Reiniciar solo el API
```bash
docker-compose restart api
```

### Reconstruir y reiniciar
```bash
docker-compose up -d --build
```

### Acceder a la base de datos
```bash
docker-compose exec postgres psql -U postgres -d qenti
```

## 📝 Configuración

Las variables de entorno están configuradas en `docker-compose.yml`:
- Firebase está en modo mock (no requiere credenciales)
- JWT_SECRET: `dev-secret-key-change-in-production`
- Base de datos conecta automáticamente a `postgres` container

## ✅ Próximos Pasos

1. **Probar el frontend admin**:
   - Asegúrate de que `VITE_API_URL=http://localhost:8080/api/v1` en `.env` del frontend
   - El login rápido (dev) debería funcionar ahora

2. **Verificar que el API responde**:
   ```bash
   curl http://localhost:8080/api/v1/auth/login -X POST -H "Content-Type: application/json" -d '{"firebase_token":"mock-firebase-token-for-dev"}'
   ```

3. **Acceder a pgAdmin** (opcional):
   - http://localhost:5050
   - Conectar a PostgreSQL usando:
     - Host: `postgres`
     - Port: `5432`
     - User: `postgres`
     - Password: `postgres`
