# 🔒 Solución: Permission Denied

## 🔴 Error

```
-bash: ./verificar-puerto.sh: Permission denied
```

Esto significa que el archivo no tiene permisos de ejecución.

---

## ✅ Solución Rápida

### Dar Permisos de Ejecución

```bash
chmod +x verificar-puerto.sh
./verificar-puerto.sh
```

### O Ejecutar con Bash Directamente

```bash
bash verificar-puerto.sh
```

---

## 📋 Para Todos los Scripts

Si tienes problemas con otros scripts también:

```bash
# Dar permisos a todos los scripts .sh
chmod +x *.sh

# Verificar
ls -la *.sh
```

Deberías ver algo como:
```
-rwxr-xr-x 1 root root 1234 verificar-puerto.sh
```
La `x` significa que tiene permisos de ejecución.

---

## 🚀 Scripts que Necesitan Permisos

```bash
chmod +x verificar-puerto.sh
chmod +x deploy-server.sh
chmod +x actualizar-api.sh
```

---

## ✅ Verificar Permisos

```bash
ls -la *.sh
```

Si no tienen `x` en los permisos, dales permisos con `chmod +x`.

---

**¡Listo! Ahora puedes ejecutar los scripts** 🚀
