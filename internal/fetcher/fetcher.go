package fetcher

import "github.com/andresgallo/evida_backend_go/internal/models"

// Fetcher es la interfaz que deben implementar todos los fetchers
type Fetcher interface {
	Fetch() ([]models.Earthquake, error)
}
