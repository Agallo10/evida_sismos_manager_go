# EVIDA Backend - Sistema de Monitoreo de Sismos

Sistema en tiempo real que extrae información de sismos de múltiples fuentes (USGS, GEOFON, SGC), los categoriza geográficamente y notifica a clientes mediante WebSocket.

## Características

- ✅ Extracción continua de datos de sismos desde:
  - **USGS** (United States Geological Survey) - Magnitud >= 4.5, última semana
  - **GEOFON** (GFZ German Research Centre for Geosciences) - Últimos 50 eventos
  - **SGC** (Servicio Geológico Colombiano) - Últimos 5 días
- ✅ Categorización geográfica mediante algoritmo Point-in-Polygon
- ✅ Clasificación por océano (Pacífico, Caribe) y región (local, regional, lejano)
- ✅ Notificaciones en tiempo real vía WebSocket
- ✅ API REST para consultar sismos
- ✅ Almacenamiento en memoria (sin base de datos)
- ✅ Lista ordenada por tiempo

## Instalación

```bash
# Clonar el repositorio
git clone <repo-url>
cd evida_backend_go

# Instalar dependencias
go mod download

# Ejecutar
go run cmd/server/main.go
```

## Uso

### WebSocket
Conectarse a: `ws://localhost:8080/ws`

Recibirás notificaciones JSON cuando lleguen sismos nuevos:
```json
{
  "type": "new_earthquake",
  "data": {
    "id": "us7000example",
    "magnitude": 5.2,
    "location": "10 km S of Example City",
    "latitude": 4.5,
    "longitude": -75.2,
    "depth": 10.5,
    "time": "2025-11-03T12:34:56Z",
    "source": "USGS",
    "oceano": "Pacifico",
    "oceanoRegion": "local"
  }
}
```

### API REST

#### Obtener todos los sismos
```bash
GET http://localhost:8080/api/earthquakes
```

#### Obtener sismos por océano
```bash
GET http://localhost:8080/api/earthquakes?oceano=Pacifico
GET http://localhost:8080/api/earthquakes?oceano=Caribe
```

#### Obtener sismos por región
```bash
GET http://localhost:8080/api/earthquakes?region=local
GET http://localhost:8080/api/earthquakes?region=regional
GET http://localhost:8080/api/earthquakes?region=lejano
```

#### Obtener estadísticas
```bash
GET http://localhost:8080/api/stats
```

#### Health check
```bash
GET http://localhost:8080/api/health
```

## Arquitectura

```
cmd/
  server/
    main.go           # Punto de entrada
internal/
  fetcher/            # Clientes para extraer datos
    usgs.go
    geofon.go
    sgc.go
  geometry/           # Algoritmo point-in-polygon
    polygon.go
  manager/            # Gestor de sismos en memoria
    earthquake_manager.go
  models/             # Estructuras de datos
    earthquake.go
  websocket/          # Servidor WebSocket
    hub.go
    client.go
```

## Configuración

### Fuentes de Datos

El sistema extrae datos de las siguientes URLs:

- **SGC**: `http://archive.sgc.gov.co/feed/v1.0/summary/five_days_all.json`
- **USGS**: `https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/4.5_week.geojson`
- **GEOFON**: `https://geofon.gfz.de/eqinfo/list.php?fmt=rss&nmax=50`

### Regiones Geográficas

Las regiones se configuran en `internal/geometry/datosLC.json`, que incluye polígonos para:

- **Pacífico**: 
  - CP World (Cinturón de Fuego del Pacífico)
  - Local
  - Regional
  - Local 20km
  
- **Caribe**:
  - CC World (Caribe Completo)
  - Regional
  - Local
  - Local Insular

### Parámetros Configurables

En `cmd/server/main.go`:

```go
// Intervalo de actualización de datos (cada 2 minutos)
fetchInterval = 2 * time.Minute

// Tiempo máximo para mantener sismos en memoria (7 días)
maxEarthquakeAge = 7 * 24 * time.Hour

// Intervalo de limpieza de sismos antiguos (cada hora)
cleanupInterval = 1 * time.Hour

// Puerto del servidor
serverPort = ":8080"
```

## Licencia

MIT
