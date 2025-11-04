package manager

import (
	"sort"
	"sync"
	"time"

	"github.com/andresgallo/evida_backend_go/internal/geometry"
	"github.com/andresgallo/evida_backend_go/internal/models"
)

// EarthquakeManager gestiona los sismos en memoria
type EarthquakeManager struct {
	mu          sync.RWMutex
	earthquakes map[string]models.Earthquake // ID -> Earthquake
	maxAge      time.Duration                // Tiempo máximo para mantener sismos en memoria

	// Canal para notificar nuevos sismos
	newEarthquakeChan chan models.Earthquake
}

// NewEarthquakeManager crea un nuevo gestor de sismos
func NewEarthquakeManager(maxAge time.Duration) *EarthquakeManager {
	return &EarthquakeManager{
		earthquakes:       make(map[string]models.Earthquake),
		maxAge:            maxAge,
		newEarthquakeChan: make(chan models.Earthquake, 100),
	}
}

// AddEarthquake agrega un sismo al gestor
// Retorna true si es un sismo nuevo y categorizado, false si ya existía o no fue categorizado
func (em *EarthquakeManager) AddEarthquake(eq models.Earthquake) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Verificar si ya existe
	if _, exists := em.earthquakes[eq.ID]; exists {
		return false
	}

	// Categorizar el sismo
	geometry.CategorizeEarthquake(&eq)

	// Solo agregar y notificar si está categorizado
	if eq.Oceano == "" || eq.Oceano == "Uncategorized" ||
		eq.OceanoRegion == "" || eq.OceanoRegion == "Uncategorized" {
		// No agregar sismos no categorizados
		return false
	}

	// Agregar al mapa
	em.earthquakes[eq.ID] = eq

	// Notificar mediante el canal (non-blocking)
	select {
	case em.newEarthquakeChan <- eq:
	default:
		// Si el canal está lleno, no bloqueamos
	}

	return true
}

// AddEarthquakes agrega múltiples sismos y retorna los nuevos
func (em *EarthquakeManager) AddEarthquakes(earthquakes []models.Earthquake) []models.Earthquake {
	newOnes := make([]models.Earthquake, 0)

	for _, eq := range earthquakes {
		if em.AddEarthquake(eq) {
			newOnes = append(newOnes, eq)
		}
	}

	return newOnes
}

// GetAll retorna todos los sismos categorizados ordenados por tiempo (más reciente primero)
// Solo retorna sismos que tienen océano y región válidos (no "Uncategorized")
func (em *EarthquakeManager) GetAll() []models.Earthquake {
	em.mu.RLock()
	defer em.mu.RUnlock()

	earthquakes := make([]models.Earthquake, 0, len(em.earthquakes))
	for _, eq := range em.earthquakes {
		// Solo agregar sismos categorizados
		if eq.Oceano != "" && eq.Oceano != "Uncategorized" &&
			eq.OceanoRegion != "" && eq.OceanoRegion != "Uncategorized" {
			earthquakes = append(earthquakes, eq)
		}
	}

	// Ordenar por tiempo (más reciente primero)
	sort.Slice(earthquakes, func(i, j int) bool {
		return earthquakes[i].Time.After(earthquakes[j].Time)
	})

	return earthquakes
}

// GetByOceano retorna sismos filtrados por océano, ordenados por tiempo
func (em *EarthquakeManager) GetByOceano(oceano string) []models.Earthquake {
	em.mu.RLock()
	defer em.mu.RUnlock()

	earthquakes := make([]models.Earthquake, 0)
	for _, eq := range em.earthquakes {
		if eq.Oceano == oceano {
			earthquakes = append(earthquakes, eq)
		}
	}

	// Ordenar por tiempo (más reciente primero)
	sort.Slice(earthquakes, func(i, j int) bool {
		return earthquakes[i].Time.After(earthquakes[j].Time)
	})

	return earthquakes
}

// GetByRegion retorna sismos filtrados por región, ordenados por tiempo
func (em *EarthquakeManager) GetByRegion(region string) []models.Earthquake {
	em.mu.RLock()
	defer em.mu.RUnlock()

	earthquakes := make([]models.Earthquake, 0)
	for _, eq := range em.earthquakes {
		if eq.OceanoRegion == region {
			earthquakes = append(earthquakes, eq)
		}
	}

	// Ordenar por tiempo (más reciente primero)
	sort.Slice(earthquakes, func(i, j int) bool {
		return earthquakes[i].Time.After(earthquakes[j].Time)
	})

	return earthquakes
}

// GetByTimeRange retorna sismos en un rango de tiempo
func (em *EarthquakeManager) GetByTimeRange(start, end time.Time) []models.Earthquake {
	em.mu.RLock()
	defer em.mu.RUnlock()

	earthquakes := make([]models.Earthquake, 0)
	for _, eq := range em.earthquakes {
		if eq.Time.After(start) && eq.Time.Before(end) {
			earthquakes = append(earthquakes, eq)
		}
	}

	// Ordenar por tiempo (más reciente primero)
	sort.Slice(earthquakes, func(i, j int) bool {
		return earthquakes[i].Time.After(earthquakes[j].Time)
	})

	return earthquakes
}

// GetCount retorna el número total de sismos categorizados
func (em *EarthquakeManager) GetCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	count := 0
	for _, eq := range em.earthquakes {
		if eq.Oceano != "" && eq.Oceano != "Uncategorized" &&
			eq.OceanoRegion != "" && eq.OceanoRegion != "Uncategorized" {
			count++
		}
	}
	return count
}

// CleanOld elimina sismos más antiguos que maxAge
func (em *EarthquakeManager) CleanOld() int {
	em.mu.Lock()
	defer em.mu.Unlock()

	cutoff := time.Now().Add(-em.maxAge)
	removed := 0

	for id, eq := range em.earthquakes {
		if eq.Time.Before(cutoff) {
			delete(em.earthquakes, id)
			removed++
		}
	}

	return removed
}

// StartCleanup inicia una goroutine que limpia sismos antiguos periódicamente
func (em *EarthquakeManager) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			removed := em.CleanOld()
			if removed > 0 {
				// Log
			}
		}
	}()
}

// GetNewEarthquakeChannel retorna el canal para recibir notificaciones de nuevos sismos
func (em *EarthquakeManager) GetNewEarthquakeChannel() <-chan models.Earthquake {
	return em.newEarthquakeChan
}

// GetStats retorna estadísticas de los sismos
func (em *EarthquakeManager) GetStats() map[string]interface{} {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total"] = len(em.earthquakes)

	// Contar por océano
	byOceano := make(map[string]int)
	for _, eq := range em.earthquakes {
		byOceano[eq.Oceano]++
	}
	stats["by_oceano"] = byOceano

	// Contar por región
	byRegion := make(map[string]int)
	for _, eq := range em.earthquakes {
		byRegion[eq.OceanoRegion]++
	}
	stats["by_region"] = byRegion

	// Contar por fuente
	bySource := make(map[string]int)
	for _, eq := range em.earthquakes {
		bySource[eq.Source]++
	}
	stats["by_source"] = bySource

	return stats
}
