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
	cfg, err := config.LoadConfig("../../commands.yaml")
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// 2. Configurar el manejador MCP
	mcpHandler := &mcp.Handler{Config: cfg}

	// 3. Envolver el manejador con el middleware de seguridad
	secureMux := http.NewServeMux()
	secureMux.Handle("/mcp/v1/tools/execute", auth.APIKeyMiddleware(cfg.Server.APIKey, mcpHandler))

	// 4. Iniciar el servidor
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Iniciando NOC MCP Server en %s...", addr)
	log.Printf("Comandos registrados: %d", len(cfg.Commands))

	if err := http.ListenAndServe(addr, secureMux); err != nil {
		log.Fatalf("Error en el servidor: %v", err)
	}
}
