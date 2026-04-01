package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"noc-mcp/internal/config"
	mcp_handler "noc-mcp/internal/mcp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 1. Cargar el catálogo de comandos (El Menú)
	cfg, err := config.LoadConfig("commands.yaml")
	if err != nil {
		log.Fatalf("Error cargando commands.yaml: %v", err)
	}

	// 2. Crear el Servidor MCP
	s := server.NewMCPServer(
		"noc-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Regex para encontrar variables dinámicas tipo {{variable}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	// 3. Registrar herramientas de forma dinámica y automática
	for name, cmdDef := range cfg.Commands {
		// Opciones base de la herramienta (Descripción y la IP obligatoria)
		opts := []mcp.ToolOption{
			mcp.WithDescription(cmdDef.Description),
			mcp.WithString("switch_ip", mcp.Required(), mcp.Description("La IP del switch a diagnosticar (ej. 10.254.254.57)")),
		}

		// Escanear el comando en busca de variables adicionales y añadirlas al esquema MCP
		matches := re.FindAllStringSubmatch(cmdDef.Command, -1)
		for _, match := range matches {
			if len(match) > 1 {
				argName := match[1]
				desc := fmt.Sprintf("Parámetro dinámico requerido: %s", argName)
				opts = append(opts, mcp.WithString(argName, mcp.Required(), mcp.Description(desc)))
			}
		}

		// Crear la herramienta con todas las opciones acumuladas
		tool := mcp.NewTool(name, opts...)

		// Añadir la herramienta al servidor vinculada a su manejador
		s.AddTool(tool, mcp_handler.CreateToolHandler(cmdDef))
	}

	// 4. Configurar Transporte SSE (Server-Sent Events) para la red
	// Si usas Ngrok/Tailscale, debes pasar la URL pública mediante esta variable de entorno
	baseURL := os.Getenv("MCP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	sse := server.NewSSEServer(s,
		server.WithBaseURL(baseURL),
		// Debe coincidir con la ruta montada abajo (el SDK usa "/message" por defecto)
		server.WithMessageEndpoint("/messages"),
	)

	// Handlers reales del SDK (HandleSSE/HandleMessage están sin implementar o ausentes en algunas versiones)
	http.Handle("/sse", sse.SSEHandler())
	http.Handle("/messages", sse.MessageHandler())

	port := ":8080"
	log.Printf("🚀 Servidor MCP nativo escuchando vía SSE en el puerto %s...", port)
	log.Printf("URL Base configurada: %s", baseURL)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error iniciando servidor HTTP/SSE: %v", err)
	}
}
