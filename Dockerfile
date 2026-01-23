# Build stage
FROM golang:1.21-alpine AS builder

# Instalar dependencias del sistema
RUN apk add --no-cache git

# Establecer directorio de trabajo
WORKDIR /app

# Copiar go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o qenti-api cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar el binario desde el builder
COPY --from=builder /app/qenti-api .

# Exponer puerto
EXPOSE 8080

# Variables de entorno por defecto
ENV APP_ENV=production
ENV PORT=8080

# Comando para ejecutar la aplicación
CMD ["./qenti-api"]

