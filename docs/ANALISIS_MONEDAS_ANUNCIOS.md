# 💰 Análisis: Sistema de Monedas por Ver Anuncios

## 📊 Estado Actual del Sistema

### ✅ Lo que está implementado:

1. **Desbloquear episodio directamente con anuncio:**
   - **Endpoint:** `POST /api/v1/app/ads/unlock-episode`
   - **Flujo:** Usuario ve anuncio → Desbloquea episodio directamente (sin dar monedas)
   - **Validación:** Valida que el anuncio fue visto (previene fraude)
   - **Límite:** Rate limit de 1 req/s, burst de 3

2. **Desbloquear episodio con monedas:**
   - **Endpoint:** `POST /api/v1/app/episodes/:id/unlock`
   - **Flujo:** Usuario gasta monedas → Desbloquea episodio
   - **Validación:** Verifica balance suficiente

3. **Estructura de base de datos:**
   - Tabla `transactions` tiene tipo `ad_reward` (preparado pero no usado)
   - Tabla `unlocks` tiene método `AD` (usado para desbloqueo directo)

### ❌ Lo que NO está implementado:

1. **Otorgar monedas por ver anuncios:**
   - No hay endpoint para recompensar monedas por ver anuncios
   - El tipo `ad_reward` en transacciones existe pero no se usa
   - No hay límite de anuncios por día/hora

---

## 🎯 Dos Modelos de Negocio Posibles

### Modelo 1: Desbloqueo Directo (Actual) ✅
```
Usuario ve anuncio → Desbloquea episodio directamente
```
**Ventajas:**
- ✅ Más simple
- ✅ Ya está implementado
- ✅ Menos pasos para el usuario

**Desventajas:**
- ⚠️ Usuario no acumula monedas
- ⚠️ No puede elegir qué desbloquear después

### Modelo 2: Monedas por Anuncios (Recomendado) ⭐
```
Usuario ve anuncio → Obtiene monedas → Usa monedas para desbloquear lo que quiera
```
**Ventajas:**
- ✅ Usuario acumula monedas
- ✅ Más flexibilidad (elige qué desbloquear)
- ✅ Mejor experiencia de usuario
- ✅ Permite estrategias de monetización (ej: ver 3 anuncios = 1 episodio)

**Desventajas:**
- ⚠️ Requiere implementar nuevo endpoint
- ⚠️ Más pasos para el usuario

---

## 🔧 Implementación Recomendada

### Opción A: Solo en la App (NO recomendada) ❌
**Problemas:**
- No hay validación del servidor
- Fácil de hacer fraude (usuario puede modificar el código)
- No hay control de límites
- No hay tracking de anuncios vistos

### Opción B: API + App (Recomendado) ✅

**Flujo completo:**

1. **App muestra anuncio** (usando AdMob/Unity Ads SDK)
2. **Usuario completa el anuncio** → SDK notifica a la app
3. **App llama al API** con el `ad_id` del SDK
4. **API valida el anuncio:**
   - Verifica que no se haya usado recientemente
   - Verifica límites (ej: máximo 10 anuncios/día)
   - Registra la transacción
5. **API otorga monedas** al usuario
6. **API responde** con el nuevo balance

---

## 📝 Endpoint Necesario

### POST /api/v1/app/ads/reward-coins

**Request:**
```json
{
  "ad_id": "ca-app-pub-123456789/123456789",  // ID del anuncio del SDK
  "ad_type": "rewarded"  // rewarded, interstitial, banner
}
```

**Response:**
```json
{
  "message": "Coins rewarded successfully",
  "coins_earned": 10,
  "new_balance": 150,
  "daily_limit_remaining": 7  // Anuncios restantes hoy
}
```

**Validaciones:**
- ✅ Verificar que el `ad_id` es válido (formato del SDK)
- ✅ Verificar que no se haya usado en los últimos 5 minutos
- ✅ Verificar límite diario (ej: máximo 10 anuncios/día)
- ✅ Verificar límite por hora (ej: máximo 3 anuncios/hora)
- ✅ Registrar transacción tipo `ad_reward`

**Rate Limiting:**
- 1 request por segundo
- Burst de 3

---

## 🏗️ Cambios Necesarios en el Backend

### 1. Nuevo Endpoint en `api/v1/app/ads.go`

```go
// RewardCoinsForAd otorga monedas por ver un anuncio
func (h *Handlers) RewardCoinsForAd(c *gin.Context) {
    // 1. Validar request
    // 2. Validar anuncio (no usado recientemente)
    // 3. Verificar límites diarios/horarios
    // 4. Calcular monedas a otorgar (configurable)
    // 5. Actualizar balance del usuario
    // 6. Registrar transacción tipo "ad_reward"
    // 7. Responder con nuevo balance
}
```

### 2. Configuración de Recompensas

Agregar a `config.go`:
```go
type AdRewardConfig struct {
    CoinsPerAd        int  // Monedas por anuncio (ej: 10)
    DailyLimit        int  // Límite diario (ej: 10)
    HourlyLimit       int  // Límite por hora (ej: 3)
    CooldownMinutes   int  // Tiempo entre anuncios (ej: 5)
}
```

### 3. Actualizar Validador de Anuncios

El `adsValidator` ya tiene validación básica, pero necesita:
- Verificar límites diarios/horarios
- Calcular recompensa
- Registrar transacción

### 4. Actualizar Repositorio de Transacciones

Ya existe, solo necesita usarse con tipo `ad_reward`.

---

## 📱 Cambios Necesarios en la App Flutter

### 1. Integrar SDK de Anuncios

**Opción recomendada: Google AdMob**
```yaml
dependencies:
  google_mobile_ads: ^3.0.0
```

### 2. Mostrar Anuncio Recompensado

```dart
// Mostrar anuncio
final RewardedAd? rewardedAd = await loadRewardedAd();

// Cuando el usuario completa el anuncio
rewardedAd?.show(
  onUserEarnedReward: (ad, reward) async {
    // Llamar al API para otorgar monedas
    await apiService.rewardCoinsForAd(
      adId: reward.adUnitId,
      adType: 'rewarded',
    );
  },
);
```

### 3. Llamar al API

```dart
Future<AdRewardResponse> rewardCoinsForAd({
  required String adId,
  required String adType,
}) async {
  final response = await dio.post(
    '/api/v1/app/ads/reward-coins',
    data: {
      'ad_id': adId,
      'ad_type': adType,
    },
  );
  return AdRewardResponse.fromJson(response.data);
}
```

---

## 🎮 Flujo Completo Recomendado

### Escenario: Usuario quiere monedas viendo anuncios

1. **Usuario abre la app** → Ve botón "Ver Anuncio por Monedas"
2. **Usuario presiona el botón** → App muestra anuncio (AdMob SDK)
3. **Usuario completa el anuncio** → SDK notifica a la app
4. **App llama al API:**
   ```
   POST /api/v1/app/ads/reward-coins
   {
     "ad_id": "ca-app-pub-.../123456",
     "ad_type": "rewarded"
   }
   ```
5. **API valida:**
   - ✅ Anuncio no usado recientemente
   - ✅ No excedió límite diario
   - ✅ No excedió límite por hora
6. **API otorga monedas:**
   - Actualiza `users.coin_balance`
   - Crea transacción tipo `ad_reward`
7. **API responde:**
   ```json
   {
     "coins_earned": 10,
     "new_balance": 150,
     "daily_limit_remaining": 7
   }
   ```
8. **App muestra confirmación:**
   - "¡Ganaste 10 monedas!"
   - "Balance: 150 monedas"
   - "Puedes ver 7 anuncios más hoy"

### Escenario: Usuario usa monedas para desbloquear

1. **Usuario elige episodio bloqueado**
2. **Usuario presiona "Desbloquear con Monedas"**
3. **App llama al API:**
   ```
   POST /api/v1/app/episodes/:id/unlock
   ```
4. **API verifica balance y desbloquea**
5. **Usuario puede ver el episodio**

---

## 🔐 Seguridad y Prevención de Fraude

### Validaciones Implementadas:

1. **Validación de Ad ID:**
   - Formato correcto (SDK de AdMob)
   - No usado recientemente (últimos 5 minutos)

2. **Límites de Tiempo:**
   - Máximo X anuncios por día
   - Máximo Y anuncios por hora
   - Cooldown entre anuncios

3. **Rate Limiting:**
   - 1 request por segundo
   - Burst de 3

### Validaciones Adicionales Recomendadas:

1. **Verificación con SDK:**
   - En producción, validar con AdMob Server-Side Verification
   - Verificar que el anuncio fue realmente visto

2. **Tracking de Dispositivo:**
   - Registrar device_id para prevenir múltiples cuentas
   - Detectar patrones sospechosos

3. **Análisis de Patrones:**
   - Detectar si un usuario ve anuncios demasiado rápido
   - Detectar si múltiples usuarios usan el mismo ad_id

---

## 📊 Configuración Recomendada

### Valores por Defecto:

```go
AdRewardConfig{
    CoinsPerAd:      10,  // 10 monedas por anuncio
    DailyLimit:      10,  // 10 anuncios por día
    HourlyLimit:     3,   // 3 anuncios por hora
    CooldownMinutes: 5,   // 5 minutos entre anuncios
}
```

### Cálculo de Recompensa:

- **Anuncio recompensado:** 10 monedas
- **Anuncio intersticial:** 5 monedas (opcional)
- **Banner:** 1 moneda (opcional, no recomendado)

### Límites:

- **Diario:** 10 anuncios = 100 monedas máximo/día
- **Por hora:** 3 anuncios = 30 monedas máximo/hora
- **Cooldown:** 5 minutos entre anuncios

---

## ✅ Conclusión y Recomendación

### Respuesta a tu pregunta:

**¿Es necesario trabajar con el API o sucede directo en la app?**

**Respuesta: DEBE trabajar con el API** porque:

1. ✅ **Seguridad:** El API valida que el anuncio fue visto realmente
2. ✅ **Prevención de fraude:** El API controla límites y cooldowns
3. ✅ **Tracking:** El API registra todas las transacciones
4. ✅ **Consistencia:** El API mantiene el balance centralizado
5. ✅ **Escalabilidad:** El API puede validar con AdMob Server-Side

### Implementación Recomendada:

1. **Backend (API):**
   - ✅ Crear endpoint `POST /api/v1/app/ads/reward-coins`
   - ✅ Validar anuncios y límites
   - ✅ Otorgar monedas y registrar transacciones

2. **Frontend (App Flutter):**
   - ✅ Integrar Google AdMob SDK
   - ✅ Mostrar anuncios recompensados
   - ✅ Llamar al API cuando el usuario complete el anuncio
   - ✅ Mostrar balance actualizado

### Próximos Pasos:

1. Implementar el endpoint en el backend
2. Integrar AdMob en la app Flutter
3. Probar el flujo completo
4. Configurar límites y recompensas según tu modelo de negocio

---

## 📚 Referencias

- **Google AdMob:** https://developers.google.com/admob
- **Flutter AdMob Plugin:** https://pub.dev/packages/google_mobile_ads
- **AdMob Server-Side Verification:** https://developers.google.com/admob/android/rewarded/server-side-verification
