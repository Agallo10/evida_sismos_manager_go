package geometry

import (
	"log"
	"math"

	"github.com/andresgallo/evida_backend_go/internal/models"
)

// PointInPolygon determina si un punto está dentro de un polígono usando el algoritmo Ray Casting
// Basado en: https://observablehq.com/@tmcw/understanding-point-in-polygon
func PointInPolygon(point models.Point, polygon models.Polygon) bool {
	if len(polygon) < 3 {
		return false
	}

	// Contador de intersecciones
	intersections := 0
	n := len(polygon)

	for i := 0; i < n; i++ {
		// Obtener el segmento actual y el siguiente (cerrando el polígono)
		p1 := polygon[i]
		p2 := polygon[(i+1)%n]

		// Verificar si el rayo horizontal desde el punto intersecta el segmento
		if rayIntersectsSegment(point, p1, p2) {
			intersections++
		}
	}

	// Si hay un número impar de intersecciones, el punto está dentro
	return intersections%2 == 1
}

// rayIntersectsSegment verifica si un rayo horizontal desde el punto intersecta el segmento
func rayIntersectsSegment(point, p1, p2 models.Point) bool {
	// El rayo va hacia la derecha desde el punto (incrementando longitud)

	// Verificar si el segmento está por encima o por debajo del punto
	if (p1.Lat > point.Lat) == (p2.Lat > point.Lat) {
		return false
	}

	// Calcular la longitud de intersección del rayo con el segmento
	// Fórmula: x = x1 + (y - y1) * (x2 - x1) / (y2 - y1)
	slope := (p2.Lon - p1.Lon) / (p2.Lat - p1.Lat)
	intersectionLon := p1.Lon + (point.Lat-p1.Lat)*slope

	// Si la intersección está a la derecha del punto, cuenta
	return point.Lon < intersectionLon
}

// CategorizeEarthquake asigna océano y región a un sismo basándose en su ubicación
func CategorizeEarthquake(eq *models.Earthquake) {
	if regionData == nil {
		log.Printf("⚠️  Datos de regiones no cargados")
		return
	}

	point := models.Point{
		Lat: eq.Latitude,
		Lon: eq.Longitude,
	}

	// Determinar región del océano Pacífico con subregión
	if PointInPolygon(point, regionData.LatlonCPWorld) {
		eq.Oceano = "Pacifico"
		eq.OceanoRegion = determinarRegionPacifico(point)
		return
	}

	// Pacífico Local
	if PointInPolygon(point, regionData.LatlonPacificoLocal) {
		eq.Oceano = "Pacifico"
		eq.OceanoRegion = "local"
		return
	}

	// Pacífico Regional
	if len(regionData.LatlonPacificoRegional) > 0 && PointInPolygon(point, regionData.LatlonPacificoRegional) {
		eq.Oceano = "Pacifico"
		eq.OceanoRegion = "regional"
		return
	}

	// Pacífico Local 20Km
	if len(regionData.LatlonPacificoLocal20Km) > 0 && PointInPolygon(point, regionData.LatlonPacificoLocal20Km) {
		eq.Oceano = "Pacifico"
		eq.OceanoRegion = "local"
		return
	}

	// Caribe Lejano
	if evaluarPuntosMultiples(point, regionData.LatlonCCWorld) {
		eq.Oceano = "Caribe"
		eq.OceanoRegion = "lejano"
		return
	}

	// Caribe Regional
	if evaluarPuntosMultiples(point, regionData.LatlonCaribeRegional) {
		eq.Oceano = "Caribe"
		eq.OceanoRegion = "regional"
		return
	}

	// Caribe Local o Insular
	if PointInPolygon(point, regionData.LatlonCaribeLocal) ||
		PointInPolygon(point, regionData.LatlonCaribeLocalInsular) {
		eq.Oceano = "Caribe"
		eq.OceanoRegion = "local"
		return
	}

	// No categorizado
	eq.Oceano = "Uncategorized"
	eq.OceanoRegion = "Uncategorized"
}

// determinarRegionPacifico determina la subregión dentro del Pacífico CP
func determinarRegionPacifico(point models.Point) string {
	if PointInPolygon(point, regionData.LatlonPacificoLocal) {
		return "local"
	}
	if PointInPolygon(point, regionData.LatlonPacificoRegional) {
		return "regional"
	}
	return "lejano"
}

// evaluarPuntosMultiples evalúa si un punto está en alguno de los polígonos de una lista
func evaluarPuntosMultiples(point models.Point, polygons []models.Polygon) bool {
	for _, polygon := range polygons {
		if PointInPolygon(point, polygon) {
			return true
		}
	}
	return false
}

// Distance calcula la distancia en kilómetros entre dos puntos usando la fórmula de Haversine
func Distance(p1, p2 models.Point) float64 {
	const earthRadius = 6371.0 // Radio de la Tierra en km

	lat1 := p1.Lat * math.Pi / 180
	lat2 := p2.Lat * math.Pi / 180
	deltaLat := (p2.Lat - p1.Lat) * math.Pi / 180
	deltaLon := (p2.Lon - p1.Lon) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
