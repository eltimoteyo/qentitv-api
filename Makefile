.PHONY: run build test clean migrate deps

# Variables
BINARY_NAME=qenti
MAIN_PATH=cmd/server/main.go

# Ejecutar la aplicación
run:
	go run $(MAIN_PATH)

# Compilar la aplicación
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Ejecutar tests
test:
	go test -v ./...

# Limpiar binarios
clean:
	go clean
	rm -f $(BINARY_NAME)

# Instalar dependencias
deps:
	go mod download
	go mod tidy

# Ejecutar migraciones
migrate:
	go run cmd/server/main.go

# Formatear código
fmt:
	go fmt ./...

# Linter (requiere golangci-lint)
lint:
	golangci-lint run

# Ejecutar en modo desarrollo con hot reload (requiere air)
dev:
	air

