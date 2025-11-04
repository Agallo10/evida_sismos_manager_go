package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andresgallo/evida_backend_go/internal/models"
)

// USGSFetcher extrae datos de sismos de USGS
type USGSFetcher struct {
	client *http.Client
}

// NewUSGSFetcher crea una nueva instancia del fetcher de USGS
func NewUSGSFetcher() *USGSFetcher {
	return &USGSFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// USGSResponse representa la respuesta de la API de USGS
type USGSResponse struct {
	Features []struct {
		ID         string `json:"id"`
		Properties struct {
			Mag    float64 `json:"mag"`
			Place  string  `json:"place"`
			Time   int64   `json:"time"` // milisegundos desde epoch
			URL    string  `json:"url"`
			Detail string  `json:"detail"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` // [lon, lat, depth]
		} `json:"geometry"`
	} `json:"features"`
}

// Fetch obtiene los sismos recientes de USGS
// Retorna sismos de la última semana con magnitud >= 4.5
func (f *USGSFetcher) Fetch() ([]models.Earthquake, error) {
	// API de USGS: sismos de la última semana, magnitud >= 4.5
	url := "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/4.5_week.geojson"

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching USGS data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("USGS API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading USGS response: %w", err)
	}

	var usgsResp USGSResponse
	if err := json.Unmarshal(body, &usgsResp); err != nil {
		return nil, fmt.Errorf("error parsing USGS JSON: %w", err)
	}

	earthquakes := make([]models.Earthquake, 0, len(usgsResp.Features))
	for _, feature := range usgsResp.Features {
		if len(feature.Geometry.Coordinates) < 3 {
			continue
		}

		eq := models.Earthquake{
			ID:        feature.ID,
			Magnitude: feature.Properties.Mag,
			Location:  feature.Properties.Place,
			Longitude: feature.Geometry.Coordinates[0],
			Latitude:  feature.Geometry.Coordinates[1],
			Depth:     feature.Geometry.Coordinates[2],
			Time:      time.UnixMilli(feature.Properties.Time),
			Source:    "USGS",
			URL:       feature.Properties.URL,
		}

		earthquakes = append(earthquakes, eq)
	}

	return earthquakes, nil
}
