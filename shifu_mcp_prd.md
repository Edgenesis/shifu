# Shifu MCP Server — Product Requirements Document

## 1. Purpose

AI coding agents (Claude Code, Cursor, etc.) need to build applications that talk to IoT devices managed by Shifu. Today, understanding what devices exist and how to call their APIs requires reading Kubernetes CRDs and ConfigMaps — something AI agents can't do natively.

The Shifu MCP Server is a **development-time tool** that gives coding agents the knowledge to write IoT applications:
1. **Discover** what devices are available in the cluster
2. **Understand** each device's HTTP API — endpoints, methods, request body schemas, response formats
3. **Test** device endpoints to verify real behavior before writing final code

The applications the agent produces do **not** use MCP at runtime. They make standard HTTP calls to DeviceShifu service endpoints inside the Kubernetes cluster. MCP is the lens through which the AI agent learns the API; the app uses the API directly.

The operator sets up devices and configures their metadata (descriptions, endpoint schemas, example payloads). The MCP server reads this metadata from Kubernetes and serves it to AI agents as structured API documentation.

## 2. Scope

**In scope:** Device discovery, API documentation, and device invocation — everything a coding agent needs to write an IoT application.

**Out of scope:** Kubernetes cluster management, device lifecycle (create/update/delete), infrastructure operations. The MCP server is read-only with respect to cluster state; it only writes when calling a device endpoint.

## 3. Architecture

There are two distinct communication planes. Understanding this separation is the key to the entire design.

### 3.1 Two Planes: Development-Time vs Runtime

```
DEVELOPMENT TIME (building the app)          RUNTIME (app running in cluster)
─────────────────────────────────────        ──────────────────────────────────

  Developer                                    App Pod
      │                                           │
      ▼                                           │  HTTP (in-cluster)
  AI Coding Agent                                 │
      │                                           ▼
      │ MCP (stdio)                          DeviceShifu Service
      ▼                                    (e.g. deviceshifu-thermometer
  ┌──────────────┐                           .deviceshifu.svc.cluster.local)
  │  MCP Server  │                                │
  │              │                                ▼
  │  Tools:      │                           DeviceShifu Pod (:8080)
  │  list_devices│                                │
  │  get_device_ │                                ▼
  │    api       │                           Physical Device
  │  call_device │
  │  test_device │
  └──────┬───────┘
         │ reads metadata from
         ▼
    K8s API (CRDs, ConfigMaps)
```

**MCP tools are for the AI agent at development time.** The agent uses them to discover devices and understand their APIs so it can write code.

**The app the agent writes does NOT use MCP.** It makes direct HTTP calls to DeviceShifu service endpoints inside the cluster — standard Kubernetes service DNS.

### 3.2 What Each Plane Does

| | MCP Plane (Development Time) | Device API Plane (Runtime) |
|---|---|---|
| **Who** | AI coding agent (Claude Code, Cursor) | The application code the agent writes |
| **Protocol** | MCP over stdio | HTTP over cluster networking |
| **Target** | MCP Server → K8s API | App Pod → DeviceShifu Service |
| **Purpose** | Discover devices, read API docs, test calls | Production device interaction |
| **Endpoint format** | `call_device("thermometer", "/temperature")` | `GET http://deviceshifu-thermometer.deviceshifu.svc.cluster.local/temperature` |
| **Lifetime** | Only during development session | Runs permanently in cluster |

### 3.3 How It Comes Together

```
┌──────────────────────────────────────────────────────────────────┐
│                       Edge Gateway (K3s)                         │
│                                                                  │
│  ┌──────────────┐   ┌────────────────┐   ┌───────────────────┐  │
│  │ Shifu        │   │ MCP Server     │   │ App Pod           │  │
│  │              │   │ (dev time)     │   │ (runtime)         │  │
│  │ EdgeDevice   ◄───┤                │   │                   │  │
│  │ CRDs         │   │ reads CRDs &   │   │ import requests   │  │
│  │              │   │ ConfigMaps to  │   │ r = requests.get( │  │
│  │ ConfigMaps   ◄───┤ serve API docs │   │  "http://device   │  │
│  │              │   │ to AI agent    │   │   shifu-thermo    │  │
│  │ DeviceShifu  ◄───┤                │   │   .deviceshifu    │  │
│  │ Services     │   │ proxies test   │   │   .svc.cluster    │  │
│  │              │   │ calls for      │   │   .local/temp")   │  │
│  │ DeviceShifu  │   │ agent to try   │   │                   │  │
│  │ Pods (:8080) ◄───┤ endpoints      │   │                   │  │
│  │              │   └───────▲────────┘   │         │         │  │
│  │              │           │            │         │         │  │
│  │              ◄───────────┼────────────┼─────────┘         │  │
│  └──────────────┘           │            └───────────────────┘  │
│                             │                                    │
└─────────────────────────────┼────────────────────────────────────┘
                              │ MCP (stdio)
                        ┌─────┴─────┐
                        │ AI Coding │
                        │ Agent     │
                        └───────────┘
```

The MCP server is **stateless**. All device information is read live from Kubernetes (EdgeDevice CRDs + DeviceShifu ConfigMaps). It never caches or stores device state.

## 4. Device Metadata Model

Operators configure device metadata when creating devices. The MCP server reads this metadata and presents it to AI agents. This requires enriching the existing ConfigMap instruction format with new fields for API documentation.

### 4.1 Current Instruction Structure

```go
// pkg/deviceshifu/deviceshifubase/deviceshifubase_config.go
type DeviceShifuInstruction struct {
    DeviceShifuInstructionProperties []DeviceShifuInstructionProperty `yaml:"argumentPropertyList,omitempty"`
    DeviceShifuProtocolProperties    map[string]string                `yaml:"protocolPropertyList,omitempty"`
    DeviceShifuGatewayProperties     map[string]string                `yaml:"gatewayPropertyList,omitempty"`
}

type DeviceShifuInstructionProperty struct {
    ValueType    string      `yaml:"valueType"`
    ReadWrite    string      `yaml:"readWrite"`
    DefaultValue interface{} `yaml:"defaultValue"`
}
```

### 4.2 Proposed New Fields

Add the following fields to `DeviceShifuInstruction`:

```go
type DeviceShifuInstruction struct {
    // --- existing fields ---
    DeviceShifuInstructionProperties []DeviceShifuInstructionProperty `yaml:"argumentPropertyList,omitempty"`
    DeviceShifuProtocolProperties    map[string]string                `yaml:"protocolPropertyList,omitempty"`
    DeviceShifuGatewayProperties     map[string]string                `yaml:"gatewayPropertyList,omitempty"`

    // --- new fields for API documentation ---
    Description  string `yaml:"description,omitempty"`   // Human/AI readable description
    HTTPMethod   string `yaml:"httpMethod,omitempty"`     // GET, POST, PUT (default: GET)
    ContentType  string `yaml:"contentType,omitempty"`    // Request content type (e.g., application/json)
    ResponseType string `yaml:"responseType,omitempty"`   // Response content type (e.g., application/json, image/jpeg)
    RequestBody  string `yaml:"requestBody,omitempty"`    // JSON schema or example of request body
    ResponseBody string `yaml:"responseBody,omitempty"`   // JSON schema or example of response body

    // --- streaming support ---
    Stream       *DeviceShifuStreamProperties `yaml:"stream,omitempty"` // Present if this endpoint is a continuous stream
}

// DeviceShifuStreamProperties describes a streaming endpoint
type DeviceShifuStreamProperties struct {
    Protocol string `yaml:"protocol"`           // "mjpeg", "rtsp", "websocket", "sse", "chunked"
    Format   string `yaml:"format,omitempty"`   // Media format: "video/h264", "image/jpeg", "application/json"
    URL      string `yaml:"url,omitempty"`       // Direct stream URL if different from the HTTP endpoint (e.g., rtsp://...)
}
```

Add a `name` field to `DeviceShifuInstructionProperty`:

```go
type DeviceShifuInstructionProperty struct {
    Name         string      `yaml:"name,omitempty"`   // NEW: parameter name
    ValueType    string      `yaml:"valueType"`
    ReadWrite    string      `yaml:"readWrite"`
    DefaultValue interface{} `yaml:"defaultValue"`
    Description  string      `yaml:"description,omitempty"`  // NEW: parameter description
    Required     *bool       `yaml:"required,omitempty"`     // NEW: is this required
}
```

Add a `description` field to the EdgeDevice CRD:

```go
type EdgeDeviceSpec struct {
    // --- existing fields ---
    Sku              *string            `json:"sku,omitempty"`
    Connection       *Connection        `json:"connection,omitempty"`
    Address          *string            `json:"address,omitempty"`
    Protocol         *Protocol          `json:"protocol,omitempty"`
    ProtocolSettings *ProtocolSettings  `json:"protocolSettings,omitempty"`
    GatewaySettings  *GatewaySettings   `json:"gatewaySettings,omitempty"`
    CustomMetadata   *map[string]string `json:"customMetadata,omitempty"`

    // --- new field ---
    Description      *string            `json:"description,omitempty"` // What this device is and does
}
```

### 4.3 Example: Fully Documented ConfigMap

This is what an operator would write when setting up a device:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-thermometer-configmap
  namespace: deviceshifu
data:
  driverProperties: |
    driverSku: Omega Thermometer
    driverImage: edgehub/deviceshifu-http-http:nightly

  instructions: |
    instructionSettings:
      defaultTimeoutSeconds: 5
    instructions:
      get_temperature:
        description: "Read current temperature from the sensor"
        httpMethod: "GET"
        responseType: "application/json"
        responseBody: |
          {
            "temperature": 36.5,
            "unit": "celsius",
            "timestamp": "2025-01-01T00:00:00Z"
          }
      set_unit:
        description: "Set the temperature unit (celsius or fahrenheit)"
        httpMethod: "POST"
        contentType: "application/json"
        requestBody: |
          {
            "unit": "fahrenheit"
          }
        responseType: "application/json"
        responseBody: |
          {
            "status": "ok",
            "unit": "fahrenheit"
          }
        argumentPropertyList:
          - name: "unit"
            valueType: "String"
            readWrite: "W"
            defaultValue: "celsius"
            description: "Temperature unit to use"
            required: true
      capture_image:
        description: "Capture a still image from the thermal camera"
        httpMethod: "GET"
        responseType: "image/jpeg"
      status:
        description: "Check if the device is online and responding"
        httpMethod: "GET"
        responseType: "text/plain"
        responseBody: "running"
```

**Example: Camera device with streaming endpoints:**

```yaml
instructions: |
  instructions:
    capture:
      description: "Capture a single still image"
      httpMethod: "GET"
      responseType: "image/jpeg"
    stream:
      description: "Live MJPEG video stream from the camera"
      httpMethod: "GET"
      stream:
        protocol: "mjpeg"
        format: "image/jpeg"
    video_feed:
      description: "H.264 RTSP video feed"
      stream:
        protocol: "rtsp"
        format: "video/h264"
        url: "rtsp://camera1.devices.svc.cluster.local:8554/live"
    status:
      description: "Camera health check"
      httpMethod: "GET"
      responseType: "text/plain"
```

### 4.4 Backward Compatibility

All new fields are optional with `omitempty`. Existing ConfigMaps work unchanged — the MCP server simply returns less documentation for those endpoints. The MCP server treats an endpoint with no new fields as:
- `httpMethod`: `"GET"` (default)
- `description`: `""` (empty)
- `contentType`: not specified
- `responseType`: not specified
- `requestBody`/`responseBody`: not specified

## 5. MCP Tools

Five tools.

### `list_devices`

Returns all devices in the cluster with a summary of what each one is.

**Parameters:** none

**Returns:** array of device summaries

**Data sources:** EdgeDevice CRDs (all namespaces) + DeviceShifu Pod status

```json
[
  {
    "name": "edgedevice-thermometer",
    "namespace": "devices",
    "sku": "Omega Thermometer",
    "description": "Industrial temperature sensor on the factory floor",
    "protocol": "HTTP",
    "phase": "Running",
    "service": "deviceshifu-thermometer.deviceshifu.svc.cluster.local"
  },
  {
    "name": "edgedevice-camera",
    "namespace": "devices",
    "sku": "RTSP Camera",
    "description": "Surveillance camera at loading dock",
    "protocol": "HTTP",
    "phase": "Running",
    "service": "deviceshifu-camera.deviceshifu.svc.cluster.local"
  }
]
```

**Implementation:** List EdgeDevice CRs across namespaces, for each find the matching DeviceShifu pod/service by scanning for `EDGEDEVICE_NAME` env var match. Populate `description` from the new `EdgeDeviceSpec.Description` field.

---

### `get_device_api`

Returns the complete API documentation for a device — everything a coding agent needs to write application code that calls this device at runtime. The `baseURL` in the response is the service endpoint the app should use for direct HTTP calls.

**Parameters:**
- `device_name: string` (required) — name of the EdgeDevice

**Returns:** device details + full API reference

**Data sources:** EdgeDevice CR + DeviceShifu ConfigMap + DeviceShifu Service

```json
{
  "name": "edgedevice-thermometer",
  "description": "Industrial temperature sensor on the factory floor",
  "sku": "Omega Thermometer",
  "protocol": "HTTP",
  "phase": "Running",
  "baseURL": "http://deviceshifu-thermometer.deviceshifu.svc.cluster.local",
  "customMetadata": {
    "vendor": "Omega",
    "location": "Building A, Floor 2"
  },
  "endpoints": [
    {
      "path": "/get_temperature",
      "method": "GET",
      "description": "Read current temperature from the sensor",
      "responseType": "application/json",
      "responseBody": {
        "temperature": 36.5,
        "unit": "celsius",
        "timestamp": "2025-01-01T00:00:00Z"
      },
      "timeoutSeconds": 5
    },
    {
      "path": "/set_unit",
      "method": "POST",
      "description": "Set the temperature unit (celsius or fahrenheit)",
      "contentType": "application/json",
      "requestBody": {
        "unit": "fahrenheit"
      },
      "responseType": "application/json",
      "responseBody": {
        "status": "ok",
        "unit": "fahrenheit"
      },
      "parameters": [
        {
          "name": "unit",
          "type": "String",
          "readWrite": "W",
          "default": "celsius",
          "description": "Temperature unit to use",
          "required": true
        }
      ],
      "timeoutSeconds": 5
    },
    {
      "path": "/status",
      "method": "GET",
      "description": "Check if the device is online and responding",
      "responseType": "text/plain",
      "responseBody": "running",
      "timeoutSeconds": 5
    }
  ]
}
```

**Implementation:**
1. Get EdgeDevice CR → `description`, `sku`, `protocol`, `customMetadata`, `phase`
2. Find DeviceShifu Service → `baseURL`
3. Read DeviceShifu ConfigMap (located via Deployment volume mounts) → parse `instructions` key to build `endpoints` array
4. Each instruction name becomes `path: "/<name>"`
5. Map new fields (`description`, `httpMethod`, `contentType`, `responseType`, `requestBody`, `responseBody`) directly
6. Map `argumentPropertyList` to `parameters`
7. `requestBody`/`responseBody` strings are parsed as JSON if valid, otherwise returned as strings

---

### `call_device`

**Development-time testing tool.** Lets the AI agent make a real call to a device endpoint to verify behavior, inspect actual response format, and debug integration issues — before writing the final application code.

This is NOT how the final app calls devices. The app uses direct HTTP to DeviceShifu service endpoints (see Section 3).

**Parameters:**
- `device_name: string` (required)
- `endpoint: string` (required) — e.g., `/get_temperature`
- `method: string` (optional, default: from `get_device_api` or `GET`)
- `body: string` (optional) — request body
- `headers: map[string]string` (optional)

**Returns:**
```json
{
  "statusCode": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"temperature\": 36.5, \"unit\": \"celsius\"}"
}
```

**Implementation:**
1. Resolve `device_name` → DeviceShifu service URL
2. Validate `endpoint` exists in the device's instruction list (from ConfigMap)
3. Make HTTP request to `http://<service>:<port>/<endpoint>`
4. Return status code, headers, body
5. For binary responses (images, etc.), base64-encode the body

---

### `test_device`

Quick health check — useful for an agent to verify a device is reachable before writing code against it.

**Parameters:**
- `device_name: string` (required)

**Returns:**
```json
{
  "device": "edgedevice-thermometer",
  "healthy": true,
  "phase": "Running",
  "podRunning": true,
  "serviceReachable": true,
  "healthEndpoint": "Device is healthy"
}
```

**Implementation:**
1. Check EdgeDevice CR phase
2. Check DeviceShifu pod is Running
3. Check Service exists and has endpoints
4. HTTP GET to DeviceShifu `/health` endpoint

---

### `get_stream_info`

Returns connection details for streaming endpoints on a device. MCP is request-response and cannot carry a live video/data stream — so this tool gives the AI agent the information it needs to write app code that connects to the stream directly.

**Parameters:**
- `device_name: string` (required)
- `endpoint: string` (optional) — specific streaming endpoint; if omitted, returns all streaming endpoints

**Returns:**
```json
{
  "device": "edgedevice-camera",
  "baseURL": "http://deviceshifu-camera.deviceshifu.svc.cluster.local",
  "streams": [
    {
      "path": "/stream",
      "description": "Live MJPEG video stream from the camera",
      "protocol": "mjpeg",
      "format": "image/jpeg",
      "url": "http://deviceshifu-camera.deviceshifu.svc.cluster.local/stream",
      "sampleCode": {
        "python": "import cv2\ncap = cv2.VideoCapture('http://deviceshifu-camera.deviceshifu.svc.cluster.local/stream')\nwhile True:\n    ret, frame = cap.read()"
      }
    },
    {
      "path": "/video_feed",
      "description": "H.264 RTSP video feed",
      "protocol": "rtsp",
      "format": "video/h264",
      "url": "rtsp://camera1.devices.svc.cluster.local:8554/live",
      "sampleCode": {
        "python": "import cv2\ncap = cv2.VideoCapture('rtsp://camera1.devices.svc.cluster.local:8554/live')"
      }
    }
  ]
}
```

**Implementation:**
1. Resolve device → ConfigMap
2. Filter instructions that have `stream` properties set
3. For each streaming instruction:
   - Use `stream.url` if set, otherwise construct from `baseURL + path`
   - Generate `sampleCode` based on `stream.protocol` (template per protocol type)

**Why this isn't `call_device`:** `call_device` makes a single HTTP request and returns a single response. Streams are continuous — the app needs to open a persistent connection (OpenCV for RTSP/MJPEG, WebSocket client, SSE reader, etc.). The MCP tool's job is to tell the agent *how* to connect, not to be the connection.

---

## 6. Example AI Agent Workflow

**User prompt:** "Build me a Python app that monitors the factory temperature and sends an alert if it goes above 40°C"

**Step 1 — Discovery (MCP tools):**

```
Agent calls: list_devices()
Agent sees:  edgedevice-thermometer — "Industrial temperature sensor on the factory floor"
```

**Step 2 — Learn the API (MCP tools):**

```
Agent calls: get_device_api("edgedevice-thermometer")
Agent learns:
  - baseURL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local
  - GET /get_temperature → {"temperature": 36.5, "unit": "celsius", "timestamp": "..."}
  - POST /set_unit       → body: {"unit": "fahrenheit"}
  - GET /status          → "running"
```

**Step 3 — Test it works (MCP tools):**

```
Agent calls: call_device("edgedevice-thermometer", "/get_temperature")
Agent gets:  {"temperature": 24.1, "unit": "celsius", "timestamp": "2025-06-01T10:30:00Z"}
             → confirms real response matches the documented schema
```

**Step 4 — Write the app (agent writes code, NO MCP involved):**

The agent now writes application code that calls the DeviceShifu service endpoint directly via HTTP. The app runs in-cluster and has no dependency on the MCP server.

```python
import requests, time

# Direct HTTP to DeviceShifu service — this is a standard K8s service endpoint.
# The app calls this URL at runtime. MCP is not involved.
THERMOMETER_URL = "http://deviceshifu-thermometer.deviceshifu.svc.cluster.local"

while True:
    resp = requests.get(f"{THERMOMETER_URL}/get_temperature")
    data = resp.json()
    if data["temperature"] > 40:
        send_alert(f"Temperature alert: {data['temperature']}°C")
    time.sleep(30)
```

**Key point:** The MCP server was used in steps 1-3 to give the AI agent the knowledge it needed. The final application (step 4) talks directly to DeviceShifu over HTTP — it has no awareness of MCP, CRDs, or ConfigMaps. The `baseURL` from `get_device_api` is the service endpoint the app should use.

### Streaming Example

**User prompt:** "Build me a Python app that detects motion from the loading dock camera"

**Steps 1-2 — Discovery (MCP tools):**

```
Agent calls: list_devices()
Agent sees:  edgedevice-camera — "Surveillance camera at loading dock"

Agent calls: get_device_api("edgedevice-camera")
Agent learns:
  - GET /capture → image/jpeg (single frame)
  - GET /stream  → streaming endpoint (mjpeg)
  - /video_feed  → streaming endpoint (rtsp)
```

**Step 3 — Get streaming details (MCP tools):**

```
Agent calls: get_stream_info("edgedevice-camera")
Agent learns:
  - /stream: MJPEG at http://deviceshifu-camera.deviceshifu.svc.cluster.local/stream
  - /video_feed: RTSP at rtsp://camera1.devices.svc.cluster.local:8554/live
  - sample code for connecting with OpenCV
```

**Step 4 — Agent can also grab a test frame to verify the camera works:**

```
Agent calls: call_device("edgedevice-camera", "/capture")
Agent gets:  [base64-encoded JPEG] → confirms camera is working
```

**Step 5 — Write the app (NO MCP involved):**

```python
import cv2

# Direct RTSP connection — app talks to the device stream, not MCP
cap = cv2.VideoCapture("rtsp://camera1.devices.svc.cluster.local:8554/live")
bg_subtractor = cv2.createBackgroundSubtractorMOG2()

while True:
    ret, frame = cap.read()
    mask = bg_subtractor.apply(frame)
    if cv2.countNonZero(mask) > 5000:
        send_alert("Motion detected at loading dock")
```

The MCP tool gave the agent the stream URL and protocol. The app connects directly — MCP is not in the data path.

## 7. Repository Structure

```
cmd/
  shifu-mcp-server/
    main.go                    # Entry point, kubeconfig flag, stdio transport

pkg/
  mcp/
    server/
      server.go                # MCP server setup and tool registration
    tools/
      list_devices.go          # list_devices tool
      get_device_api.go        # get_device_api tool
      call_device.go           # call_device tool
      test_device.go           # test_device tool
    device/
      resolver.go              # EdgeDevice CR → DeviceShifu Service/ConfigMap resolution
      configmap.go             # ConfigMap parser for instruction metadata
```

## 8. Configuration & Deployment

### Local Development (stdio)

Claude Code MCP config:
```json
{
  "shifu": {
    "command": "shifu-mcp-server",
    "args": ["--kubeconfig", "~/.kube/config"]
  }
}
```

### RBAC

The MCP server needs **read-only** access to cluster resources plus the ability to make HTTP calls to DeviceShifu services (in-cluster networking):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: shifu-mcp-server
rules:
  - apiGroups: ["shifu.edgenesis.io"]
    resources: ["edgedevices"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps"]
    verbs: ["get", "list"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list"]
```

## 9. Device Resolution Logic

Shifu splits resources across namespaces. The MCP server must correlate them:

| Resource | Namespace | How to find |
|----------|-----------|-------------|
| EdgeDevice CR | `devices` (configurable) | List all EdgeDevice CRs |
| DeviceShifu Deployment | `deviceshifu` | Find Deployment where env `EDGEDEVICE_NAME` matches EdgeDevice name |
| DeviceShifu ConfigMap | `deviceshifu` | From Deployment's volume mounts |
| DeviceShifu Service | `deviceshifu` | From Deployment's label selector |

This resolution happens in `pkg/mcp/device/resolver.go`. It scans all DeviceShifu deployments once per `list_devices` call and caches the mapping for subsequent `get_device_api` calls within the same request.

## 10. Error Handling

Tools return errors in a consistent format:

```json
{
  "error": "DEVICE_NOT_FOUND",
  "message": "EdgeDevice 'camera1' not found in any namespace"
}
```

| Error | Meaning |
|-------|---------|
| `DEVICE_NOT_FOUND` | No EdgeDevice CR with that name |
| `DEVICE_UNHEALTHY` | Device exists but pod is not running |
| `ENDPOINT_NOT_FOUND` | Endpoint not in device's instruction list |
| `DEVICE_CALL_FAILED` | HTTP call to DeviceShifu returned error |
| `DEVICE_TIMEOUT` | Call exceeded timeout |

## 11. Dependencies

| Dependency | Version | Notes |
|------------|---------|-------|
| Go | 1.25.5 | Match existing project |
| `k8s.io/client-go` | v0.35.1 | Match existing project |
| `github.com/modelcontextprotocol/go-sdk` | v1.4.0+ | Official MCP Go SDK (by Google + Anthropic). v1.x stable API. |

Tool definitions use the official SDK's generics-based pattern — define Go structs for tool inputs, schemas are auto-generated from `jsonschema` struct tags:

```go
type ListDevicesInput struct {
    // no required params
}

type GetDeviceAPIInput struct {
    DeviceName string `json:"device_name" jsonschema:"description=Name of the EdgeDevice"`
}

type CallDeviceInput struct {
    DeviceName string            `json:"device_name" jsonschema:"description=Name of the EdgeDevice"`
    Endpoint   string            `json:"endpoint"    jsonschema:"description=API endpoint path e.g. /get_temperature"`
    Method     string            `json:"method"      jsonschema:"description=HTTP method,enum=GET,enum=POST,enum=PUT"`
    Body       string            `json:"body"        jsonschema:"description=Request body (optional)"`
    Headers    map[string]string `json:"headers"     jsonschema:"description=Request headers (optional)"`
}

type TestDeviceInput struct {
    DeviceName string `json:"device_name" jsonschema:"description=Name of the EdgeDevice"`
}

type GetStreamInfoInput struct {
    DeviceName string `json:"device_name" jsonschema:"description=Name of the EdgeDevice"`
    Endpoint   string `json:"endpoint"    jsonschema:"description=Specific streaming endpoint (optional)"`
}
```

## 12. Changes Required to Shifu Core

These are additive, backward-compatible changes:

1. **`DeviceShifuInstruction`** — add `Description`, `HTTPMethod`, `ContentType`, `ResponseType`, `RequestBody`, `ResponseBody` fields (all `omitempty`)
2. **`DeviceShifuInstructionProperty`** — add `Name`, `Description`, `Required` fields (all `omitempty`)
3. **`EdgeDeviceSpec`** — add `Description` field (`omitempty`)
4. **CRD regeneration** — run `make manifests generate` from `pkg/k8s/crd/` after EdgeDeviceSpec change
5. **Update example ConfigMaps** — enrich a few examples with the new fields as reference

None of these changes affect runtime behavior. DeviceShifu ignores unknown YAML keys and the new struct fields default to zero values.
