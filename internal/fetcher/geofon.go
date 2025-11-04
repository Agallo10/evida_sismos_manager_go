package fetcher

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andresgallo/evida_backend_go/internal/models"
)

// GEOFONFetcher extrae datos de sismos de GEOFON
type GEOFONFetcher struct {
	client *http.Client
}

// NewGEOFONFetcher crea una nueva instancia del fetcher de GEOFON
func NewGEOFONFetcher() *GEOFONFetcher {
	return &GEOFONFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GEOFONItem representa un item en el feed RSS de GEOFON
type GEOFONItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// GEOFONChannel representa el canal RSS de GEOFON
type GEOFONChannel struct {
	Items []GEOFONItem `xml:"item"`
}

// GEOFONFeed representa el feed RSS de GEOFON
type GEOFONFeed struct {
	XMLName xml.Name      `xml:"rss"`
	Channel GEOFONChannel `xml:"channel"`
}

// Fetch obtiene los sismos recientes de GEOFON
func (f *GEOFONFetcher) Fetch() ([]models.Earthquake, error) {
	// Feed RSS de GEOFON con los últimos 50 sismos
	url := "https://geofon.gfz.de/eqinfo/list.php?fmt=rss&nmax=50"

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching GEOFON data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GEOFON API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading GEOFON response: %w", err)
	}

	var feed GEOFONFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("error parsing GEOFON XML: %w", err)
	}

	earthquakes := make([]models.Earthquake, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		eq, err := parseGEOFONItem(item)
		if err != nil {
			// Log y continuar con el siguiente
			continue
		}
		earthquakes = append(earthquakes, eq)
	}

	return earthquakes, nil
}

// parseGEOFONItem convierte un item RSS de GEOFON a un Earthquake
func parseGEOFONItem(item GEOFONItem) (models.Earthquake, error) {
	// El título tiene formato: "M 5.2, NEAR COAST OF CENTRAL CHILE"
	// La descripción tiene formato: "2025-11-03 22:30:52  52.26   160.25    10 km    A"
	// Formato: FECHA HORA  LATITUD  LONGITUD  PROFUNDIDAD  TIPO

	eq := models.Earthquake{
		ID:       item.GUID,
		Location: item.Title,
		Source:   "GEOFON",
		URL:      item.Link,
	}

	// Parsear magnitud y ubicación del título
	if strings.HasPrefix(item.Title, "M ") {
		parts := strings.SplitN(item.Title, ",", 2)
		if len(parts) == 2 {
			magStr := strings.TrimSpace(strings.TrimPrefix(parts[0], "M "))
			if mag, err := strconv.ParseFloat(magStr, 64); err == nil {
				eq.Magnitude = mag
			}
			eq.Location = strings.TrimSpace(parts[1])
		}
	}

	// Parsear la descripción: "2025-11-03 22:30:52  52.26   160.25    10 km    A"
	description := strings.TrimSpace(item.Description)
	fields := strings.Fields(description) // Divide por espacios en blanco

	if len(fields) >= 5 {
		// fields[0] = fecha (2025-11-03)
		// fields[1] = hora (22:30:52)
		// fields[2] = latitud (52.26)
		// fields[3] = longitud (160.25)
		// fields[4] = profundidad (10)
		// fields[5] = "km"
		// fields[6] = tipo ("A", "M", "C")

		// Parsear fecha y hora
		dateTimeStr := fields[0] + " " + fields[1]
		if t, err := time.Parse("2006-01-02 15:04:05", dateTimeStr); err == nil {
			eq.Time = t
		}

		// Parsear latitud
		if lat, err := strconv.ParseFloat(fields[2], 64); err == nil {
			eq.Latitude = lat
		}

		// Parsear longitud
		if lon, err := strconv.ParseFloat(fields[3], 64); err == nil {
			eq.Longitude = lon
		}

		// Parsear profundidad
		if depth, err := strconv.ParseFloat(fields[4], 64); err == nil {
			eq.Depth = depth
		}
	}

	return eq, nil
}

// extractFloat extrae un float del texto entre prefix y suffix
func extractFloat(text, prefix, suffix string) float64 {
	start := strings.Index(text, prefix)
	if start == -1 {
		return 0
	}
	start += len(prefix)

	end := strings.Index(text[start:], suffix)
	if end == -1 {
		return 0
	}

	valStr := strings.TrimSpace(text[start : start+end])
	val, _ := strconv.ParseFloat(valStr, 64)
	return val
}

// extractString extrae un string del texto entre prefix y suffix
func extractString(text, prefix, suffix string) string {
	start := strings.Index(text, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)

	end := strings.Index(text[start:], suffix)
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(text[start : start+end])
}
