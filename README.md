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

MCP_BASE_URL=URL_IMPORTANTE_PARA_CONECTARCE

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

# 🧠 Integración con Agentes IA (n8n / Cline)

Gracias a la implementación nativa del Model Context Protocol (MCP) mediante el SDK oficial (mcp-go), la integración con agentes de IA ahora es Plug & Play. Ya no es necesario crear flujos complejos de lectura REST (GET/POST) ni escribir descripciones manuales gigantes en formato JSON; el servidor se auto-documenta y expone sus esquemas estrictos directamente a la IA.

Para el correcto funcionamiento del ecosistema, configura el Agente con el siguiente System Message:

```Plaintext
Eres un ingeniero experto del NOC de una empresa de telecomunicaciones. Tu trabajo es diagnosticar equipos de red.
TIENES a tu disposición un catálogo de herramientas MCP nativas para diagnosticar switches HPE/Comware.
Cuando el usuario te pida revisar algo en un equipo, NO inventes la respuesta ni asumas estados. DEBES usar obligatoriamente tus herramientas para conectarte al equipo, obtener la salida real y luego explicarle al usuario de forma humana y resumida lo que encontraste.
```

### a) Integración Remota (n8n vía SSE)

Dado que n8n se ejecuta típicamente en un servidor independiente, la conexión se realiza a través del estándar de red de MCP: Server-Sent Events (SSE).

    Crear la Credencial:

        En n8n, ve a Credentials -> Add Credential.

        Busca y selecciona Model Context Protocol (MCP).

        En Transport, selecciona SSE.

        En URL, ingresa la ruta pública y segura de tu servidor añadiendo el endpoint /sse (ej. https://tu-túnel-seguro.app/sse).

        Guarda la credencial.

    Configurar el Agente:

        Agrega un nodo AI Agent y conéctalo a tu modelo de lenguaje (ej. OpenAI, Anthropic).

        En la entrada de Tools del Agente, agrega el nodo nativo Model Context Protocol Tool.

        Selecciona la credencial que creaste en el paso 1.

¡Listo! Al conectar este nodo, n8n negociará la conexión por SSE, descargará instantáneamente las 21 herramientas, sus descripciones semánticas y sus parámetros obligatorios (como interface o switch_ip), inyectándolos directamente en el cerebro de la IA.

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