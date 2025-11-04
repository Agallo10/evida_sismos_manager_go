package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andresgallo/evida_backend_go/internal/manager"
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
