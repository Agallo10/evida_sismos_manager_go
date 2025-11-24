package models

import (
	"encoding/json"
	"time"
)

// Earthquake representa un sismo con toda su información
type Earthquake struct {
	ID                string    `json:"id"`
	Magnitud          float64   `json:"magnitud"`
	Place             string    `json:"place"`
	CloserTowns       string    `json:"closerTowns,omitempty"`
	Latitud           float64   `json:"latitud"`
	Longitud          float64   `json:"longitud"`
	LongitudOperativa float64   `json:"longitudOperativa"`      // longitud - 360 para sismos en CPWorldEste, o igual a longitud para otros
	Profundidad       float64   `json:"profundidad"`            // en kilómetros
	Time              time.Time `json:"-"`                      // Campo interno (UTC)
	Fuente            string    `json:"fuente"`                 // USGS, GEOFON, SGC
	Oceano            string    `json:"oceano,omitempty"`       // Pacifico, Caribe
	OceanoRegion      string    `json:"oceanoRegion,omitempty"` // local, regional, lejano
	URL               string    `json:"url,omitempty"`
}

// MarshalJSON personaliza la serialización del Earthquake para formatear el tiempo
func (e Earthquake) MarshalJSON() ([]byte, error) {
	type Alias Earthquake
	return json.Marshal(&struct {
		LocalTime string `json:"localTime"`
		*Alias
	}{
		LocalTime: e.Time.Format("2006-01-02 15:04:05"),
		Alias:     (*Alias)(&e),
	})
}

// Point representa un punto geográfico
type Point struct {
	Lat float64
	Lon float64
}

// UnmarshalJSON personaliza la deserialización para aceptar arrays [lat, lon]
func (p *Point) UnmarshalJSON(data []byte) error {
	var arr []float64
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) >= 2 {
		p.Lat = arr[0]
		p.Lon = arr[1]
	}
	return nil
}

// MarshalJSON personaliza la serialización para retornar objetos {lat, lon}
func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}{
		Lat: p.Lat,
		Lon: p.Lon,
	})
}

// Polygon representa un polígono definido por una lista de puntos
type Polygon []Point

// Region representa una región geográfica con su nombre y polígono
type Region struct {
	Name    string  `json:"name"`
	Polygon Polygon `json:"polygon"`
}
