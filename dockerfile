# ==========================================
# Etapa 1: Compilación (Builder)
# ==========================================
FROM golang:1.25-alpine AS builder

# Instalar git por si hay dependencias privadas
RUN apk update && apk add --no-cache git

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos de dependencias primero (aprovecha la caché de capas de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY . .

# Compilar el binario optimizado estáticamente
# CGO_ENABLED=0 asegura que no haya dependencias dinámicas de C, ideal para Alpine
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-server ./cmd/mcp-server/

# ==========================================
# Etapa 2: Producción (Imagen Ligera)
# ==========================================
FROM alpine:latest

# Añadir certificados raíz (por si el MCP se comunica por HTTPS) y datos de zona horaria
RUN apk --no-cache add ca-certificates tzdata

# Crear un usuario no-root para mayor seguridad en el despliegue
RUN adduser -D -g '' nocuser

WORKDIR /app

# Copiar solo el binario compilado desde la etapa anterior
COPY --from=builder /app/mcp-server .

# Otorgar permisos de ejecución y cambiar al usuario no-root
RUN chown nocuser:nocuser /app/mcp-server
USER nocuser

# Exponer el puerto
EXPOSE 8080

# Ejecutar el servidor
CMD ["./mcp-server"]