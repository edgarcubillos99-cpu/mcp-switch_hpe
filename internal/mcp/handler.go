package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"noc-mcp/internal/config"
	"noc-mcp/internal/ssh"
)

// ToolRequest representa la petición que enviaría el Agente IA
type ToolRequest struct {
	ToolName   string `json:"tool_name"`
	SwitchIP   string `json:"switch_ip"`
	SwitchUser string `json:"switch_user"`
	SwitchPass string `json:"switch_pass"` // Idealmente, esto vendría de un Vault, no del agente
}

type Handler struct {
	Config *config.Config
}

// ServeHTTP maneja la ejecución de las herramientas (Tools) del MCP
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Validar que el comando (Tool) exista en la configuración dinámica
	realCommand, exists := h.Config.Commands[req.ToolName]
	if !exists {
		http.Error(w, fmt.Sprintf("Tool %s no está registrada", req.ToolName), http.StatusBadRequest)
		return
	}

	// 2. Ejecutar el comando vía SSH en el switch
	output, err := ssh.ExecuteCommand(req.SwitchIP, req.SwitchUser, req.SwitchPass, realCommand)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Devolver el resultado al Agente IA
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"output": output,
	})
}
