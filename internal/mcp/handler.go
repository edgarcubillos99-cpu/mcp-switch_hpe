package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"noc-mcp/internal/config"
	"noc-mcp/internal/telnet"
)

// CreateToolHandler devuelve una función compatible con el SDK de mcp-go
func CreateToolHandler(cmdDef config.CommandDef) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// MCP entrega Arguments como any (JSON object); hay que asertar a mapa.
		argsRaw := request.Params.Arguments
		args, ok := argsRaw.(map[string]any)
		if !ok {
			if argsRaw == nil {
				args = map[string]any{}
			} else {
				return mcp.NewToolResultError("Los argumentos deben ser un objeto JSON (mapa de claves a valores)."), nil
			}
		}

		// 1. Extraer la IP del switch (que ahora será un argumento nativo)
		ipRaw, ok := args["switch_ip"].(string)
		if !ok || ipRaw == "" {
			return mcp.NewToolResultError("El parámetro 'switch_ip' es obligatorio."), nil
		}

		realCommand := cmdDef.Command

		// 2. Reemplazar dinámicamente cualquier otra variable (como {{interface}})
		for key, valRaw := range args {
			if key == "switch_ip" {
				continue
			}
			if val, ok := valRaw.(string); ok {
				placeholder := fmt.Sprintf("{{%s}}", key)
				realCommand = strings.ReplaceAll(realCommand, placeholder, val)
			}
		}

		// Validar que no haya quedado ninguna variable sin reemplazar
		if strings.Contains(realCommand, "{{") {
			return mcp.NewToolResultError(fmt.Sprintf("Faltan argumentos para completar el comando: %s", realCommand)), nil
		}

		// 3. Obtener credenciales seguras del entorno (Zero-Knowledge)
		user := os.Getenv("NOC_SWITCH_USER")
		pass := os.Getenv("NOC_SWITCH_PASSWORD")

		if user == "" || pass == "" {
			return mcp.NewToolResultError("Error crítico: Las credenciales NOC_SWITCH_USER o NOC_SWITCH_PASSWORD no están en el entorno."), nil
		}

		// 4. Ejecutar el comando en el switch
		output, err := telnet.ExecuteCommand(ipRaw, user, pass, realCommand)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error de red conectando al switch: %v", err)), nil
		}

		// 5. Devolver el texto limpio al LLM
		return mcp.NewToolResultText(output), nil
	}
}