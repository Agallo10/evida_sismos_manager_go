# 📊 Configuración de Datos Geográficos

Este proyecto utiliza archivos de polígonos geográficos grandes para la categorización de sismos.

## 📁 Archivos Requeridos

Los siguientes archivos deben estar en `internal/geometry/`:

1. **datosLC.json** (~64MB)
   - Contiene polígonos para: Pacífico CP, Caribe, regiones locales
   
2. **latlonCPWorldEste.json** (~49MB)
   - Contiene el polígono del Pacífico CP Este (oeste del Pacífico)
   - Sismos en Japón, Indonesia, Papua Nueva Guinea, etc.

## 🚀 Opción 1: Git LFS (Recomendado)

Si el repositorio usa Git LFS, los archivos se descargarán automáticamente:

```bash
# Instalar Git LFS (si no lo tienes)
brew install git-lfs  # macOS
# o: sudo apt-get install git-lfs  # Linux

# Inicializar Git LFS
git lfs install

# Clonar el repositorio (descargará automáticamente los archivos grandes)
git clone https://github.com/Agallo10/evida_sismos_manager_go.git
```

## 📥 Opción 2: Descarga Manual

Si los archivos no están en el repositorio, contacta al administrador para obtenerlos.

Los archivos deben colocarse en:
```
internal/geometry/datosLC.json
internal/geometry/latlonCPWorldEste.json
```

## ✅ Verificación

Para verificar que los archivos están correctamente instalados:

```bash
# Verificar que existen
ls -lh internal/geometry/datosLC.json
ls -lh internal/geometry/latlonCPWorldEste.json

# Ejecutar el servidor
go run cmd/server/main.go
```

Si todo está correcto, verás en los logs:
```
✅ Datos de regiones cargados correctamente
   - Pacífico CP: 698123 puntos
   - Pacífico CP Este: 698123 puntos
   - Pacífico Local: 75503 puntos
   ...
```

## 🔧 Reducción de Tamaño

El archivo `datosLC.json` se redujo de 138MB a 64MB extrayendo `latlonCPWorldEste` a su propio archivo.
El sistema carga ambos archivos automáticamente al iniciar.

## ⚠️ Nota sobre Git

GitHub tiene límites de tamaño de archivo:
- **Advertencia:** archivos > 50MB
- **Error:** archivos > 100MB

Por eso recomendamos usar Git LFS para estos archivos grandes.
