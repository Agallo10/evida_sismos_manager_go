# Variables
BINARY_NAME=evida-server
GO=go
GOFLAGS=-v

# Colores para output
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

.PHONY: all build run test clean deps help

all: deps test build

## help: Muestra esta ayuda
help:
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## deps: Descarga las dependencias
deps:
	@echo "${GREEN}📦 Descargando dependencias...${NC}"
	$(GO) mod download
	$(GO) mod tidy

## build: Compila la aplicación
build:
	@echo "${GREEN}🔨 Compilando...${NC}"
	$(GO) build $(GOFLAGS) -o bin/$(BINARY_NAME) cmd/server/main.go
	@echo "${GREEN}✅ Compilación exitosa: bin/$(BINARY_NAME)${NC}"

## run: Ejecuta la aplicación
run:
	@echo "${GREEN}🚀 Ejecutando servidor...${NC}"
	$(GO) run cmd/server/main.go

## test: Ejecuta los tests
test:
	@echo "${GREEN}🧪 Ejecutando tests...${NC}"
	$(GO) test -v ./...

## test-cover: Ejecuta tests con cobertura
test-cover:
	@echo "${GREEN}🧪 Ejecutando tests con cobertura...${NC}"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}✅ Reporte de cobertura: coverage.html${NC}"

## clean: Limpia archivos generados
clean:
	@echo "${YELLOW}🧹 Limpiando...${NC}"
	$(GO) clean
	rm -rf bin/
	rm -f coverage.out coverage.html

## fmt: Formatea el código
fmt:
	@echo "${GREEN}✨ Formateando código...${NC}"
	$(GO) fmt ./...

## lint: Ejecuta linter
lint:
	@echo "${GREEN}🔍 Ejecutando linter...${NC}"
	golangci-lint run

## docker-build: Construye imagen Docker
docker-build:
	@echo "${GREEN}🐳 Construyendo imagen Docker...${NC}"
	docker build -t evida-backend:latest .

## docker-run: Ejecuta contenedor Docker
docker-run:
	@echo "${GREEN}🐳 Ejecutando contenedor Docker...${NC}"
	docker run -p 8080:8080 evida-backend:latest

## dev: Ejecuta en modo desarrollo con recarga automática (requiere air)
dev:
	@echo "${GREEN}🔥 Modo desarrollo con recarga automática...${NC}"
	air
