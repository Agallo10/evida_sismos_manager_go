package geometry

import (
	"encoding/json"
	"log"
	"os"

	"github.com/andresgallo/evida_backend_go/internal/models"
)

// RegionData contiene todos los polígonos de las regiones
type RegionData struct {
	LatlonCPWorld            models.Polygon   `json:"latlonCPWorld"`
	LatlonCPWorldEste        models.Polygon   `json:"latlonCPWorldEste"`
	LatlonPacificoLocal      models.Polygon   `json:"latlonPacificoLocal"`
	LatlonPacificoRegional   models.Polygon   `json:"latlonPacificoRegional"`
	LatlonPacificoLocal20Km  models.Polygon   `json:"latlonPacificoLocal20Km"`
	LatlonCCWorld            []models.Polygon `json:"latlonCCWorld"`
	LatlonCaribeRegional     []models.Polygon `json:"latlonCaribeRegional"`
	LatlonCaribeLocal        models.Polygon   `json:"latlonCaribeLocal"`
	LatlonCaribeLocalInsular models.Polygon   `json:"latlonCaribeLocalInsular"`
}

var regionData *RegionData

// LoadRegionData carga los datos de regiones desde el archivo JSON
func LoadRegionData(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	regionData = &RegionData{}
	if err := json.Unmarshal(data, regionData); err != nil {
		return err
	}

	// Cargar latlonCPWorldEste desde su archivo separado
	cpWorldEstePath := "internal/geometry/latlonCPWorldEste.json"
	cpWorldEsteData, err := os.ReadFile(cpWorldEstePath)
	if err != nil {
		log.Printf("⚠️  No se pudo cargar %s: %v", cpWorldEstePath, err)
	} else {
		var cpWorldEsteFile struct {
			LatlonCPWorldEste models.Polygon `json:"latlonCPWorldEste"`
		}
		if err := json.Unmarshal(cpWorldEsteData, &cpWorldEsteFile); err != nil {
			log.Printf("⚠️  Error parseando %s: %v", cpWorldEstePath, err)
		} else {
			regionData.LatlonCPWorldEste = cpWorldEsteFile.LatlonCPWorldEste
		}
	}

	log.Printf("✅ Datos de regiones cargados correctamente")
	log.Printf("   - Pacífico CP: %d puntos", len(regionData.LatlonCPWorld))
	log.Printf("   - Pacífico CP Este: %d puntos", len(regionData.LatlonCPWorldEste))
	log.Printf("   - Pacífico Local: %d puntos", len(regionData.LatlonPacificoLocal))
	log.Printf("   - Caribe CC: %d polígonos", len(regionData.LatlonCCWorld))
	log.Printf("   - Caribe Regional: %d polígonos", len(regionData.LatlonCaribeRegional))

	return nil
}

// GetRegionData retorna los datos de regiones cargados
func GetRegionData() *RegionData {
	return regionData
}
