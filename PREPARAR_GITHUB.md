# 📤 Preparar Repositorio para GitHub

## Pasos Rápidos

### 1. Inicializar Git (si no está inicializado)

```powershell
cd D:\PROJECTS\QENTITV\QENTITV-API
git init
```

### 2. Verificar .gitignore

Asegúrate de que `.gitignore` incluya:
- `.env` y `.env.production`
- `firebase-credentials.json`
- `*.exe` y binarios
- Logs y archivos temporales

### 3. Agregar Archivos

```powershell
git add .
```

### 4. Commit Inicial

```powershell
git commit -m "QENTITV API - Lista para desplegar"
```

### 5. Crear Repositorio en GitHub

1. Ve a https://github.com/new
2. Nombre: `qentitv-api` (o el que prefieras)
3. **NO** inicialices con README, .gitignore o licencia
4. Clic en "Create repository"

### 6. Conectar y Subir

```powershell
# Agregar remote (reemplaza TU_USUARIO)
git remote add origin https://github.com/TU_USUARIO/qentitv-api.git

# O si prefieres SSH:
# git remote add origin git@github.com:TU_USUARIO/qentitv-api.git

# Subir código
git branch -M main
git push -u origin main
```

### 7. Si el Repositorio es Privado

Si necesitas autenticación:

**Opción A: Personal Access Token**
1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generar nuevo token con permisos `repo`
3. Usar como contraseña al hacer push

**Opción B: SSH Key**
1. Generar SSH key: `ssh-keygen -t ed25519 -C "tu_email@example.com"`
2. Agregar a GitHub: Settings → SSH and GPG keys → New SSH key
3. Usar URL SSH: `git@github.com:TU_USUARIO/qentitv-api.git`

---

## ✅ Verificar que NO se Suban Archivos Sensibles

Antes de hacer push, verifica:

```powershell
# Ver qué archivos se van a subir
git status

# Verificar que .env.production NO esté en la lista
git ls-files | Select-String "\.env"
```

Si ves `.env.production` en la lista, **NO** hagas push. Agrega a `.gitignore`:

```powershell
# Verificar .gitignore
Get-Content .gitignore | Select-String "\.env"
```

---

## 📋 Checklist Antes de Push

- [ ] `.gitignore` configurado correctamente
- [ ] `.env.production` NO está en el repositorio
- [ ] `firebase-credentials.json` NO está en el repositorio
- [ ] Binarios (`*.exe`) NO están en el repositorio
- [ ] Repositorio creado en GitHub
- [ ] Remote configurado
- [ ] Listo para hacer push

---

## 🚀 Después de Subir a GitHub

Sigue las instrucciones en `DEPLOY_HOSTINGER.md` para desplegar en el VPS.
