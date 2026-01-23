# Tareas Pendientes - Estado Actual

## ✅ COMPLETADO

### Crítico para Producción
- ✅ Firebase Admin SDK integrado
- ✅ Generación de JWT con estructura propuesta (sub, role, email, jti, iat, exp)
- ✅ Refresh tokens con almacenamiento en DB
- ✅ Verificación de admin desde DB y Firebase
- ✅ Token signing de Bunny.net (HMAC SHA256)
- ✅ Rate limiting implementado

### Importante para MVP
- ✅ Tablas de DB: transactions, views, bans, user_roles, refresh_tokens
- ✅ Analytics básicos (dashboard con métricas)
- ✅ Algoritmo de feed básico (trending y recomendados)
- ✅ Validación básica de anuncios
- ✅ Historial de transacciones
- ✅ Registro de vistas

---

## 🔴 PENDIENTE - Crítico (Para producción real)

### 1. ✅ Implementación Real de Bans - COMPLETADO
- ✅ **Endpoint BanUser completo**: Guarda en DB, valida usuario, previene bans duplicados
- ✅ **Repositorio de bans**: Métodos completos para gestión de bans

### 2. ✅ Historial de Usuario en Admin - COMPLETADO
- ✅ **GetUserByID completo**: Historial real de visionado, transacciones y bans

### 3. ✅ Validación de Entrada - COMPLETADO
- ✅ **Validación de env vars**: Verifica variables críticas al iniciar

---

## 🟡 PENDIENTE - Importante (Para MVP completo)

### 4. Endpoints de Pago Faltantes
- [ ] **Estado de suscripción**: `GET /app/payment/subscription-status`
  - Verificar estado desde RevenueCat o DB
- [ ] **Planes disponibles**: `GET /app/payment/offer`
  - Retornar planes de suscripción disponibles

### 5. Algoritmo de Feed Mejorado
- [ ] **Trending real**: Basado en vistas recientes (últimas 24-48h)
  - Archivo: `api/v1/app/feed.go`
  - Usar tabla `views` para calcular trending
- [ ] **Recomendación personalizada**: Basado en historial del usuario
  - Series que el usuario ya vio
  - Series similares

### 6. Validación de Anuncios Avanzada
- [ ] **Integración con SDK de ads**: Validar realmente que el anuncio fue visto
  - Archivo: `api/v1/app/ads.go`
  - Integrar con AdMob, Unity Ads, etc.
  - Prevenir reutilización del mismo `ad_id`

---

## 🟢 PENDIENTE - Mejoras de UX

### 7. Búsqueda
- [ ] **Endpoint de búsqueda**: `GET /app/search?q=query`
  - Buscar en series y episodios
  - Búsqueda por título, descripción

### 8. Favoritos
- [ ] **Sistema de favoritos**:
  - `POST /app/series/{id}/favorite` - Agregar a favoritos
  - `DELETE /app/series/{id}/favorite` - Quitar de favoritos
  - `GET /app/user/favorites` - Listar favoritos
  - Tabla `favorites` en DB

### 9. Continuar Viendo
- [ ] **Track del último episodio visto**:
  - `GET /app/user/continue-watching` - Últimos episodios vistos
  - Usar tabla `views` para determinar progreso

### 10. Notificaciones Push
- [ ] **Sistema de notificaciones**:
  - Nuevos episodios de series favoritas
  - Ofertas especiales
  - Integración con FCM (Firebase Cloud Messaging)

---

## 🔵 PENDIENTE - Infraestructura

### 11. Testing
- [ ] **Tests unitarios**: Para repositorios y servicios
- [ ] **Tests de integración**: Para endpoints HTTP
- [ ] **Tests de carga**: Para validar performance

### 12. Logging y Monitoreo
- [ ] **Logging estructurado**: Implementar con Zap
- [ ] **Métricas**: Integración con Prometheus
- [ ] **Health checks avanzados**: Verificar DB, Bunny.net, Firebase
- [ ] **Error tracking**: Integración con Sentry

### 13. Migraciones de Base de Datos
- [ ] **Sistema de migraciones robusto**: Usar `golang-migrate` o similar
- [ ] **Rollback**: Capacidad de revertir migraciones
- [ ] **Seeds**: Datos de prueba para desarrollo

### 14. Validación y Configuración
- [ ] **Validación de env vars**: Al iniciar el servidor
- [ ] **Configuración por ambiente**: Diferentes configs para dev/staging/prod
- [ ] **Sanitización de inputs**: Validar y sanitizar todos los inputs

---

## 🟣 PENDIENTE - Seguridad y Performance

### 15. Seguridad Adicional
- [ ] **CORS restrictivo**: Configurar CORS por ambiente (actualmente permite todo)
- [ ] **HTTPS enforcement**: Redirigir HTTP a HTTPS en producción
- [ ] **Input validation**: Validar todos los parámetros de entrada con reglas específicas
- [ ] **SQL injection prevention**: Revisar que todos los queries usen parámetros (ya está bien, pero verificar)
- [ ] **XSS prevention**: Sanitizar outputs JSON si es necesario

### 16. Optimización de Performance
- [ ] **Caching**: Implementar cache para series populares (Redis)
- [ ] **Índices de DB**: Revisar y agregar índices adicionales si es necesario
- [ ] **Connection pooling**: Optimizar pool de conexiones a DB
- [ ] **Paginación**: Implementar en todos los listados grandes (feed, series, etc.)

---

## 📊 Resumen por Prioridad

### Para Lanzar MVP Básico (Falta poco):
1. ✅ Estructura completa
2. ✅ Autenticación y JWT
3. ✅ Endpoints principales
4. 🔴 **Completar BanUser** (5 min)
5. 🔴 **Completar GetUserByID con historial** (15 min)
6. 🟡 **Endpoints de pago** (30 min)

### Para MVP Completo:
7. 🟡 **Feed mejorado con trending real** (1 hora)
8. 🟡 **Validación avanzada de anuncios** (2 horas)
9. 🟢 **Búsqueda** (1 hora)
10. 🟢 **Favoritos** (2 horas)

### Para Producción Robusta:
11. 🔵 **Tests** (1-2 días)
12. 🔵 **Logging estructurado** (4 horas)
13. 🔵 **Monitoreo** (1 día)
14. 🟣 **CORS y seguridad** (2 horas)
15. 🟣 **Caching** (1 día)

---

## 🎯 Recomendación Inmediata

**Para tener un MVP funcional completo, falta:**

1. **Completar BanUser** (5 minutos) - Usar tabla `bans` ya creada
2. **Completar GetUserByID** (15 minutos) - Agregar historial real
3. **Endpoints de pago** (30 minutos) - subscription-status y offer
4. **Feed mejorado** (1 hora) - Trending basado en vistas reales

**Total estimado: ~2 horas para MVP completo funcional**

¿Quieres que implemente estos 4 puntos ahora?

