# NOC MCP Server - AI Network Agent

Este proyecto es un servidor **MCP (Model Context Protocol)** desarrollado en **Golang**. Actúa como un puente seguro entre un Agente de Inteligencia Artificial (IA) y los equipos de red del NOC (Network Operations Center), específicamente switches HPE Comware y Allied Telesis.

El sistema permite a la IA ejecutar comandos de diagnóstico (`display`) en los switches, capturando la salida y devolviéndola en formato JSON para su análisis, todo ello manejando automáticamente las complejidades de la interacción por consola.

---

## 🚀 Características Principales

* **Integración Telnet Inteligente:** Diseñado para equipos de red legacy o entornos donde SSH no está habilitado.
* **Captura Dinámica de Prompt:** El servidor detecta automáticamente el nombre real del switch (ej. `<os.aghp.cid.cidra>`) al iniciar sesión. Esto evita cortes prematuros en la lectura al procesar configuraciones que contienen el carácter `#` (usado frecuentemente en HPE Comware como separador).
* **Bypass Automático de Paginación:** Envía automáticamente el comando `screen-length disable` (o `terminal length 0` como fallback) para evitar interrupciones por `--More--` en salidas largas.
* **Sistema de Comandos Dinámicos y Parametrizados:** Los comandos del NOC se definen en un archivo `commands.yaml` externo. Permite inyectar variables en tiempo real (ej. `display interface {{interface}}`) sin tener que recompilar el código Go.
* **Seguridad:** Protegido mediante un middleware de validación por **API Key**.
* **Contenedorización Optimizada:** Utiliza Docker Multi-stage build (Go + Alpine) resultando en una imagen ultraligera y segura, gestionada mediante `docker-compose`.

---

## 📂 Estructura del Proyecto

```text
/noc-mcp
├── cmd/mcp-server/
│   └── main.go              # Punto de entrada de la aplicación
├── internal/
│   ├── auth/middleware.go   # Middleware de seguridad (API Key)
│   ├── config/loader.go     # Carga de comandos dinámicos desde YAML
│   ├── mcp/handler.go       # Procesamiento de requests del Agente IA y reemplazo de variables
│   └── telnet/client.go     # Módulo Telnet interactivo (autenticación, prompts, paginación)
├── commands.yaml            # Listado de comandos autorizados y plantillas
├── docker-compose.yml       # Orquestación de contenedores
├── Dockerfile               # Configuración de imagen Multi-stage
├── .env.template            # Plantilla de variables de entorno
├── go.mod
└── go.sum
```

---

# ⚙️ Configuración y Despliegue

### 1. Variables de Entorno

Copia el archivo de plantilla para crear tu configuración local

Edita el archivo .env para definir el puerto y tu API Key segura.

### 2. Comandos Dinámicos (commands.yaml)

El archivo commands.yaml contiene los comandos permitidos. Los comandos que requieran argumentos específicos de la IA deben usar la sintaxis {{variable}}:

```plaintext
commands:
  display_current_config: "display current-configuration"
  display_interface: "display interface {{interface}}"
  display_mac_vsi_vlan: "display mac-address vsi {{vlan}}"
```

### 3. Despliegue con Docker (Recomendado)

El proyecto incluye un docker-compose.yml que monta el archivo commands.yaml como un volumen de solo lectura. Esto permite actualizar los comandos del NOC sin reiniciar el contenedor.

```bash
docker compose up -d --build
```

Para ver los logs en tiempo real:

```bash
docker compose logs -f
```

---

# 🔌 Uso e Integración (API Reference)

El servidor expone un endpoint HTTP POST configurado para recibir peticiones JSON estructuradas.

Endpoint: POST /mcp/v1/tools/execute

Headers Requeridos:

    Authorization: Bearer <TU_API_KEY>

    Content-Type: application/json

Ejemplo de Petición con Parámetros (IA -> MCP)

Cuando el Agente de IA necesita consultar una interfaz específica, envía un payload con el objeto arguments:

```bash
curl -X POST http://localhost:8080/mcp/v1/tools/execute \
-H "Authorization: Bearer super-secret-noc-key" \
-H "Content-Type: application/json" \
-d '{
  "tool_name": "display_interface",
  "switch_ip": "10.254.254.57",
  "switch_user": "allied",
  "switch_pass": "4ll13d",
  "arguments": {
    "interface": "Ten-GigabitEthernet1/0/25"
  }
}'
```

Respuesta Exitosa

El MCP devuelve un JSON estructurado con la salida pura del equipo de red, lista para que el LLM la interprete:

```bash
{
  "status": "success",
  "output": "Ten-GigabitEthernet1/0/25 current state: UP\nLine protocol current state: UP\n..."
}
```
