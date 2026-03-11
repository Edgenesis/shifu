# Shifu MCP Server — Product Requirements Document

## 1. Purpose

Users of ShifuDev and Shifu — specifically engineers and POs — want to "vibe code" edge-connected applications with AI coding assistants (Claude Code, Cursor, etc.). To generate meaningful integration code, the AI needs access to **real device data** — not just metadata.

The Shifu MCP Server is a **development-time knowledge layer** that provides AI agents with the **knowhow** on what devices exist and how to use their DeviceShifu APIs:
1. **Discover** device APIs available in the cluster
2. **Understand** endpoint contracts — methods, request/response schemas, protocol behavior
3. **Test** endpoints to verify connectivity with real responses
4. **Generate** applications using real APIs

The MCP server **does not provide direct access to DeviceShifu APIs**. To call an API to get its info or manipulate a device is done via DeviceShifu APIs provided by Shifu inside the K8s cluster, like before. What MCP provides is the knowhow on what the device is and how to use the DeviceShifu.

The applications the agent produces do **not** use MCP at runtime. They call DeviceShifu service endpoints directly inside the cluster. MCP is the lens through which the AI agent learns the API; the app uses the API directly. Realistically, this makes IoT development into easy cloud app development.

## 2. Scope

**In scope:** Device discovery, API documentation, and endpoint testing — everything a coding agent needs to understand devices and write an IoT application.

**Out of scope:** Direct device API access (done via DeviceShifu APIs in-cluster), Kubernetes cluster management, device lifecycle (create/update/delete), infrastructure operations. The MCP server is read-only with respect to cluster state.

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
      │ MCP (SSE)                            DeviceShifu Service
      ▼                                    (e.g. deviceshifu-thermometer
  ┌──────────────┐                           .deviceshifu.svc.cluster.local)
  │  MCP Server  │                                │
  │              │                                ▼
  │  Tools:      │                           DeviceShifu Pod (:8080)
  │  list_devices│                                │
  │  get_device_ │                                ▼
  │    desc      │                           Physical Device
  │  test_device │
  └──────┬───────┘
         │ reads metadata from
         ▼
    K8s API (EdgeDevice CRDs, DeviceAPIDoc CRs)
```

**MCP tools are for the AI agent at development time.** The agent uses them to discover devices and understand their APIs so it can write code.

**The app the agent writes does NOT use MCP.** It makes direct HTTP calls to DeviceShifu service endpoints inside the cluster — standard Kubernetes service DNS.

### 3.2 What Each Plane Does

| | MCP Plane (Development Time) | Device API Plane (Runtime) |
|---|---|---|
| **Who** | AI coding agent (Claude Code, Cursor) | The application code the agent writes |
| **Protocol** | MCP over SSE | HTTP over cluster networking |
| **Target** | MCP Server → K8s API | App Pod → DeviceShifu Service |
| **Purpose** | Discover devices, read API docs, verify health | Production device interaction |
| **Endpoint format** | `get_device_desc("thermometer")` → learns endpoints | `GET http://deviceshifu-thermometer.deviceshifu.svc.cluster.local/temperature` |
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
│  │ CRDs         │   │ reads CRDs,   │   │ import requests   │  │
│  │              │   │ DeviceAPIDoc  │   │ r = requests.get( │  │
│  │ DeviceAPI-   ◄───┤ CRs, and      │   │  "http://device   │  │
│  │ Doc CRs      │   │ Services to   │   │   shifu-thermo    │  │
│  │              │   │ serve API     │   │   .deviceshifu    │  │
│  │ DeviceShifu  ◄───┤ docs to the   │   │   .svc.cluster    │  │
│  │ Services     │   │ AI agent      │   │   .local/temp")   │  │
│  │              │   │               │   │                   │  │
│  │ DeviceShifu  │   │ proxies test  │   │                   │  │
│  │ Pods (:8080) ◄───┤ calls         │   │                   │  │
│  │              │   └───────▲────────┘   │         │         │  │
│  │              │           │            │         │         │  │
│  │              ◄───────────┼────────────┼─────────┘         │  │
│  └──────────────┘           │            └───────────────────┘  │
│                             │                                    │
└─────────────────────────────┼────────────────────────────────────┘
                              │ MCP (SSE)
                        ┌─────┴─────┐
                        │ AI Coding │
                        │ Agent     │
                        └───────────┘
```

The MCP server is **stateless**. All device information is read live from Kubernetes (EdgeDevice CRDs + DeviceAPIDoc CRs). It never caches or stores device state.

## 4. Device Metadata Model

The MCP server reads device metadata from two sources:

1. **Operational config** (existing, unchanged) — EdgeDevice CRDs and DeviceShifu ConfigMaps that define instructions, protocol settings, and telemetry. These are mounted into DeviceShifu pods and parsed at runtime.
2. **API documentation** (new) — a `DeviceAPIDoc` Custom Resource per device containing endpoint descriptions written as free-form text (markdown). Only read by the MCP server, never by DeviceShifu.

### 4.1 Why a CRD

The API documentation is structured data (device name, endpoint names, HTTP methods) combined with free-form text (descriptions, usage examples). A CRD is the right fit because:

- **Schema validation** — `kubectl apply` rejects typos (e.g., wrong `httpMethod` enum value) immediately, instead of silently producing bad data
- **No YAML-in-YAML** — endpoints are proper typed array items, not string blobs inside a ConfigMap `data` key
- **First-class UX** — `kubectl get apidoc`, `kubectl describe apidoc thermometer` show structured output
- **Owner references** — can auto-delete when the EdgeDevice is removed
- **Shifu already uses CRDs** — operators and the codebase (`controller-gen`, kubebuilder) are set up for this

The existing DeviceShifu ConfigMap stays unchanged — it's operational config mounted into pods. The `DeviceAPIDoc` CR is completely independent. DeviceShifu never sees it.

### 4.2 DeviceAPIDoc CRD

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: deviceapidocs.shifu.edgenesis.io
spec:
  group: shifu.edgenesis.io
  names:
    kind: DeviceAPIDoc
    listKind: DeviceAPIDocList
    plural: deviceapidocs
    singular: deviceapidoc
    shortNames: ["apidoc"]
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: ["deviceName"]
              properties:
                deviceName:
                  type: string
                  description: "Name of the EdgeDevice this documents"
                description:
                  type: string
                  description: "Free-form device description (markdown supported)"
                endpoints:
                  type: array
                  items:
                    type: object
                    required: ["name"]
                    properties:
                      name:
                        type: string
                        description: "Endpoint name — becomes the HTTP path (/<name>)"
                      httpSpec:
                        type: object
                        properties:
                          method:
                            type: string
                            enum: ["GET", "POST", "PUT", "DELETE"]
                          contentType:
                            type: string
                          responseType:
                            type: string
                      stream:
                        type: object
                        properties:
                          protocol:
                            type: string
                          format:
                            type: string
                          url:
                            type: string
                      description:
                        type: string
                        description: "Free-form endpoint documentation (markdown supported)"
      additionalPrinterColumns:
        - name: Device
          type: string
          jsonPath: .spec.deviceName
        - name: Endpoints
          type: integer
          jsonPath: ".spec.endpoints[*].name"
```

### 4.3 DeviceAPIDoc Examples

**Thermometer (HTTP device):**

```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: DeviceAPIDoc
metadata:
  name: edgedevice-thermometer
  namespace: deviceshifu
spec:
  deviceName: edgedevice-thermometer
  description: |
    Industrial temperature sensor on the factory floor, mounted on the
    main assembly line. Reads ambient temperature via thermocouple.
    Calibrated for -40°C to 200°C range.

  endpoints:
    - name: get_temperature
      httpSpec:
        method: GET
        responseType: application/json
      description: |
        Read current temperature from the sensor.

        ## Response
        ```json
        {"temperature": 36.5, "unit": "celsius", "timestamp": "2025-01-01T00:00:00Z"}
        ```

        Temperature updates every 3 seconds. Value is a float in the
        configured unit (default celsius).

    - name: set_unit
      httpSpec:
        method: POST
        contentType: application/json
      description: |
        Set the temperature unit. Accepts "celsius" or "fahrenheit".
        Changes the unit for all subsequent readings from this sensor.
        Does not affect the physical device, only the DeviceShifu response format.

        ## Request
        ```json
        {"unit": "fahrenheit"}
        ```

        ## Response
        ```json
        {"status": "ok", "unit": "fahrenheit"}
        ```

    - name: status
      httpSpec:
        method: GET
        responseType: text/plain
      description: |
        Returns "running" if the device is online and responsive.
```

**Camera (streaming device):**

```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: DeviceAPIDoc
metadata:
  name: edgedevice-camera
  namespace: deviceshifu
spec:
  deviceName: edgedevice-camera
  description: |
    Surveillance camera at loading dock B. Supports still capture
    and live streaming via MJPEG and RTSP.

  endpoints:
    - name: capture
      httpSpec:
        method: GET
        responseType: image/jpeg
      description: |
        Capture a single still image from the camera.
        Returns a JPEG image. Resolution is 1920x1080.

    - name: stream
      httpSpec:
        method: GET
      stream:
        protocol: mjpeg
        format: image/jpeg
      description: |
        Live MJPEG video stream from the camera.
        Connect with `cv2.VideoCapture(url)` or open in a browser.

    - name: video_feed
      stream:
        protocol: rtsp
        format: video/h264
        url: rtsp://camera1.devices.svc.cluster.local:8554/live
      description: |
        H.264 RTSP video feed for high-quality recording.
        Use `cv2.VideoCapture("rtsp://...")` to connect.
        Resolution is 1920x1080 at 30fps.

    - name: status
      httpSpec:
        method: GET
        responseType: text/plain
      description: |
        Returns "running" if the camera is online.
```

### 4.4 Design Principles

**Structured fields** (`httpSpec`, `stream`) are for what the MCP server needs to resolve programmatically — HTTP method, content type, stream protocol/URL.

**Free-form `description`** is for the AI agent. The operator writes whatever the agent needs to understand the endpoint: prose, markdown, code examples, caveats, response schemas. The MCP server passes it through as-is. The AI agent is the consumer — it reads natural language perfectly.

### 4.5 Discovery

The MCP server lists `DeviceAPIDoc` CRs and correlates them to EdgeDevice CRs via `spec.deviceName`.

### 4.6 Graceful Degradation

If no `DeviceAPIDoc` exists for a device, the MCP server falls back to the operational ConfigMap — it returns instruction names (from the `instructions` key) and any existing `argumentPropertyList` / `protocolPropertyList` data. Functional but less descriptive.

### 4.7 Changes to Shifu CRD Types

A new `DeviceAPIDoc` type is added to `pkg/k8s/api/v1alpha1/`. This is a new CRD in the existing `shifu.edgenesis.io` API group — it does not modify the `EdgeDevice` type or any existing code. The CRD definition is generated by `controller-gen` and included in `shifu_install.yml`.

## 5. MCP Tools

Three tools. The MCP server is a **knowledge layer** — it tells the AI agent everything it needs to write correct device integration code. It does not provide a generic "call any device" tool because device protocols have different semantics and safety characteristics.

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
    "name": "edgedevice-humidity",
    "namespace": "devices",
    "sku": "MQTT Humidity Sensor",
    "description": "Humidity sensor publishing to MQTT broker",
    "protocol": "MQTT",
    "phase": "Running",
    "service": "deviceshifu-humidity.deviceshifu.svc.cluster.local"
  }
]
```

**Implementation:** List EdgeDevice CRs across namespaces, for each find the matching DeviceShifu pod/service by scanning for `EDGEDEVICE_NAME` env var match. Populate `description` from the matching `DeviceAPIDoc` CR's `spec.description` (found via `spec.deviceName` match). If no `DeviceAPIDoc` exists, `description` is omitted.

---

### `get_device_desc`

Returns the full API documentation for a device — endpoints, methods, request/response schemas, protocol behavior, and safety info. Everything a coding agent needs to write application code that calls this device at runtime. The `baseURL` in the response is the DeviceShifu service endpoint the app should use for direct calls.

Protocol-specific information is included:
- **For HTTP devices:** standard request-response patterns
- **For video streaming:** stream protocol, URL, format, auth keys if needed
- **For MQTT:** topic name, connection info, message format
- **For OPC UA:** NodeID mappings, read/write safety

Critically, this includes **protocol behavior** so the agent understands how the HTTP proxy layer maps to the underlying device protocol and can write correct code.

**Parameters:**
- `device_name: string` (required) — name of the EdgeDevice

**Returns:** device details + full API reference with protocol context

**Data sources:** EdgeDevice CR + DeviceAPIDoc CR + DeviceShifu Service

**Example: HTTP device**
```json
{
  "name": "edgedevice-thermometer",
  "description": "Industrial temperature sensor on the factory floor, mounted on the\nmain assembly line. Reads ambient temperature via thermocouple.\nCalibrated for -40°C to 200°C range.",
  "sku": "Omega Thermometer",
  "protocol": "HTTP",
  "protocolBehavior": "Direct HTTP proxy — requests are forwarded to the device and responses returned as-is.",
  "phase": "Running",
  "baseURL": "http://deviceshifu-thermometer.deviceshifu.svc.cluster.local",
  "customMetadata": {
    "vendor": "Omega",
    "location": "Building A, Floor 2"
  },
  "endpoints": [
    {
      "path": "/get_temperature",
      "httpSpec": { "method": "GET", "responseType": "application/json" },
      "description": "Read current temperature from the sensor.\n\n## Response\n```json\n{\"temperature\": 36.5, \"unit\": \"celsius\", \"timestamp\": \"2025-01-01T00:00:00Z\"}\n```\n\nTemperature updates every 3 seconds. Value is a float in the\nconfigured unit (default celsius)."
    },
    {
      "path": "/set_unit",
      "httpSpec": { "method": "POST", "contentType": "application/json" },
      "description": "Set the temperature unit. Accepts \"celsius\" or \"fahrenheit\".\nChanges the unit for all subsequent readings from this sensor.\nDoes not affect the physical device, only the DeviceShifu response format.\n\n## Request\n```json\n{\"unit\": \"fahrenheit\"}\n```\n\n## Response\n```json\n{\"status\": \"ok\", \"unit\": \"fahrenheit\"}\n```"
    }
  ]
}
```

**Example: MQTT device**
```json
{
  "name": "edgedevice-humidity",
  "description": "Humidity sensor publishing to MQTT broker",
  "sku": "MQTT Humidity Sensor",
  "protocol": "MQTT",
  "protocolBehavior": "MQTT subscription — each endpoint maps to an MQTT topic. Calling the HTTP endpoint returns the last received message on that topic. The response may be empty if no message has arrived yet. Data is event-driven; the app should poll periodically or use the telemetry service for push-based collection.",
  "phase": "Running",
  "baseURL": "http://deviceshifu-humidity.deviceshifu.svc.cluster.local",
  "endpoints": [
    {
      "path": "/get_humidity",
      "httpSpec": { "method": "GET", "responseType": "application/json" },
      "description": "Last received humidity reading from MQTT topic `/sensors/humidity`.\n\n## Response\n```json\n{\"humidity\": 65.2, \"unit\": \"percent\"}\n```\n\nResponse may be empty if no MQTT message has arrived yet.\nApp should poll periodically."
    }
  ]
}
```

**Example: OPC UA device**
```json
{
  "name": "edgedevice-plc",
  "description": "Siemens PLC controlling conveyor belt",
  "sku": "S7-1200",
  "protocol": "OPCUA",
  "protocolBehavior": "OPC UA node read/write — each endpoint maps to an OPC UA NodeID. Read endpoints return current node values. Write endpoints modify node values and may control physical equipment. Do not call write endpoints without explicit user intent.",
  "phase": "Running",
  "baseURL": "http://deviceshifu-plc.deviceshifu.svc.cluster.local",
  "endpoints": [
    {
      "path": "/get_speed",
      "httpSpec": { "method": "GET", "responseType": "application/json" },
      "description": "Current conveyor belt speed. Maps to OPC UA NodeID `ns=2;i=3`.\n\n## Response\n```json\n{\"speed\": 100, \"unit\": \"rpm\"}\n```"
    },
    {
      "path": "/set_speed",
      "httpSpec": { "method": "POST", "contentType": "application/json" },
      "description": "Set conveyor belt speed. Maps to OPC UA NodeID `ns=2;i=4`.\n\n**CAUTION: controls physical equipment.** Do not call without explicit user intent.\n\n## Request\n```json\n{\"speed\": 100}\n```"
    }
  ]
}
```

**Example: Camera with streaming endpoints**

Streaming info is part of the endpoint list — no separate tool needed. The `stream` field tells the agent the protocol and URL to connect to directly.

```json
{
  "name": "edgedevice-camera",
  "description": "Surveillance camera at loading dock B. Supports still capture and live streaming via MJPEG and RTSP.",
  "protocol": "HTTP",
  "protocolBehavior": "Direct HTTP proxy — requests are forwarded to the device and responses returned as-is.",
  "baseURL": "http://deviceshifu-camera.deviceshifu.svc.cluster.local",
  "endpoints": [
    {
      "path": "/capture",
      "httpSpec": { "method": "GET", "responseType": "image/jpeg" },
      "description": "Capture a single still image from the camera.\nReturns a JPEG image. Resolution is 1920x1080."
    },
    {
      "path": "/stream",
      "httpSpec": { "method": "GET" },
      "stream": { "protocol": "mjpeg", "format": "image/jpeg" },
      "description": "Live MJPEG video stream from the camera.\nConnect with `cv2.VideoCapture(url)` or open in a browser."
    },
    {
      "path": "/video_feed",
      "stream": {
        "protocol": "rtsp",
        "format": "video/h264",
        "url": "rtsp://camera1.devices.svc.cluster.local:8554/live"
      },
      "description": "H.264 RTSP video feed for high-quality recording.\nUse `cv2.VideoCapture(\"rtsp://...\")` to connect.\nResolution is 1920x1080 at 30fps."
    }
  ]
}
```

The agent sees the `stream` field and knows this isn't a request-response endpoint — it writes app code that opens the stream directly. The `description` gives it the usage details in natural language.

**`protocolBehavior` generation:** The MCP server generates this string from the device's `protocol` field in the EdgeDevice CR:

| Protocol | Generated `protocolBehavior` |
|---|---|
| HTTP | Direct HTTP proxy — requests forwarded to device, responses returned as-is. |
| MQTT | MQTT subscription — returns last received message on topic. May be empty. App should poll. |
| OPCUA | OPC UA node read/write. Reads return node values. Writes control physical equipment. |
| Socket | Raw socket — sends bytes, returns response. Semantics are device-specific. |
| TCP | TCP connection — sends bytes, returns response. Semantics are device-specific. |
| LwM2M | LwM2M object read/write. Reads return object values. Writes may control device. |

**Implementation:**
1. Get EdgeDevice CR → `sku`, `protocol`, `customMetadata`, `phase`
2. Generate `protocolBehavior` from `protocol`
3. Find DeviceShifu Service → `baseURL`
4. Look up `DeviceAPIDoc` CR where `spec.deviceName` matches the EdgeDevice name
5. If `DeviceAPIDoc` exists:
   - `spec.description` → `description`
   - `spec.endpoints` → build `endpoints` array: each entry's `name` becomes `path: "/<name>"`, `httpSpec` and `stream` are passed through, `description` (free-form markdown) is passed through as-is
6. If no `DeviceAPIDoc` exists (graceful degradation):
   - Read operational DeviceShifu ConfigMap (located via Deployment volume mounts) → parse `instructions` key
   - Each instruction name becomes `path: "/<name>"` with minimal metadata
   - Derive `httpSpec.method` per endpoint from `argumentPropertyList` entries, or default to `GET`

---

### `test_device`

Health check — optionally reads a safe endpoint to verify connectivity. The agent writes the actual device calls directly in app code — it has all the information it needs from `get_device_desc`.

Does **not** call write endpoints or endpoints with physical side effects.

**Parameters:**
- `device_name: string` (required)
- `probe_endpoint: string` (optional) — a read-safe endpoint to call for e2e verification (e.g., `/status`)

**Returns:**
```json
{
  "device": "edgedevice-thermometer",
  "healthy": true,
  "phase": "Running",
  "podRunning": true,
  "serviceReachable": true,
  "healthEndpoint": "Device is healthy",
  "probe": {
    "endpoint": "/status",
    "statusCode": 200,
    "body": "running"
  }
}
```

**Implementation:**
1. Check EdgeDevice CR phase
2. Check DeviceShifu pod is Running
3. Check Service exists and has endpoints
4. HTTP GET to DeviceShifu `/health` endpoint

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
Agent calls: get_device_desc("edgedevice-thermometer")
Agent learns:
  - protocol: HTTP
  - protocolBehavior: "Direct HTTP proxy — requests forwarded to device..."
  - baseURL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local
  - GET /get_temperature (readWrite: R) → {"temperature": 36.5, "unit": "celsius"}
  - POST /set_unit (readWrite: W)       → body: {"unit": "fahrenheit"}
  - GET /status (readWrite: R)          → "running"
```

**Step 3 — Verify device is healthy (MCP tools):**

```
Agent calls: test_device("edgedevice-thermometer", probe_endpoint="/status")
Agent gets:  healthy=true, probe: 200 "running"
             → device is reachable, safe to write code against it
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

**Key point:** The MCP server was used in steps 1-3 to give the AI agent the knowledge it needed. The final application (step 4) talks directly to DeviceShifu over HTTP — it has no awareness of MCP, CRDs, or ConfigMaps. The `baseURL` from `get_device_desc` is the service endpoint the app should use.

### Streaming Example

**User prompt:** "Build me a Python app that detects motion from the loading dock camera"

**Steps 1-2 — Discovery (MCP tools):**

```
Agent calls: list_devices()
Agent sees:  edgedevice-camera — "Surveillance camera at loading dock"

Agent calls: get_device_desc("edgedevice-camera")
Agent learns:
  - GET /capture → image/jpeg (single frame, request-response)
  - GET /stream  → stream.protocol: "mjpeg", stream.url: "http://...deviceshifu.../stream"
  - /video_feed  → stream.protocol: "rtsp", stream.url: "rtsp://camera1.../live"
```

**Step 3 — Verify camera is healthy (MCP tools):**

```
Agent calls: test_device("edgedevice-camera")
Agent gets:  healthy=true, healthEndpoint: "Device is healthy"
```

**Step 4 — Write the app (NO MCP involved):**

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
    main.go                    # Entry point, kubeconfig flag, SSE transport

pkg/
  k8s/
    api/
      v1alpha1/
        deviceapidoc_types.go  # DeviceAPIDoc CRD type definition
  mcp/
    server/
      server.go                # MCP server setup and tool registration
    tools/
      list_devices.go          # list_devices tool
      get_device_desc.go        # get_device_desc tool (includes streaming endpoint info)
      test_device.go           # test_device tool
    device/
      resolver.go              # EdgeDevice CR → DeviceShifu Service / DeviceAPIDoc resolution
      configmap.go             # ConfigMap parser for instruction metadata (fallback)
```

## 8. Configuration & Deployment

### Installation

The MCP server is deployed as part of the standard Shifu install (`shifu_install.yml`). It runs as a pod in `shifu-system` alongside the controller, with its own ServiceAccount and read-only ClusterRole.

The MCP server Deployment, Service, ServiceAccount, ClusterRole, and ClusterRoleBinding are added to the existing kustomize build in `pkg/k8s/crd/config/`. A Dockerfile is added at `dockerfiles/Dockerfile.mcpServer`.

### Connecting an AI agent

The MCP server runs in-cluster and exposes an SSE endpoint via a LoadBalancer Service. On K3s, ServiceLB maps this to the gateway's host IP automatically — no port-forward needed. On other Kubernetes distributions, configure the LoadBalancer or use NodePort as appropriate.

Developers point their AI agent at the SSE URL:

```bash
claude mcp add shifu --transport sse http://<gateway-ip>:8443/sse
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
    resources: ["edgedevices", "deviceapidocs"]
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
| DeviceShifu Service | `deviceshifu` | From Deployment's label selector |
| DeviceAPIDoc CR | `deviceshifu` | List all DeviceAPIDoc CRs, match via `spec.deviceName` |
| Operational ConfigMap | `deviceshifu` | From Deployment's volume mounts (fallback only — used when no DeviceAPIDoc exists) |

**Resolution flow:**

1. List all EdgeDevice CRs across namespaces → device names + protocol + phase
2. For each device, scan DeviceShifu Deployments for `EDGEDEVICE_NAME` env var match → find Service
3. Look up `DeviceAPIDoc` CR where `spec.deviceName` matches the EdgeDevice name
4. If `DeviceAPIDoc` exists → use it for device description, endpoint `httpSpec`, `stream`, and free-form `description`
5. If no `DeviceAPIDoc` → fall back to operational ConfigMap (from Deployment volume mounts) for instruction names

This resolution happens in `pkg/mcp/device/resolver.go`. It scans all DeviceShifu deployments once per `list_devices` call and caches the mapping for subsequent `get_device_desc` calls within the same request.

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

type GetDeviceDescInput struct {
    DeviceName string `json:"device_name" jsonschema:"description=Name of the EdgeDevice"`
}

type TestDeviceInput struct {
    DeviceName    string `json:"device_name"    jsonschema:"description=Name of the EdgeDevice"`
    ProbeEndpoint string `json:"probe_endpoint" jsonschema:"description=Optional read-safe endpoint to probe for e2e check (e.g. /status)"`
}

```

## 12. Changes Required to Shifu Core

**Minimal.** A new `DeviceAPIDoc` CRD type is added to the existing `shifu.edgenesis.io/v1alpha1` API group. No changes to the `EdgeDevice` type, DeviceShifu structs, controller logic, or existing ConfigMap formats.

New artifacts added to the Shifu install:

1. **`DeviceAPIDoc` CRD type** — `pkg/k8s/api/v1alpha1/deviceapidoc_types.go` + generated deepcopy
2. **`DeviceAPIDoc` CRD definition** — generated by `controller-gen`, added to `shifu_install.yml`
3. **MCP server Deployment + Service** — added to `pkg/k8s/crd/config/` kustomize build
4. **MCP server ServiceAccount + ClusterRole + ClusterRoleBinding** — read-only RBAC (includes `deviceapidocs`)
5. **Dockerfile** — `dockerfiles/Dockerfile.mcpServer`
6. **MCP server binary** — `cmd/shifu-mcp-server/main.go`
7. **MCP server packages** — `pkg/mcp/`
8. **Example DeviceAPIDoc CRs** — added alongside existing device examples in `examples/`
