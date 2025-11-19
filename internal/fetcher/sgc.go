package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andresgallo/evida_backend_go/internal/models"
)

// SGCFetcher extrae datos de sismos del Servicio Geológico Colombiano
type SGCFetcher struct {
	client *http.Client
}

// NewSGCFetcher crea una nueva instancia del fetcher de SGC
func NewSGCFetcher() *SGCFetcher {
	return &SGCFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SGCResponse representa la respuesta de la API de SGC
type SGCResponse struct {
	Features []struct {
		Type     string `json:"type"`
		ID       string `json:"id"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"` // [lon, lat, depth]
		} `json:"geometry"`
		Properties struct {
			Mag         float64     `json:"mag"`
			Place       string      `json:"place"`
			CloserTowns string      `json:"closerTowns"` // Pueblos cercanos
			Time        int64       `json:"time"`        // milisegundos (puede ser null)
			UTCTime     string      `json:"utcTime"`     // formato: "2025-11-04 02:42"
			LocalTime   string      `json:"localTime"`   // formato: "2025-11-03 21:42"
			Updated     interface{} `json:"updated"`     // puede ser string o int64
			TZ          int         `json:"tz"`
			URL         string      `json:"url"`
			Detail      string      `json:"detail"`
			Felt        int         `json:"felt"`
			CDI         float64     `json:"cdi"`
			MMI         float64     `json:"mmi"`
			Alert       string      `json:"alert"`
			Status      string      `json:"status"`
			Tsunami     int         `json:"tsunami"`
			Sig         int         `json:"sig"`
			Net         string      `json:"net"`
			Code        string      `json:"code"`
			IDS         string      `json:"ids"`
			Sources     string      `json:"sources"`
			Types       string      `json:"types"`
			NST         int         `json:"nst"`
			Dmin        float64     `json:"dmin"`
			RMS         float64     `json:"rms"`
			Gap         float64     `json:"gap"`
			MagType     string      `json:"magType"`
			Type        string      `json:"type"`
			Title       string      `json:"title"`
		} `json:"properties"`
	} `json:"features"`
}

// Fetch obtiene los sismos recientes del SGC
// Retorna sismos de los últimos 5 días
func (f *SGCFetcher) Fetch() ([]models.Earthquake, error) {
	// API del SGC: sismos de los últimos 5 días en formato GeoJSON
	url := "http://archive.sgc.gov.co/feed/v1.0/summary/five_days_all.json"

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching SGC data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SGC API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading SGC response: %w", err)
	}

	var sgcResp SGCResponse
	if err := json.Unmarshal(body, &sgcResp); err != nil {
		return nil, fmt.Errorf("error parsing SGC JSON: %w", err)
	}

	earthquakes := make([]models.Earthquake, 0, len(sgcResp.Features))
	for _, feature := range sgcResp.Features {
		if len(feature.Geometry.Coordinates) < 3 {
			continue
		}

		// Zona horaria de Bogotá (UTC-5)
		bogotaLocation := time.FixedZone("America/Bogota", -5*60*60)

		// Parsear el tiempo desde utcTime (formato: "2025-11-04 02:42")
		var eqTime time.Time
		if feature.Properties.UTCTime != "" {
			// Intentar parsear con segundos
			if t, err := time.Parse("2006-01-02 15:04:05", feature.Properties.UTCTime); err == nil {
				// Convertir UTC a hora local de Bogotá
				eqTime = t.In(bogotaLocation)
			} else if t, err := time.Parse("2006-01-02 15:04", feature.Properties.UTCTime); err == nil {
				// Formato sin segundos - convertir UTC a hora local de Bogotá
				eqTime = t.In(bogotaLocation)
			}
		} else if feature.Properties.Time > 0 {
			// Fallback a time en milisegundos si existe - convertir a hora local
			utcTime := time.UnixMilli(feature.Properties.Time)
			eqTime = utcTime.In(bogotaLocation)
		} else {
			// Si no hay tiempo válido, usar el tiempo actual en zona horaria de Bogotá
			eqTime = time.Now().In(bogotaLocation)
		}

		// Construir URL del mapa usando el ID del sismo
		mapURL := fmt.Sprintf("https://archive.sgc.gov.co/events/%s/map.gif", feature.ID)

		// Determinar closerTowns: si existe y no es "None", usar closerTowns, sino usar place
		closerTowns := feature.Properties.CloserTowns
		if closerTowns == "" || closerTowns == "None" {
			closerTowns = feature.Properties.Place
		}

		eq := models.Earthquake{
			ID:          feature.ID,
			Magnitud:    feature.Properties.Mag,
			Place:       feature.Properties.Place,
			CloserTowns: closerTowns,
			Longitud:    feature.Geometry.Coordinates[1],
			Latitud:     feature.Geometry.Coordinates[0],
			Profundidad: feature.Geometry.Coordinates[2],
			Time:        eqTime,
			Fuente:      "SGC",
			URL:         mapURL,
		}

		earthquakes = append(earthquakes, eq)
	}

	return earthquakes, nil
}

// FetchMock retorna datos de ejemplo del SGC para pruebas
// Úsalo mientras configuras la integración real con SGC
func (f *SGCFetcher) FetchMock() []models.Earthquake {
	bogotaLocation := time.FixedZone("America/Bogota", -5*60*60)
	return []models.Earthquake{
		{
			ID:          "sgc-" + time.Now().Format("20060102150405"),
			Magnitud:    3.5,
			Place:       "15 km al norte de Bogotá",
			CloserTowns: "15 km al norte de Bogotá",
			Latitud:     4.8,
			Longitud:    -74.0,
			Profundidad: 10.0,
			Time:        time.Now().In(bogotaLocation).Add(-1 * time.Hour),
			Fuente:      "SGC",
			URL:         "https://www.sgc.gov.co",
		},
	}
}
