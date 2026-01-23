# 🔧 Solución: Git Pull con Ramas Divergentes

## 🔴 Problema

Cuando ejecutas `git pull` y hay cambios locales que no están en GitHub, Git no sabe cómo combinarlos.

---

## ✅ Soluciones

### Opción 1: Merge (Recomendado para Producción)

Combina los cambios locales con los de GitHub:

```bash
git config pull.rebase false
git pull origin main
```

**O directamente:**
```bash
git pull --no-rebase origin main
```

### Opción 2: Descartar Cambios Locales (Si no son importantes)

Si los cambios locales no son importantes y quieres usar solo lo de GitHub:

```bash
# Ver qué archivos cambiaron
git status

# Descartar todos los cambios locales
git reset --hard origin/main

# O descartar cambios en archivos específicos
git checkout -- archivo.txt
```

### Opción 3: Rebase (Solo si sabes lo que haces)

Reorganiza los commits locales encima de los de GitHub:

```bash
git config pull.rebase true
git pull origin main
```

**⚠️ Cuidado:** Esto puede causar conflictos si hay cambios importantes.

---

## 🎯 Recomendación para Servidor

**Para un servidor de producción, usa Opción 1 (Merge):**

```bash
cd /opt/qentitv/qentitv-api

# Configurar merge como estrategia por defecto
git config pull.rebase false

# Hacer pull
git pull origin main

# Si hay conflictos, resolverlos manualmente
# Luego continuar con el despliegue
```

---

## 🔍 Ver Qué Cambió

Antes de decidir, revisa qué cambió:

```bash
# Ver cambios locales
git status

# Ver diferencias
git diff

# Ver commits locales que no están en GitHub
git log origin/main..HEAD

# Ver commits de GitHub que no tienes localmente
git log HEAD..origin/main
```

---

## 🚀 Después de Resolver

Una vez resuelto el pull:

```bash
# Continuar con el despliegue
./deploy-server.sh
```

---

## 💡 Prevenir en el Futuro

Para evitar esto, **NO edites archivos directamente en el servidor**. 

Siempre:
1. Edita en tu PC local
2. Sube a GitHub
3. En el servidor: `git pull`

Si necesitas cambiar algo en el servidor:
1. Edita el archivo
2. Haz commit: `git commit -am "Cambio en servidor"`
3. Push: `git push origin main`
4. En tu PC: `git pull`
