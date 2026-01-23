# TODO - Funcionalidades Pendientes

## 🔴 CRÍTICO (Para producción)

### 1. Autenticación Real
- [ ] **Firebase Admin SDK**: Implementar verificación real de tokens JWT
  - Archivo: `internal/pkg/auth/service.go`
  - Actualmente usa mock, debe verificar tokens reales de Firebase
- [ ] **Generación de JWT propios**: Crear tokens JWT con roles para la API
  - Archivo: `api/v1/auth/handlers.go`
  - Actualmente retorna tokens mock
- [ ] **Sistema de roles**: Implementar verificación de admin
  - Opciones: Custom claims en Firebase, tabla de roles en DB, o lista en config
  - Archivo: `internal/pkg/auth/service.go` - método `IsAdmin()`

### 2. Seguridad de Video Streaming
- [ ] **Token signing de Bunny.net**: Implementar generación real de tokens firmados
  - Archivo: `internal/pkg/bunny/service.go` - método `GetSignedPlaybackURL()`
  - Actualmente retorna URLs sin firma
  - Necesario para prevenir hotlinking

### 3. Rate Limiting
- [ ] Implementar rate limiting en endpoints críticos:
  - `/auth/login` y `/auth/refresh`
  - `/admin/auth/login`
  - `/app/episodes/{id}/unlock`
  - `/app/ads/unlock-episode`

---

## 🟡 IMPORTANTE (Para MVP funcional)

### 4. Base de Datos - Tablas Faltantes
- [ ] **Tabla de transacciones**: Para historial completo de wallet
  ```sql
  CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    type VARCHAR(20), -- 'unlock', 'purchase', 'gift', 'ad_reward'
    amount INTEGER,
    episode_id UUID REFERENCES episodes(id),
    method VARCHAR(20), -- 'COIN', 'AD', 'SUB', 'GIFT'
    created_at TIMESTAMP
  );
  ```
- [ ] **Tabla de reproducciones/vistas**: Para analytics y tracking
  ```sql
  CREATE TABLE views (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    episode_id UUID REFERENCES episodes(id),
    watched_seconds INTEGER,
    completed BOOLEAN,
    created_at TIMESTAMP
  );
  ```
- [ ] **Tabla de bans**: Para gestión de usuarios baneados
  ```sql
  CREATE TABLE bans (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    reason TEXT,
    banned_by UUID REFERENCES users(id),
    expires_at TIMESTAMP,
    created_at TIMESTAMP
  );
  ```
- [ ] **Tabla de roles**: Para gestión de permisos admin
  ```sql
  CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role VARCHAR(20), -- 'admin', 'moderator', 'user'
    granted_by UUID REFERENCES users(id),
    created_at TIMESTAMP,
    PRIMARY KEY (user_id, role)
  );
  ```

### 5. Analytics y Métricas
- [ ] **Retención por episodio**: Calcular tasa de abandono
  - Archivo: `api/v1/admin/dashboard.go`
- [ ] **Top dramas por reproducciones**: Ranking de contenido popular
- [ ] **Usuarios activos**: Contar usuarios activos en últimos 7/30 días
- [ ] **Ingresos por suscripciones**: Métricas de RevenueCat
- [ ] **Historial de visionado**: Para usuarios en admin panel
  - Archivo: `api/v1/admin/users.go` - método `GetUserByID()`

### 6. Algoritmo de Feed
- [ ] **Trending real**: Basado en vistas recientes y engagement
  - Archivo: `api/v1/app/feed.go`
- [ ] **Recomendación personalizada**: Basado en historial del usuario
- [ ] **Categorías/Tags**: Sistema de categorización de series

### 7. Validación de Anuncios
- [ ] **Verificación de ads**: Validar que el anuncio fue realmente visto
  - Archivo: `api/v1/app/ads.go`
  - Integrar con SDK de ads (AdMob, Unity Ads, etc.)
  - Prevenir fraude

---

## 🟢 MEJORAS (Para mejor UX)

### 8. Funcionalidades Adicionales
- [ ] **Búsqueda**: Endpoint para buscar series y episodios
  - `GET /app/search?q=query`
- [ ] **Favoritos**: Sistema de favoritos/seguimiento de series
  - `POST /app/series/{id}/favorite`
  - `GET /app/user/favorites`
- [ ] **Continuar viendo**: Track del último episodio visto
  - `GET /app/user/continue-watching`
- [ ] **Notificaciones**: Sistema de notificaciones push
  - Nuevos episodios de series favoritas
  - Ofertas especiales

### 9. Endpoints de Pago
- [ ] **Estado de suscripción**: `GET /app/payment/subscription-status`
- [ ] **Planes disponibles**: `GET /app/payment/offer`
- [ ] **Comprar monedas**: `POST /app/payment/purchase-coins` (si aplica)

### 10. Gestión de Contenido Admin
- [ ] **Categorías/Tags**: CRUD de categorías para series
- [ ] **Miniaturas automáticas**: Generar thumbnails de videos
- [ ] **Bulk operations**: Operaciones masivas (activar/desactivar múltiples series)

---

## 🔵 INFRAESTRUCTURA Y DEVOPS

### 11. Testing
- [ ] **Tests unitarios**: Para repositorios y servicios
- [ ] **Tests de integración**: Para endpoints HTTP
- [ ] **Tests de carga**: Para validar performance

### 12. Logging y Monitoreo
- [ ] **Logging estructurado**: Implementar con Zap o similar
- [ ] **Métricas**: Integración con Prometheus
- [ ] **Health checks avanzados**: Verificar DB, Bunny.net, Firebase
- [ ] **Error tracking**: Integración con Sentry o similar

### 13. Migraciones de Base de Datos
- [ ] **Sistema de migraciones**: Usar migrate o similar
- [ ] **Rollback**: Capacidad de revertir migraciones
- [ ] **Seeds**: Datos de prueba para desarrollo

### 14. Validación y Configuración
- [ ] **Validación de env vars**: Verificar que todas las variables requeridas estén presentes
- [ ] **Configuración por ambiente**: Diferentes configs para dev/staging/prod
- [ ] **Sanitización de inputs**: Validar y sanitizar todos los inputs

---

## 🟣 SEGURIDAD ADICIONAL

### 15. Mejoras de Seguridad
- [ ] **CORS restrictivo**: Configurar CORS por ambiente
- [ ] **HTTPS enforcement**: Redirigir HTTP a HTTPS en producción
- [ ] **Input validation**: Validar todos los parámetros de entrada
- [ ] **SQL injection prevention**: Asegurar que todos los queries usen parámetros
- [ ] **XSS prevention**: Sanitizar outputs JSON

### 16. Optimización de Performance
- [ ] **Caching**: Implementar cache para series populares
- [ ] **Índices de DB**: Agregar índices para queries frecuentes
- [ ] **Connection pooling**: Optimizar pool de conexiones a DB
- [ ] **Paginación**: Implementar en todos los listados grandes

---

## 📋 RESUMEN POR PRIORIDAD

### Para MVP Mínimo:
1. ✅ Estructura base y endpoints
2. 🔴 Firebase Auth real
3. 🔴 JWT generation real
4. 🔴 Admin role verification
5. 🔴 Bunny.net token signing
6. 🟡 Tabla de transacciones
7. 🟡 Tabla de vistas/reproducciones
8. 🟡 Rate limiting básico

### Para MVP Completo:
9. 🟡 Analytics básicos
10. 🟡 Algoritmo de feed básico
11. 🟡 Validación de anuncios
12. 🟢 Búsqueda
13. 🟢 Favoritos

### Para Producción:
14. 🔵 Tests completos
15. 🔵 Logging estructurado
16. 🔵 Monitoreo y métricas
17. 🔵 Migraciones robustas
18. 🔵 Validación completa

---

## 📝 NOTAS

- Los TODOs marcados con 🔴 son **bloqueantes** para producción
- Los marcados con 🟡 son **importantes** para MVP funcional
- Los marcados con 🟢 son **mejoras** de UX
- Los marcados con 🔵 son **infraestructura** necesaria para escalar

