package telnet

import (
	"fmt"
	"strings"
	"time"

	"github.com/ziutek/telnet"
)

// ExecuteCommand conecta por Telnet, captura el prompt real y ejecuta el comando
func ExecuteCommand(host, user, password, command string) (string, error) {
	address := fmt.Sprintf("%s:23", host)

	// 1. Conectar
	t, err := telnet.Dial("tcp", address)
	if err != nil {
		return "", fmt.Errorf("error conectando por telnet a %s: %w", host, err)
	}
	defer t.Close()

	// Timeout inicial para la autenticación
	t.SetReadDeadline(time.Now().Add(10 * time.Second))
	t.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Enviar un salto de línea por si el switch necesita "despertar" el prompt
	t.Write([]byte("\r\n"))

	// 2. Autenticación (Usuario)
	if err := expect(t, "sername:", "ogin:", "sername", "ogin"); err != nil {
		return "", fmt.Errorf("timeout esperando prompt de usuario: %w", err)
	}
	if err := send(t, user); err != nil {
		return "", err
	}

	// 3. Autenticación (Password)
	if err := expect(t, "assword:", "assword"); err != nil {
		return "", fmt.Errorf("timeout esperando prompt de password: %w", err)
	}
	if err := send(t, password); err != nil {
		return "", err
	}

	// 4. CAPTURA DINÁMICA DEL PROMPT REAL
	// Leemos hasta el primer símbolo genérico de prompt
	data, err := t.ReadUntil(">", "#", "]")
	if err != nil {
		return "", fmt.Errorf("timeout esperando prompt inicial: %w", err)
	}

	// ReadUntil devuelve los datos incluyendo el delimitador al final.
	// Extraemos la última línea no vacía como prompt exacto.
	// Ejemplo: data="\r\nWelcome!\r\n<Switch>" → actualPrompt="<Switch>"
	fullOutput := string(data)
	lines := strings.Split(strings.ReplaceAll(fullOutput, "\r\n", "\n"), "\n")
	actualPrompt := strings.TrimSpace(lines[len(lines)-1])

	// Fallback por seguridad si la limpieza dejó el string vacío
	if actualPrompt == "" && len(data) > 0 {
		actualPrompt = string(data[len(data)-1:])
	}

	// 5. Desactivar paginación (--More--)
	if err := send(t, "screen-length disable"); err != nil {
		return "", err
	}
	// Leemos la salida hasta que vuelva a aparecer nuestro prompt exacto
	_, err = t.ReadUntil(actualPrompt)
	if err != nil {
		// Fallback: Si no acepta screen-length, intentamos con el comando de Allied Telesis
		send(t, "terminal length 0")
		t.ReadUntil(actualPrompt)
	}

	// 6. Ejecutar el comando del NOC
	// Aumentamos el timeout a 30 segundos porque "display current-configuration" es pesado
	t.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err := send(t, command); err != nil {
		return "", err
	}

	// 7. Leer TODA la salida hasta encontrar de nuevo el prompt exacto
	outData, err := t.ReadUntil(actualPrompt)
	if err != nil {
		return "", fmt.Errorf("error leyendo salida del comando: %w", err)
	}

	return cleanOutput(string(outData), command), nil
}

// Helpers
func expect(t *telnet.Conn, prompts ...string) error {
	return t.SkipUntil(prompts...)
}

func send(t *telnet.Conn, data string) error {
	buf := make([]byte, len(data)+2)
	copy(buf, data)
	copy(buf[len(data):], "\r\n")
	_, err := t.Write(buf)
	return err
}

func cleanOutput(raw string, command string) string {
	lines := strings.Split(raw, "\n")
	var cleaned []string
	for _, line := range lines {
		// Limpiamos retornos de carro
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))

		// Ignoramos la línea si es el eco del comando
		if line != "" && !strings.Contains(line, command) && !strings.Contains(line, "screen-length disable") {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}
