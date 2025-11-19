# 🌍 EVIDA Backend - Sistema de Monitoreo de Sismos

## ✅ Proyecto Completado

He creado exitosamente una aplicación en **Go** que:

### 🎯 Funcionalidades Principales

1. **Extracción de datos en tiempo real** desde tres fuentes:
   - **USGS**: Sismos M >= 4.5 de la última semana
   - **GEOFON**: Últimos 50 eventos
   - **SGC**: Eventos de los últimos 5 días

2. **Categorización geográfica avanzada**:
   - Usa el algoritmo **Point-in-Polygon** (Ray Casting)
   - Determina el **océano** (Pacífico / Caribe)
   - Determina la **región** (local / regional / lejano)
   - Utiliza los polígonos del archivo `datosLC.json` (698K puntos para Pacífico)

3. **Sistema en tiempo real**:
   - Corre continuamente como servicio
   - Actualiza datos cada 2 minutos
   - Detecta sismos nuevos automáticamente
   - **Detecta actualizaciones de parámetros** (magnitud, profundidad, ubicación, etc.)
   - Notifica mediante **WebSocket** a clientes conectados
   - Distingue entre sismos nuevos (🔔) y actualizados (🔄)

4. **API REST completa**:
   - `GET /api/earthquakes` - Todos los sismos
   - `GET /api/earthquakes?oceano=Pacifico` - Filtrar por océano
   - `GET /api/earthquakes?region=local` - Filtrar por región
   - `GET /api/stats` - Estadísticas
   - `GET /api/health` - Health check

5. **Gestión en memoria**:
   - Thread-safe con `sync.RWMutex`
   - Detección automática de duplicados
   - **Detección de cambios en sismos existentes**
   - Limpieza automática de sismos antiguos (>7 días)
   - Lista siempre ordenada por tiempo (más reciente primero)

### 📁 Estructura del Proyecto

```
evida_backend_go/
├── cmd/
│   └── server/
│       └── main.go              # Punto de entrada, orquestación
├── internal/
│   ├── api/
│   │   └── server.go            # Servidor HTTP/WebSocket
│   ├── fetcher/
│   │   ├── fetcher.go          # Interfaz común
│   │   ├── usgs.go             # Cliente USGS
│   │   ├── geofon.go           # Cliente GEOFON
│   │   └── sgc.go              # Cliente SGC
│   ├── geometry/
│   │   ├── polygon.go          # Algoritmo Point-in-Polygon
│   │   ├── regions_data.go     # Carga de polígonos
│   │   ├── regions.go          # (Legacy)
│   │   ├── datosLC.json        # Polígonos geográficos (1.2M líneas)
│   │   └── polygon_test.go     # Tests unitarios
│   ├── manager/
│   │   └── earthquake_manager.go  # Gestor thread-safe
│   ├── models/
│   │   └── earthquake.go       # Modelos de datos
│   └── websocket/
│       └── hub.go              # Hub de WebSocket
├── web/
│   └── index.html              # Interfaz web de prueba
├── go.mod                      # Dependencias
├── Dockerfile                  # Docker
├── Makefile                    # Tareas comunes
└── README.md                   # Documentación

### 🚀 Cómo Usar

#### Iniciar el servidor:

```bash
# Método 1: Ejecutar directamente
go run cmd/server/main.go

# Método 2: Compilar y ejecutar
go build -o bin/evida-server cmd/server/main.go
./bin/evida-server

# Método 3: Usar Makefile
make run
```

#### Probar las APIs:

```bash
# Ver todos los sismos
curl http://localhost:8080/api/earthquakes

# Filtrar por Pacífico
curl http://localhost:8080/api/earthquakes?oceano=Pacifico

# Filtrar por región local
curl http://localhost:8080/api/earthquakes?region=local

# Ver estadísticas
curl http://localhost:8080/api/stats
```

#### Conectar WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === 'new_earthquake') {
        console.log('Nuevo sismo:', message.data);
        // {
        //   id: "us7000xxx",
        //   magnitud: 5.2,
        //   place: "...",
        //   closerTowns: "...",
        //   latitud: 4.5,
        //   longitud: -75.2,
        //   profundidad: 10.5,
        //   localTime: "2025-11-03 07:34:56",
        //   fuente: "USGS",
        //   oceano: "Pacifico",
        //   oceanoRegion: "local",
        //   url: "https://earthquake.usgs.gov/..."
        // }
    }
};
```

**Detección de Actualizaciones:**

El sistema monitorea cambios en los siguientes parámetros de sismos existentes:
- `magnitud` - Las agencias frecuentemente revisan las magnitudes
- `place` y `closerTowns` - Actualizaciones en la descripción de ubicación
- `latitud` y `longitud` - Refinamiento de coordenadas
- `profundidad` - Ajustes en la profundidad del hipocentro
- `oceano` y `oceanoRegion` - Recategorización si cambian las coordenadas
- `localTime` - Corrección del tiempo del evento

Cuando se detecta cualquier cambio, el sistema:
1. Actualiza el sismo en memoria
2. Envía notificación por WebSocket con los datos actualizados
3. Registra en logs: `🔄 Sismo actualizado: M5.2 - ...`

### 🔧 Configuración

#### URLs de Fuentes de Datos:

En los archivos de `internal/fetcher/*.go`:
- SGC: `http://archive.sgc.gov.co/feed/v1.0/summary/five_days_all.json`
- USGS: `https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/4.5_week.geojson`
- GEOFON: `https://geofon.gfz.de/eqinfo/list.php?fmt=rss&nmax=50`

#### Parámetros Configurables:

En `cmd/server/main.go`:

```go
const (
    fetchInterval    = 2 * time.Minute       // Frecuencia de actualización
    maxEarthquakeAge = 7 * 24 * time.Hour   // Tiempo máximo en memoria
    cleanupInterval  = 1 * time.Hour        // Frecuencia de limpieza
    serverPort       = ":8080"              // Puerto del servidor
)
```

### 📊 Estado Actual

**El servidor está funcionando correctamente:**

```
✅ Datos de regiones cargados correctamente
   - Pacífico CP: 698,123 puntos
   - Pacífico Local: 75,503 puntos
   - Caribe CC: 5 polígonos
   - Caribe Regional: 5 polígonos

✅ Tres fuentes de datos integradas y funcionando:
   - USGS: ~90 sismos M >= 4.5
   - GEOFON: ~20 sismos RSS feed
   - SGC: ~350 sismos últimos 5 días

✅ Sistema de actualización implementado:
   - Detecta cambios en parámetros de sismos
   - Notifica actualizaciones por WebSocket
   - Logs diferencian nuevos (🔔) vs actualizados (🔄)

✅ Categorización funcionando:
   - Pacifico local: Sismos cercanos a Colombia
   - Pacifico regional: Costa Rica, Ecuador, Panamá
   - Pacifico lejano: Resto del Pacífico
   - Caribe local: Mar Caribe cerca de Colombia
   - Caribe regional: Venezuela, República Dominicana, etc.
   - Caribe lejano: Atlántico Norte
```

### 🎨 Interfaz Web

Abre `http://localhost:8080` en tu navegador para ver la interfaz de prueba (necesitarás servir el archivo `web/index.html` con un servidor estático).

### 🐳 Docker

```bash
# Construir imagen
docker build -t evida-backend:latest .

# Ejecutar contenedor
docker run -p 8080:8080 evida-backend:latest
```

### 📝 Notas Importantes

1. **USGS**: ✅ Funcionando perfectamente con GeoJSON
2. **GEOFON**: ✅ Parser de RSS implementado y funcionando
3. **SGC**: ✅ Parser de JSON implementado con manejo de `closerTowns`
4. **Campos en español**: Todos los campos principales están en español (magnitud, place, closerTowns, latitud, longitud, profundidad, fuente)
5. **Zona horaria**: Todos los tiempos se convierten a UTC-5 (Bogotá) con formato `localTime`
6. **URLs de mapas SGC**: Formato `https://archive.sgc.gov.co/events/{ID}/map.gif`
7. **Sistema de actualizaciones**: Detecta y notifica cambios en cualquier parámetro de sismos existentes

### 🔄 Mejoras Recientes

- ✅ **Sistema de detección de actualizaciones** (19/11/2025)
  - Compara todos los parámetros importantes al recibir datos
  - Envía notificaciones diferenciadas por WebSocket
  - Logs distinguen entre sismos nuevos y actualizados
  - Thread-safe con manejo de canales non-blocking

### 🔄 Próximos Pasos (Opcionales)

- [ ] Agregar persistencia en base de datos (PostgreSQL/MongoDB)
- [ ] Implementar caché de polígonos para mejorar rendimiento
- [ ] Agregar autenticación para el WebSocket
- [ ] Implementar rate limiting
- [ ] Agregar métricas con Prometheus
- [ ] Crear cliente de línea de comandos
- [ ] Implementar filtros adicionales (magnitud mínima, fecha)
- [ ] Historial de cambios de sismos (tracking de revisiones)
- [ ] Panel de administración web
- [ ] Notificaciones push a dispositivos móviles

### 📜 Licencia

MIT

---

**Desarrollado con ❤️ usando Go 1.21**
