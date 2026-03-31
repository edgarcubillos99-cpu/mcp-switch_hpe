package ssh

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// ExecuteCommand conecta al switch HPE, ejecuta el comando y devuelve la salida
func ExecuteCommand(host, user, password, command string) (string, error) {
	// Configuración de SSH con Timeout de 10 segundos
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password), // En producción, preferir llaves SSH (PublicKeys)
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Nota: En prod, validar el HostKey
		Timeout:         10 * time.Second,
	}

	// Conectar al switch
	address := fmt.Sprintf("%s:22", host)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return "", fmt.Errorf("error conectando al switch %s: %w", host, err)
	}
	defer client.Close()

	// Iniciar sesión
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("error creando sesión SSH: %w", err)
	}
	defer session.Close()

	// Capturar salida estándar y errores
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	// Ejecutar el comando
	err = session.Run(command)
	if err != nil {
		return "", fmt.Errorf("error ejecutando comando: %w", err)
	}

	return stdoutBuf.String(), nil
}
