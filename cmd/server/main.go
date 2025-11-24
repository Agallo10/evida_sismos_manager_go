package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andresgallo/evida_backend_go/internal/api"
	"github.com/andresgallo/evida_backend_go/internal/fetcher"
	"github.com/andresgallo/evida_backend_go/internal/geometry"
	"github.com/andresgallo/evida_backend_go/internal/manager"
	"github.com/andresgallo/evida_backend_go/internal/websocket"
)

const (
	// Intervalo de actualización de datos (cada 2 minutos)
	fetchInterval = 2 * time.Minute

	// Tiempo máximo para mantener sismos en memoria (7 días)
	maxEarthquakeAge = 7 * 24 * time.Hour

	// Intervalo de limpieza de sismos antiguos (cada hora)
	cleanupInterval = 1 * time.Hour

	// Puerto del servidor
	serverPort = ":8080"
)

func main() {
	log.Println("🌍 Iniciando EVIDA Backend - Sistema de Monitoreo de Sismos")

	// Cargar datos de regiones desde archivo JSON
	regionDataPath := "internal/geometry/datosLC.json"
	if err := geometry.LoadRegionData(regionDataPath); err != nil {
		log.Fatalf("❌ Error cargando datos de regiones: %v", err)
	}

	// Crear gestor de sismos
	earthquakeManager := manager.NewEarthquakeManager(maxEarthquakeAge)
	log.Println("✅ Gestor de sismos inicializado")

	// Iniciar limpieza automática de sismos antiguos
	earthquakeManager.StartCleanup(cleanupInterval)
	log.Println("✅ Limpieza automática configurada")

	// Crear hub de WebSocket
	hub := websocket.NewHub()
	go hub.Run()
	log.Println("✅ Hub WebSocket iniciado")

	// Crear fetchers
	usgsFetcher := fetcher.NewUSGSFetcher()
	geofonFetcher := fetcher.NewGEOFONFetcher()
	sgcFetcher := fetcher.NewSGCFetcher()

	fetchers := []fetcher.Fetcher{
		usgsFetcher,
		geofonFetcher,
		sgcFetcher,
	}
	log.Printf("✅ Configurados %d fetchers de datos", len(fetchers))

	// Iniciar recolección de datos
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startDataCollection(ctx, fetchers, earthquakeManager, hub)
	log.Println("✅ Recolección de datos iniciada")

	// Iniciar notificaciones de WebSocket
	go startWebSocketNotifications(earthquakeManager, hub)
	log.Println("✅ Sistema de notificaciones iniciado")

	// Configurar servidor HTTP
	server := api.NewServer(earthquakeManager, hub)
	mux := server.SetupRoutes()

	httpServer := &http.Server{
		Addr:         serverPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Printf("🚀 Servidor HTTP escuchando en %s", serverPort)
		log.Println("   - WebSocket: ws://localhost:8080/ws")
		log.Println("   - API: http://localhost:8080/api/earthquakes")
		log.Println("   - Stats: http://localhost:8080/api/stats")
		log.Println("   - Health: http://localhost:8080/api/health")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Apagando servidor...")

	// Apagar servidor gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error apagando servidor: %v", err)
	}

	log.Println("✅ Servidor apagado correctamente")
}

// startDataCollection inicia la recolección periódica de datos de sismos
func startDataCollection(ctx context.Context, fetchers []fetcher.Fetcher, manager *manager.EarthquakeManager, hub *websocket.Hub) {
	// Ejecutar inmediatamente al inicio
	fetchAllData(fetchers, manager)

	// Luego ejecutar periódicamente
	ticker := time.NewTicker(fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Deteniendo recolección de datos")
			return
		case <-ticker.C:
			fetchAllData(fetchers, manager)
		}
	}
}

// fetchAllData obtiene datos de todos los fetchers
func fetchAllData(fetchers []fetcher.Fetcher, manager *manager.EarthquakeManager) {
	log.Println("🔄 Obteniendo datos de sismos...")

	totalNew := 0
	for i, f := range fetchers {
		earthquakes, err := f.Fetch()
		if err != nil {
			log.Printf("⚠️  Error fetching from source %d: %v", i+1, err)
			continue
		}

		newOnes := manager.AddEarthquakes(earthquakes)
		totalNew += len(newOnes)

		if len(newOnes) > 0 {
			log.Printf("   ➕ Fuente %d: %d nuevos sismos de %d totales", i+1, len(newOnes), len(earthquakes))
		}
	}

	if totalNew > 0 {
		log.Printf("✅ Total: %d nuevos sismos agregados", totalNew)
	} else {
		log.Println("   ℹ️  No hay sismos nuevos")
	}

	log.Printf("   📊 Total en memoria: %d sismos", manager.GetCount())
}

// startWebSocketNotifications escucha nuevos sismos y los envía por WebSocket
func startWebSocketNotifications(manager *manager.EarthquakeManager, hub *websocket.Hub) {
	earthquakeChan := manager.GetNewEarthquakeChannel()

	for event := range earthquakeChan {
		if event.IsUpdate {
			log.Printf("🔄 Sismo actualizado: M%.1f - %s [%s %s]",
				event.Earthquake.Magnitud, event.Earthquake.Place, event.Earthquake.Oceano, event.Earthquake.OceanoRegion)
		} else {
			log.Printf(" Nuevo sismo detectado: M%.1f - %s [%s %s]",
				event.Earthquake.Magnitud, event.Earthquake.Place, event.Earthquake.Oceano, event.Earthquake.OceanoRegion)
		}
		hub.BroadcastEarthquake(event.Earthquake, event.IsUpdate)
	}
}
