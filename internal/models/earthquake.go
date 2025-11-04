package models

import (
	"encoding/json"
	"time"
)

// Earthquake representa un sismo con toda su información
type Earthquake struct {
	ID           string    `json:"id"`
	Magnitude    float64   `json:"magnitude"`
	Location     string    `json:"location"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Depth        float64   `json:"depth"`                  // en kilómetros
	Time         time.Time `json:"-"`                      // Ocultamos el campo original
	Source       string    `json:"source"`                 // USGS, GEOFON, SGC
	Oceano       string    `json:"oceano,omitempty"`       // Pacifico, Caribe
	OceanoRegion string    `json:"oceanoRegion,omitempty"` // local, regional, lejano
	URL          string    `json:"url,omitempty"`
}

// MarshalJSON personaliza la serialización del Earthquake para formatear el tiempo
func (e Earthquake) MarshalJSON() ([]byte, error) {
	type Alias Earthquake
	return json.Marshal(&struct {
		Time string `json:"time"`
		*Alias
	}{
		Time:  e.Time.Format("2006-01-02 15:04:05"),
		Alias: (*Alias)(&e),
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
