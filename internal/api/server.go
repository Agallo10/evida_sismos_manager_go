package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/andresgallo/evida_backend_go/internal/manager"
	"github.com/andresgallo/evida_backend_go/internal/models"
	"github.com/andresgallo/evida_backend_go/internal/websocket"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Permitir todas las conexiones (cambiar en producción)
		return true
	},
}

// Server representa el servidor HTTP/WebSocket
type Server struct {
	manager *manager.EarthquakeManager
	hub     *websocket.Hub
}

// NewServer crea un nuevo servidor
func NewServer(manager *manager.EarthquakeManager, hub *websocket.Hub) *Server {
	return &Server{
		manager: manager,
		hub:     hub,
	}
}

// SetupRoutes configura las rutas del servidor
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// API REST endpoints
	mux.HandleFunc("/api/earthquakes", s.handleGetEarthquakes)
	mux.HandleFunc("/api/stats", s.handleGetStats)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/test/earthquakes", s.handleTestEarthquakes)

	return mux
}

// handleWebSocket maneja las conexiones WebSocket
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	websocket.ServeWs(s.hub, conn)
}

// handleGetEarthquakes retorna la lista de sismos
func (s *Server) handleGetEarthquakes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Obtener parámetros de consulta
	oceano := r.URL.Query().Get("oceano")
	region := r.URL.Query().Get("region")

	var earthquakes interface{}
	if oceano != "" {
		earthquakes = s.manager.GetByOceano(oceano)
	} else if region != "" {
		earthquakes = s.manager.GetByRegion(region)
	} else {
		earthquakes = s.manager.GetAll()
	}

	// Enviar respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(earthquakes); err != nil {
		log.Printf("Error encoding earthquakes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleGetStats retorna estadísticas de los sismos
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.manager.GetStats()
	stats["websocket_clients"] = s.hub.GetClientCount()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleHealth retorna el estado del servidor
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":            "ok",
		"earthquake_count":  s.manager.GetCount(),
		"websocket_clients": s.hub.GetClientCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(response)
}

// handleTestEarthquakes retorna sismos de prueba
func (s *Server) handleTestEarthquakes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Crear fecha/hora actual
	dateTimeZone := time.Now()

	// Crear los sismos de prueba
	testEarthquakes := []models.Earthquake{
		// {
		// 	ID:                "manualNew4",
		// 	Latitud:           2,
		// 	Longitud:          -80,
		// 	LongitudOperativa: -80,
		// 	Magnitud:          6,
		// 	Profundidad:       10,
		// 	Place:             "117 km S of La Libertad, El Salvador place",
		// 	CloserTowns:       "117 km S of La Libertad, El Salvador closerTowns",
		// 	Fuente:            ",us,",
		// 	Time:              dateTimeZone,
		// 	Oceano:            "Pacifico",
		// 	OceanoRegion:      "local",
		// },
		// {
		// 	ID:                "manual124",
		// 	Latitud:           -7,
		// 	Longitud:          -80,
		// 	LongitudOperativa: -80,
		// 	Magnitud:          7,
		// 	Profundidad:       10,
		// 	Place:             "117 km S of La Libertad, El Salvador",
		// 	CloserTowns:       "117 km S of La Libertad, El Salvador",
		// 	Fuente:            ",us,",
		// 	Time:              dateTimeZone,
		// 	Oceano:            "Pacifico",
		// 	OceanoRegion:      "regional",
		// },
		// {
		// 	ID:                "manualNew5",
		// 	Latitud:           2,
		// 	Longitud:          -80,
		// 	LongitudOperativa: -80,
		// 	Magnitud:          8,
		// 	Profundidad:       10,
		// 	Place:             "117 km S of La Libertad, El Salvador place",
		// 	CloserTowns:       "117 km S of La Libertad, El Salvador closerTowns",
		// 	Fuente:            ",us,",
		// 	Time:              dateTimeZone,
		// 	Oceano:            "Pacifico",
		// 	OceanoRegion:      "local",
		// },
		{
			ID:                "manualKamchaka",
			Latitud:           52.53,
			Longitud:          161.8352,
			LongitudOperativa: -199.8352,
			Magnitud:          8.8,
			Profundidad:       19.2,
			Place:             "Kamchatka Peninsula, Russia",
			CloserTowns:       "1Kamchatka Peninsula, Russia",
			Fuente:            "USGS",
			Time:              dateTimeZone,
			Oceano:            "Pacifico",
			OceanoRegion:      "lejano",
		},
		{
			ID:                "manualPiscoPeru",
			Latitud:           -13.39,
			Longitud:          -76.60,
			LongitudOperativa: -76.60,
			Magnitud:          8,
			Profundidad:       39,
			Place:             "Pisco, Peru",
			CloserTowns:       "Pisco, Peru",
			Fuente:            "USGS",
			Time:              dateTimeZone,
			Oceano:            "Pacifico",
			OceanoRegion:      "regional",
		},
		{
			ID:                "manualCaribeOriente",
			Latitud:           18.9115,
			Longitud:          -81.258,
			LongitudOperativa: -81.258,
			Magnitud:          8,
			Profundidad:       10,
			Place:             "Oriente - Caribbean",
			CloserTowns:       "Oriente - Caribbean",
			Fuente:            "USGS",
			Time:              dateTimeZone,
			Oceano:            "Caribe",
			OceanoRegion:      "regional",
		},
		{
			ID:                "manualAtlantico",
			Latitud:           24.0,
			Longitud:          -46.0,
			LongitudOperativa: -46.0,
			Magnitud:          8,
			Profundidad:       10,
			Place:             "Central_MARidge - Atlantico",
			CloserTowns:       "Central_MARidge - Atlantico",
			Fuente:            "USGS",
			Time:              dateTimeZone,
			Oceano:            "Atlantico",
			OceanoRegion:      "lejano",
		},
		{
			ID:                "manualPacificoLocal",
			Latitud:           2,
			Longitud:          -80,
			LongitudOperativa: -80,
			Magnitud:          8,
			Profundidad:       10,
			Place:             "Pacifico Local - Pacific",
			CloserTowns:       "Pacifico Local - Pacific",
			Fuente:            "SGC",
			Time:              dateTimeZone,
			Oceano:            "Pacifico",
			OceanoRegion:      "local",
		},
	}

	// Enviar respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(testEarthquakes); err != nil {
		log.Printf("Error encoding test earthquakes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
