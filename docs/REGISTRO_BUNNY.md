# 🐰 Guía Completa: Registro y Configuración en Bunny.net

## 📋 Paso 1: Crear Cuenta en Bunny.net

1. **Ve a la página de registro:**
   - Abre tu navegador y visita: https://bunny.net
   - Haz clic en **"Sign Up"** o **"Get Started"** (botón en la esquina superior derecha)

2. **Completa el formulario de registro:**
   - Ingresa tu **email**
   - Crea una **contraseña**
   - Acepta los términos y condiciones
   - Haz clic en **"Create Account"**

3. **Verifica tu email:**
   - Revisa tu bandeja de entrada
   - Haz clic en el enlace de verificación que te enviaron

4. **Completa tu perfil (opcional):**
   - Puedes agregar información de facturación más tarde
   - Por ahora, puedes usar la cuenta en modo de prueba

---

## 🎬 Paso 2: Crear una Librería de Stream

1. **Accede al Dashboard:**
   - Una vez iniciada sesión, verás el dashboard principal
   - En el menú lateral izquierdo, busca **"Stream"** y haz clic

2. **Crear Nueva Librería:**
   - Haz clic en el botón **"Add Library"** o **"Create Library"**
   - Completa el formulario:
     - **Name:** Un nombre para tu librería (ej: "QENTITV Videos")
     - **Replication Regions:** Selecciona las regiones donde quieres que se repliquen los videos (puedes dejar las opciones por defecto)
   - Haz clic en **"Add Library"** o **"Create"**

3. **Espera a que se cree la librería:**
   - Esto puede tomar unos segundos
   - Verás la librería en la lista de librerías

---

## 🔑 Paso 3: Obtener las Credenciales

### 3.1 API Key y Library ID

1. **Abre tu librería:**
   - Haz clic en el nombre de la librería que acabas de crear
   - O haz clic en el ícono de configuración (⚙️) junto a la librería

2. **Encuentra la sección "API":**
   - Busca la pestaña o sección llamada **"API"** o **"Settings"**
   - Aquí encontrarás:
     - **API Key:** Una cadena larga de caracteres (ej: `abc123def456...`)
     - **Library ID:** Un número o UUID (ej: `123456` o `abc-def-123`)

3. **Copia estas credenciales:**
   - **⚠️ IMPORTANTE:** Guarda estas credenciales en un lugar seguro
   - Las necesitarás para configurar tu backend

### 3.2 CDN Hostname

1. **En la misma página de la librería:**
   - Busca la sección **"CDN"** o **"Pull Zone"**
   - O busca **"Hostname"** o **"Stream URL"**
   - Verás algo como: `abc123.b-cdn.net` o `video.bunnycdn.com/library/123`

2. **Copia el hostname:**
   - Si ves un hostname completo, cópialo
   - Si solo ves una URL, extrae el hostname (la parte antes de `/library/`)

### 3.3 Security Key (Opcional pero Recomendado)

1. **Ve a Configuración de Stream:**
   - En el menú lateral, ve a **"Stream"** → **"Settings"** o **"Security"**
   - O busca la sección de seguridad en la configuración de tu librería

2. **Genera o copia el Security Key:**
   - Si ya existe uno, cópialo
   - Si no existe, haz clic en **"Generate Security Key"** o **"Create Security Key"**
   - **⚠️ IMPORTANTE:** Solo se muestra una vez, guárdalo inmediatamente

---

## 📝 Paso 4: Configurar Variables de Entorno

Una vez que tengas todas las credenciales, necesitas configurarlas en tu backend.

### Opción A: Usar el Script de Validación (Recomendado)

Ejecuta el script que te pedirá las credenciales:

```powershell
cd QENTITV-API
.\validar-bunny.ps1
```

El script te pedirá:
- `BUNNY_STREAM_API_KEY`
- `BUNNY_STREAM_LIBRARY_ID`
- `BUNNY_CDN_HOSTNAME` (opcional)
- `BUNNY_SECURITY_KEY` (opcional)

### Opción B: Configurar Manualmente en PowerShell

Abre PowerShell y ejecuta:

```powershell
# Configurar para la sesión actual
$env:BUNNY_STREAM_API_KEY = "tu-api-key-aqui"
$env:BUNNY_STREAM_LIBRARY_ID = "tu-library-id-aqui"
$env:BUNNY_CDN_HOSTNAME = "tu-hostname.b-cdn.net"
$env:BUNNY_SECURITY_KEY = "tu-security-key-aqui"

# Para hacerlo permanente (solo para tu usuario)
[System.Environment]::SetEnvironmentVariable('BUNNY_STREAM_API_KEY', 'tu-api-key-aqui', 'User')
[System.Environment]::SetEnvironmentVariable('BUNNY_STREAM_LIBRARY_ID', 'tu-library-id-aqui', 'User')
[System.Environment]::SetEnvironmentVariable('BUNNY_CDN_HOSTNAME', 'tu-hostname.b-cdn.net', 'User')
[System.Environment]::SetEnvironmentVariable('BUNNY_SECURITY_KEY', 'tu-security-key-aqui', 'User')
```

### Opción C: Crear archivo .env (Si usas un gestor de variables)

Si tu proyecto usa un paquete para cargar `.env`, crea un archivo `.env` en la raíz de `QENTITV-API`:

```env
BUNNY_STREAM_API_KEY=tu-api-key-aqui
BUNNY_STREAM_LIBRARY_ID=tu-library-id-aqui
BUNNY_CDN_HOSTNAME=tu-hostname.b-cdn.net
BUNNY_SECURITY_KEY=tu-security-key-aqui
```

**⚠️ IMPORTANTE:** No subas el archivo `.env` a Git. Agrégalo a `.gitignore`.

---

## ✅ Paso 5: Validar la Conexión

Una vez configuradas las variables, ejecuta la validación:

```powershell
cd QENTITV-API
go run scripts/validate_bunny.go
```

O usa el script interactivo:

```powershell
.\validar-bunny.ps1
```

Si todo está correcto, verás:
```
✅ Conexión exitosa con Bunny.net
✅ Video de prueba creado exitosamente
✨ Validación completada
```

---

## 💰 Planes y Precios

Bunny.net ofrece un plan gratuito con límites:
- **Free Tier:** 1 GB de almacenamiento, 10 GB de ancho de banda/mes
- **Pay-as-you-go:** $0.01 por GB de almacenamiento, $0.01 por GB de tráfico

Para producción, considera:
- **Stream Plan:** Desde $1/mes por 1 TB de almacenamiento
- **Storage Plan:** Para videos grandes

**Nota:** Puedes empezar con el plan gratuito para pruebas.

---

## 🆘 Solución de Problemas

### "No puedo encontrar el API Key"
- Asegúrate de estar en la página de configuración de tu librería
- Busca la pestaña "API" o "Settings"
- Si no lo ves, intenta hacer clic en "Show API Key" o "Reveal"

### "El Library ID no funciona"
- Verifica que estés usando el ID correcto de la librería
- Asegúrate de que la librería esté activa (no eliminada)
- El Library ID puede ser un número o un UUID

### "Error 401 al validar"
- Verifica que el API Key sea correcto
- Asegúrate de copiar todo el API Key (puede ser muy largo)
- Verifica que no haya espacios al inicio o final

### "Error 404 al validar"
- Verifica que el Library ID sea correcto
- Asegúrate de que la librería exista y esté activa
- Intenta crear una nueva librería si el problema persiste

---

## 📚 Recursos Adicionales

- **Documentación oficial:** https://docs.bunny.net/
- **API Reference:** https://docs.bunny.net/reference/stream-api-overview
- **Dashboard:** https://bunny.net/dashboard

---

## 🎯 Resumen Rápido

1. ✅ Regístrate en https://bunny.net
2. ✅ Crea una librería de Stream
3. ✅ Copia: API Key, Library ID, CDN Hostname, Security Key
4. ✅ Configura las variables de entorno
5. ✅ Ejecuta la validación

¡Listo! Ya puedes empezar a subir videos a Bunny.net desde tu aplicación.
