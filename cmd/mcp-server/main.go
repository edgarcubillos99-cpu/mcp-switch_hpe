package main

import (
	"fmt"
	"log"
	"net/http"

	"noc-mcp/internal/auth"
	"noc-mcp/internal/config"
	"noc-mcp/internal/mcp"
)

func main() {
	// 1. Cargar configuración y comandos dinámicos
	// Nota: Si ejecutas con "go run cmd/mcp-server/main.go" desde la raíz, la ruta es "commands.yaml"
	cfg, err := config.LoadConfig("commands.yaml")
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// 2. Configurar el manejador MCP
	mcpHandler := &mcp.Handler{Config: cfg}

	// 3. Configurar el enrutador HTTP (Mux)
	secureMux := http.NewServeMux()

	// Ruta POST: Para ejecutar comandos (La que ya usaba n8n)
	secureMux.Handle("/mcp/v1/tools/execute", auth.APIKeyMiddleware(cfg.Server.APIKey, http.HandlerFunc(mcpHandler.ServeHTTP)))

	// NUEVA Ruta GET: Para descubrir herramientas (Auto-documentación para la IA)
	secureMux.Handle("/mcp/v1/tools", auth.APIKeyMiddleware(cfg.Server.APIKey, http.HandlerFunc(mcpHandler.ListTools)))

	// 4. Iniciar el servidor
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Iniciando NOC MCP Server en %s...", addr)
	log.Printf("Comandos registrados y autodescriptivos: %d", len(cfg.Commands))

	if err := http.ListenAndServe(addr, secureMux); err != nil {
		log.Fatalf("Error crítico en el servidor: %v", err)
	}
}
