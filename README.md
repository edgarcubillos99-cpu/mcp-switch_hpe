# 🤖 NOC OmniDiagnostic Service – Switch HPE + MCP + Golang

Este proyecto es un servidor **MCP (Model Context Protocol)** desarrollado en **Golang**. Actúa como un puente seguro entre un Agente de Inteligencia Artificial (IA) y los equipos de red del NOC (Network Operations Center), específicamente switches HPE Comware y Allied Telesis.

✅ Implementación nativa del Model Context Protocol (MCP) vía HTTP REST
✅ Endpoint de "Descubrimiento" para auto-documentación de herramientas en el LLM
✅ Conexión Telnet resiliente con captura dinámica de prompts de consola
✅ Evasión automatizada de paginación (--More--) en equipos HPE/Comware
✅ Inyección dinámica de variables (interfaces, VLANs) mediante plantillas YAML
✅ Seguridad Zero-Knowledge: La IA nunca conoce las contraseñas de los switches
✅ Validación estricta mediante API Keys
✅ Despliegue ultraligero con Docker Multi-stage (Go + Alpine)
✅ Integración probada con n8n (Advanced AI Agents)

El sistema permite a la IA ejecutar comandos de diagnóstico (`display`) en los switches, capturando la salida y devolviéndola en formato JSON para su análisis, todo ello manejando automáticamente las complejidades de la interacción por consola.

---

# 📑 Índice

    🔎 Descripción General
    
    📁 Estructura del Proyecto
    
    🏗 Arquitectura del Sistema
    
    ⚙️ Configuración del Entorno
    
    🧠 Integración con Agentes IA (n8n / MCP)
    
    🐳 Ejecución con Docker
    
    🧩 Diseño del Sistema
    
    💾 Modelo de Datos (Payloads)
    
    🧪 Pruebas y Verificación
    
    🚨 Troubleshooting

---

# 🔎 Descripción General

Este microservicio actúa como un "brazo robótico" y orquestador entre los Agentes de Inteligencia Artificial (LLMs) y la infraestructura física de red (Switches HPE Comware / Allied Telesis). Su objetivo es traducir las intenciones cognitivas de la IA en comandos de red precisos, ejecutarlos de forma segura y devolver los resultados puros.

El sistema aísla por completo a la IA de la complejidad de la consola de red: maneja automáticamente los tiempos de espera de Telnet, desactiva la paginación interactiva, sortea los banners legales de inicio de sesión y detecta dinámicamente el nombre del switch para saber exactamente cuándo termina la salida de un comando. A través de su arquitectura MCP, expone un catálogo vivo de herramientas (Tools) que la IA puede leer, comprender y ejecutar bajo demanda.

---

## 📂 Estructura del Proyecto

```text
noc-mcp/
├── cmd/
│   └── mcp-server/
│       └── main.go              → Bootstrap, enrutamiento HTTP y middlewares
├── internal/
│   ├── auth/                    → Lógica de validación de API Key
│   ├── config/                  → Parseo de commands.yaml (CommandDef)
│   ├── mcp/                     → Endpoints de Ejecución (POST) y Descubrimiento (GET)
│   └── telnet/                  → Cliente Telnet, captura de prompts y manejo de buffers
├── commands.yaml                → Catálogo dinámico de comandos, descripciones y reglas
├── .env.template                → Plantilla de variables de entorno
├── docker-compose.yml           → Orquestación de infraestructura local
├── Dockerfile                   → Construcción Multi-stage (Golang builder -> Alpine)
├── go.mod
└── go.sum
```

---

# 🏗 Arquitectura del Sistema

```plaintext
A[Usuario en Teams/WhatsApp] <--> B[Agente IA en n8n]
B -->|1. GET /mcp/v1/tools| C[MCP Server: Endpoint de Descubrimiento]
C -->|Devuelve Catálogo YAML| B
B -->|2. POST /mcp/v1/tools/execute| D[MCP Server: Endpoint de Ejecución]
D -->|Valida API Key y YAML| E[Módulo Telnet]
E -->|Inyecta Credenciales del Entorno| F[Conexión TCP:23]
F -->|Autenticación + Captura Dinámica| G[Switch HPE / Comware]
G -->|Ejecuta Comando (ej. display version)| E
E -->|Limpia Output| D
D -->|Devuelve JSON de Éxito| B
B -->|Análisis Cognitivo| A[Diagnóstico Humano]
```

---
# ⚙️ Configuración y Despliegue

 💻 Aplicación y Puerto

PORT=8080
TZ=America/Bogota

--------------------------------------------------------
### 🟢 SEGURIDAD DE LA API (MCP)
--------------------------------------------------------

MCP_API_KEY=super-secret-noc-key-cambiar-en-produccion

--------------------------------------------------------
### 🔐 CREDENCIALES DE RED (ZERO-KNOWLEDGE)
--------------------------------------------------------
IMPORTANTE: Estas credenciales son inyectadas directamente en la sesión Telnet.
La IA jamás tiene acceso a estas variables, previniendo fugas de seguridad por "Prompt Injection".

NOC_SWITCH_USER=

NOC_SWITCH_PASSWORD=

---

# 🧠 Integración con Agentes IA (n8n / MCP)

Para el correcto funcionamiento del ecosistema, el Agente de IA debe configurarse con el siguiente System Message y en dos fases dentro de herramientas como n8n:

```plaintext
Eres un ingeniero experto del NOC de una empresa de telecomunicaciones. Tu trabajo es diagnosticar equipos de red.
TIENES a tu disposición una herramienta llamada 'ejecutar_comando_switch'.
Cuando el usuario te pida revisar algo en un equipo, NO inventes la respuesta. DEBES usar obligatoriamente tu herramienta para conectarte al equipo, obtener la salida real y luego explicarle al usuario de forma humana y resumida lo que encontraste.
```

### a) Descubrimiento (El Menú)

Configura un nodo HTTP Request que se ejecute antes del Agente.

    Método: GET

    URL: http://<APP_BASE_URL>:8080/mcp/v1/tools

    El resultado de este nodo debe inyectarse en el System Prompt de la IA para que aprenda qué herramientas existen y qué argumentos requieren.

### b) Ejecución (El Brazo)

Conecta un nodo de Herramienta (HTTP Request Tool) al Agente.

    Método: POST

    URL: http://<APP_BASE_URL>:8080/mcp/v1/tools/execute

    Body: La IA debe generar dinámicamente el tool_name, switch_ip y arguments basados en el catálogo descubierto.

    Description:

```plaintext
Ejecuta comandos de diagnóstico en switches HPE/Comware del NOC. 
PARÁMETROS OBLIGATORIOS a generar en tu respuesta JSON:

1. 'switch_ip': La dirección IP del equipo a diagnosticar (ej. 10.254.254.57).
2. 'tool_name': El comando exacto a ejecutar. DEBES elegir EXACTAMENTE uno de la siguiente lista. PROHIBIDO usar espacios, respeta los guiones bajos.
3. 'arguments': Un objeto JSON. Su contenido depende del 'tool_name' elegido.

--- COMANDOS ESTÁTICOS ---
(Para estos, envía SIEMPRE un objeto vacío en los argumentos: "arguments": {})
- 'display_version' : Ver versión, modelo, firmware y uptime.
- 'display_current_config' : Ver la configuración completa.
- 'display_device' : Ver estado del hardware.
- 'display_cpu' : Ver uso general de procesador.
- 'display_memory' : Ver uso de memoria RAM.
- 'display_clock' : Ver hora y fecha configurada.
- 'display_interfaces_brief' : Resumen rápido del estado de todos los puertos.
- 'display_ip_int_brief' : Resumen de interfaces con configuración IP.
- 'display_vlan_all' : Mostrar todas las VLANs configuradas.
- 'display_routing_table' : Ver la tabla de enrutamiento IP.
- 'display_arp' : Ver la tabla ARP (relación IP/MAC).
- 'display_lldp_neighbors' : Ver resumen de todos los equipos vecinos conectados.
- 'display_ospf_peer' : Ver resumen de vecinos OSPF.
- 'display_ospf_peer_verbose' : Ver detalles profundos de vecinos OSPF.
- 'display_vsi' : Listar todas las Virtual Switch Instances.

--- COMANDOS DINÁMICOS ---
(Para estos, "arguments" DEBE ser un objeto JSON con la clave indicada)

* Requieren la clave "interface" (Ejemplo: "arguments": { "interface": "Ten-GigabitEthernet1/0/25" }):
- 'display_interface' : Ver estado detallado, tráfico y errores de un puerto específico.
- 'display_lldp_interface' : Ver qué equipo está conectado a un puerto específico.
- 'display_transceiver' : Ver niveles ópticos (luz) del transceptor (SFP) de un puerto específico.
- 'display_mac_interface' : Ver qué direcciones MAC se aprenden en un puerto específico.

* Requieren otras claves:
- 'display_mac_vsi_vlan' : Ver tabla MAC de una VLAN específica. (Ejemplo: "arguments": { "vlan": "800" })
- 'display_vsi_verbose' : Ver detalles completos de una VSI. (Ejemplo: "arguments": { "vsi_name": "vlan2219" })

REGLA FINAL: Nunca inventes tool_names. Limítate a usar los de esta lista.
```

---

# 🐳 Ejecución con Docker

El despliegue está contenerizado mediante un Multi-stage build para garantizar que el entorno de producción no contenga el código fuente, sino únicamente un binario compilado estáticamente corriendo sobre un SO Alpine ultra-ligero.

```Bash
docker-compose up -d --build
```

El proceso:

    Inicia la etapa builder con Golang 1.22, descarga dependencias y compila.

    Inicia la etapa de producción transfiriendo solo el binario a una imagen Alpine.

    El archivo commands.yaml se monta como un volumen read-only (ro). Esto permite modificar o agregar comandos al vuelo sin tener que reconstruir la imagen Docker.

Para ver registros en tiempo real:

```Bash
docker compose logs -f
```

---

# 🧩 Diseño del Sistema

✔ Captura Dinámica del Prompt: Dado que la configuración de los switches Comware suele contener caracteres como #, el sistema no depende de delimitadores fijos. Al iniciar sesión, el sistema lee la última línea (ej. <os.aghp.cid.cidra>) y la almacena en memoria. El comando solo se da por finalizado cuando el switch vuelve a imprimir esa cadena exacta.

✔ Evasión de Paginación: El cliente Telnet envía automáticamente screen-length disable (HPE) o terminal length 0 (Allied Telesis) antes de cualquier diagnóstico, asegurando que comandos pesados como display current-configuration no se bloqueen por prompts de --More--.

✔ Plantillas YAML Agnósticas: La lógica de Go no conoce los comandos reales. Lee el archivo commands.yaml donde se definen reglas estrictas (si un comando requiere variables como {{interface}}). Esto permite que el equipo del NOC agregue nuevas capacidades sin necesidad de intervención de los desarrolladores.

---

# 💾 Modelo de Datos (Payloads)

### 1. Payload de Descubrimiento (GET /mcp/v1/tools)

Devuelve el catálogo de herramientas y sus reglas.

```Bash
{
  "tools_available": 21,
  "catalog": {
    "display_interface": {
      "description": "Muestra el estado detallado, tráfico y errores de un puerto específico.",
      "requires_arguments": true,
      "argument_hint": "Requiere la clave 'interface' en los argumentos. Ejemplo: {'interface': 'Ten-GigabitEthernet1/0/25'}"
    }
  }
}
```

### 2. Payload de Ejecución (POST /mcp/v1/tools/execute)

Enviado por la IA hacia el servidor MCP.

```Bash
{
  "tool_name": "display_interface",
  "switch_ip": "10.254.254.57",
  "arguments": {
    "interface": "Ten-GigabitEthernet1/0/25"
  }
}
```

--- 

# 🧪 Pruebas y Verificación

✅ Validar Catálogo MCP (Descubrimiento)
Verifica que el servidor esté leyendo correctamente el archivo YAML:

```Bash
curl -X GET http://localhost:8080/mcp/v1/tools \
-H "Authorization: Bearer super-secret-noc-key"
```

✅ Simular Ejecución de IA (Comando Básico)
Asegúrate de que la conexión Telnet funciona hacia tu equipo de pruebas:

```Bash
curl -X POST http://localhost:8080/mcp/v1/tools/execute \
-H "Authorization: Bearer super-secret-noc-key" \
-H "Content-Type: application/json" \
-d '{
  "tool_name": "display_version",
  "switch_ip": "10.254.254.57"
}'
```

---

# 🚨 Troubleshooting

❌ "Timeout esperando prompt de usuario" (HTTP 500)

    Causa: El equipo de red es inalcanzable, el puerto 23 está cerrado, o el banner de bienvenida es inusualmente largo.

    Solución: Valida el acceso ejecutando telnet <IP> manualmente desde el servidor donde corre el Docker. Revisa el log interno de Go que imprime el MOTD (Banner) exacto en caso de fallo.

❌ "Tool no está registrada" (HTTP 400)

    Causa: El Agente IA envió un espacio en lugar de un guion bajo (ej. display version en vez de display_version), o el JSON enviado por el LLM tiene un formato inválido.

    Solución: Revisa el System Prompt en n8n y asegúrate de usar JSON.stringify en el body de la petición HTTP para que los valores nulos no rompan la estructura JSON.

❌ La salida devuelve un # o texto incompleto

    Causa: El cliente Telnet falló al capturar dinámicamente el prompt y utilizó un delimitador genérico prematuro.

    Solución: Asegúrate de que las credenciales son correctas. Si la clave es errónea, el equipo devuelve un prompt secundario (Login incorrect) que engaña a la validación. Valida las variables .env.