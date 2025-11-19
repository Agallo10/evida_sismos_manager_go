# EVIDA Backend - Sistema de Monitoreo de Sismos

Sistema en tiempo real que extrae información de sismos de múltiples fuentes (USGS, GEOFON, SGC), los categoriza geográficamente y notifica a clientes mediante WebSocket.

## Características

- ✅ Extracción continua de datos de sismos desde:
  - **USGS** (United States Geological Survey) - Magnitud >= 4.5, última semana
  - **GEOFON** (GFZ German Research Centre for Geosciences) - Últimos 50 eventos
  - **SGC** (Servicio Geológico Colombiano) - Últimos 5 días
- ✅ Categorización geográfica mediante algoritmo Point-in-Polygon
- ✅ Clasificación por océano (Pacífico, Caribe) y región (local, regional, lejano)
- ✅ **Detección automática de actualizaciones** - Detecta cambios en parámetros de sismos existentes
- ✅ Notificaciones en tiempo real vía WebSocket para sismos nuevos y actualizados
- ✅ API REST para consultar sismos
- ✅ Almacenamiento en memoria con gestión thread-safe
- ✅ Lista ordenada por tiempo
- ✅ Limpieza automática de sismos antiguos (>7 días)

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

Recibirás notificaciones JSON cuando lleguen sismos nuevos o se actualicen:

**Sismo Nuevo:**
```json
{
  "type": "new_earthquake",
  "data": {
    "id": "us7000example",
    "magnitud": 5.2,
    "place": "10 km S of Example City",
    "closerTowns": "Example City",
    "latitud": 4.5,
    "longitud": -75.2,
    "profundidad": 10.5,
    "localTime": "2025-11-03 07:34:56",
    "fuente": "USGS",
    "oceano": "Pacifico",
    "oceanoRegion": "local",
    "url": "https://earthquake.usgs.gov/earthquakes/eventpage/us7000example"
  }
}
```

**Sismo Actualizado:**
Los sismos pueden actualizarse cuando las agencias revisan parámetros como magnitud, profundidad o ubicación. El sistema detecta automáticamente estos cambios comparando:
- Magnitud
- Ubicación (place y closerTowns)
- Coordenadas (latitud, longitud)
- Profundidad
- Categorización (océano y región)
- Tiempo del evento

Cuando se detecta un cambio, se envía la misma estructura JSON con los valores actualizados.

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
