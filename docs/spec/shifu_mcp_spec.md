# Shifu MCP Server — Specification

## 1. Purpose

Users of Shifu — specifically engineers and POs — want to "vibe code" edge-connected applications with AI coding assistants (Claude Code, Cursor, etc.). To generate meaningful integration code, the AI needs access to **real device data** — not just metadata.

The Shifu MCP Server is a **development-time knowledge layer** that provides AI agents with the **knowhow** on what devices exist and how to interact with them:
1. **Discover** devices available in the cluster
2. **Understand** interaction contracts — protocols, message formats, connection details, safety characteristics
3. **Generate** applications using the correct protocol for each device

**Key insight:** Shifu doesn't just turn everything into HTTP. Shifu's core job is **protocol transformation** — a physical device speaks its native protocol (PLC, Modbus, OPC UA, etc.) and Shifu transforms it into a developer-friendly app-facing protocol:

```
Physical Device (PLC)    ──►  DeviceShifu  ──►  MQTT topics     (app publishes/subscribes)
Physical Device (Modbus) ──►  DeviceShifu  ──►  HTTP endpoints  (app sends HTTP requests)
Physical Device (Serial) ──►  DeviceShifu  ──►  NATS subjects   (app publishes/subscribes)
Physical Device (RTSP)   ──►  DeviceShifu  ──►  HTTP + RTSP     (app requests + streams)
```

The MCP server **does not provide direct access to DeviceShifu services**. To interact with a device, the app does so via DeviceShifu using the app-facing protocol (MQTT, HTTP, NATS, etc.) inside the K8s cluster. What MCP provides is the knowhow on what the device is and how to use it.

The applications the agent produces do **not** use MCP at runtime. They interact with DeviceShifu services directly using the appropriate protocol. MCP is the lens through which the AI agent learns the device interface; the app uses it directly.

## 2. Scope

**In scope:** Device discovery, interaction documentation (protocol-agnostic), and connectivity testing — everything a coding agent needs to understand devices and write an IoT application, regardless of the app-facing protocol (HTTP, MQTT, NATS, etc.).

**Out of scope:** Direct device access (done via DeviceShifu in-cluster), Kubernetes cluster management, device lifecycle (create/update/delete), infrastructure operations. The MCP server is read-only with respect to cluster state.

## 3. Architecture

There are two distinct communication planes. Understanding this separation is the key to the entire design.

### 3.1 Two Planes: Development-Time vs Runtime

```
DEVELOPMENT TIME (building the app)          RUNTIME (app running in cluster)
─────────────────────────────────────        ──────────────────────────────────

  Developer                                    App Pod
      │                                           │
      ▼                                           │  protocol-specific
  AI Coding Agent                                 │  (HTTP / MQTT / NATS / ...)
      │                                           ▼
      │ MCP (Streamable HTTP)                DeviceShifu Service
      ▼                                    (e.g. MQTT broker, HTTP server,
  ┌──────────────┐                          NATS server — depends on device)
  │  MCP Server  │                                │
  │  (sidecar in │                                ▼
  │  controller  │                           DeviceShifu Pod
  │  pod)        │                                │
  │              │                                ▼
  │  Tools:      │                           Physical Device
  │  list_devices│
  │  get_device_ │
  │    desc      │
  └──────┬───────┘
         │ reads metadata from
         ▼
    K8s API (EdgeDevice CRDs, ConfigMaps, Services)
```

The MCP server runs as a **sidecar container** in the `shifu-crd-controller-manager` Deployment. It shares the controller's ServiceAccount and reads device metadata from the Kubernetes API.

**MCP tools are for the AI agent at development time.** The agent uses them to discover devices and understand their interaction patterns so it can write code.

**The app the agent writes does NOT use MCP.** It talks directly to DeviceShifu services using the appropriate protocol — HTTP requests, MQTT publish/subscribe, NATS publish/subscribe, RTSP streams, etc.

### 3.2 What Each Plane Does

| | MCP Plane (Development Time) | Device Interaction Plane (Runtime) |
|---|---|---|
| **Who** | AI coding agent (Claude Code, Cursor) | The application code the agent writes |
| **Protocol** | MCP over Streamable HTTP | Device-specific: HTTP, MQTT, NATS, RTSP, etc. |
| **Target** | MCP Server → K8s API | App Pod → DeviceShifu Service |
| **Purpose** | Discover devices, read docs | Production device interaction |
| **Example** | `get_device_desc("robot-arm")` → learns MQTT topics | `client.publish("robot-arm/commands/move_joint", ...)` |
| **Lifetime** | Only during development session | Runs permanently in cluster |

### 3.3 How It Comes Together

```
┌──────────────────────────────────────────────────────────────────┐
│                       Edge Gateway (K3s)                         │
│                                                                  │
│  ┌──────────────────────────────┐   ┌───────────────────┐       │
│  │ shifu-crd-controller-manager │   │ App Pod           │       │
│  │                              │   │ (runtime)         │       │
│  │  ┌──────────┐ ┌───────────┐  │   │                   │       │
│  │  │controller│ │MCP Server │  │   │ # MQTT example:   │       │
│  │  │(manager) │ │(sidecar)  │  │   │ client.publish(   │       │
│  │  └──────────┘ └─────┬─────┘  │   │   "robot-arm/     │       │
│  │                     │        │   │   commands/move",  │       │
│  └─────────────────────┼────────┘   │   payload)         │       │
│                        │            │                   │       │
│          reads K8s API │            │ # HTTP example:   │       │
│          ┌─────────────┘            │ requests.get(     │       │
│          ▼                          │  "http://device   │       │
│  ┌──────────────┐                   │   shifu-.../temp")│       │
│  │ K8s API      │                   │         │         │       │
│  │              │                   │         │         │       │
│  │ EdgeDevice   │                   │         │         │       │
│  │ CRDs         │                   └─────────┼─────────┘       │
│  │              │                             │                 │
│  │ ConfigMaps   │                             │                 │
│  │ (interac-    │                             ▼                 │
│  │  tionDocs)   │                     DeviceShifu Pods          │
│  │              │                             │                 │
│  │ Services     │                             ▼                 │
│  └──────────────┘                     Physical Devices          │
│                                                                  │
└──────────────────────────────────────┬───────────────────────────┘
                                       │ MCP (Streamable HTTP)
                                 ┌─────┴─────┐
                                 │ AI Coding │
                                 │ Agent     │
                                 └───────────┘
```

The MCP server is **stateless**. All device information is read live from Kubernetes (EdgeDevice CRDs + ConfigMaps with `interactionDocs` key). It never caches or stores device state.

## 4. Device Metadata Model

The MCP server reads device metadata from two sources:

1. **Operational config** (existing, unchanged) — EdgeDevice CRDs and DeviceShifu ConfigMaps that define instructions, protocol settings, and telemetry. These are mounted into DeviceShifu pods and parsed at runtime.
2. **Interaction documentation** (new) — an `interactionDocs` key added to the existing DeviceShifu ConfigMap, containing free-form descriptions (markdown) of what the device is, how to connect, and how each interaction works. Only read by the MCP server, never by DeviceShifu.

### 4.1 Why a ConfigMap Key

The interaction documentation is structured metadata (device name, interaction names, read/write safety) combined with free-form text (descriptions, usage examples, message formats). A new key in the existing DeviceShifu ConfigMap is the right fit because:

- **No new resource type** — no additional CRD to register, no `controller-gen` changes, no extra RBAC rules for a new resource kind
- **Same lifecycle** — the ConfigMap already exists for each DeviceShifu; adding a key keeps documentation co-located with operational config
- **Simple RBAC** — the MCP server (sidecar in the controller pod) already has access to ConfigMaps via the controller's ServiceAccount; just add `configmaps` read to the existing `shifu-crd-manager-role` ClusterRole
- **Backward-compatible** — the new key is optional; existing ConfigMaps without `interactionDocs` continue to work. DeviceShifu ignores keys it doesn't know about
- **Operator-friendly** — `kubectl edit configmap <device-configmap>` to add documentation; no new `kubectl` resource type to learn

The existing DeviceShifu ConfigMap keys (`instructions`, `driverProperties`, `telemetries`) stay unchanged — they're operational config mounted into pods. The `interactionDocs` key is completely independent. DeviceShifu never reads it.

### 4.2 ConfigMap `interactionDocs` Key — Protocol-Agnostic Design

The `interactionDocs` key uses **one set of field names** for all protocols. The word "interaction" replaces protocol-specific terms like "endpoint" (HTTP), "topic" (MQTT), or "subject" (NATS). Protocol-specific details (how to connect, message formats, code examples) go into free-form `description` fields — the AI agent reads prose perfectly.

The value of `interactionDocs` is a YAML document with this structure:

```yaml
# Inside an existing DeviceShifu ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-robot-arm-configmap
  namespace: deviceshifu
data:
  # Existing keys (unchanged)
  instructions: |
    ...
  driverProperties: |
    ...
  telemetries: |
    ...

  # New key — interaction documentation for the MCP server
  interactionDocs: |
    protocol: MQTT
    description: |
      6-axis industrial robot arm (FANUC M-20iD) on the main assembly line.
      Physical device speaks Siemens S7 PLC protocol. Shifu translates PLC
      registers into MQTT topics — your app publishes commands and subscribes
      to status topics.

      **SAFETY:** This device controls physical machinery. Command interactions
      (`robot-arm/commands/*`) actuate real joints and the gripper.
    connectionInfo: |
      MQTT broker: mqtt://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883
      No authentication required for in-cluster access.
    interactions:
      - name: move_joint
        readWrite: W
        safe: false
        description: |
          Move a specific joint to a target angle.
          ## Topic
          `robot-arm/commands/move_joint`
          ## Message format (JSON)
          {"joint": 1, "angle": 45.0, "speed": 50}
      - name: joint_positions
        readWrite: R
        safe: true
        description: |
          Subscribe to `robot-arm/status/joint_positions` for real-time positions.
```

#### Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `protocol` | No | App-facing protocol Shifu exposes (HTTP, MQTT, NATS, etc.). Read from EdgeDevice if omitted. |
| `description` | No | Free-form device description (markdown). What it is, safety notes, etc. |
| `connectionInfo` | No | Free-form connection instructions (markdown). URLs, broker addresses, auth, code examples. |
| `interactions` | No | Array of interaction objects |
| `interactions[].name` | Yes | Interaction identifier |
| `interactions[].readWrite` | No | `R` = read/subscribe, `W` = write/publish, `RW` = both |
| `interactions[].safe` | No | `true` if this interaction has no side effects (safe to probe) |
| `interactions[].description` | No | Free-form documentation (markdown). Protocol details, message formats, code examples, caveats. |

### 4.3 ConfigMap `interactionDocs` Examples

#### MQTT Device — Robot Arm (PLC transformed to MQTT)

A 6-axis robot arm. Physical device speaks Siemens S7 PLC protocol. Shifu transforms it into MQTT topics. The app is an MQTT client that publishes commands and subscribes to status updates.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-robot-arm-configmap
  namespace: deviceshifu
data:
  instructions: |
    instructions:
      move_joint:
      gripper:
      joint_positions:
      error:
      emergency_stop:
  driverProperties: |
    driverSku: FANUC-M20iD
    driverImage: edgehub/deviceshifu-plc-driver:latest
  interactionDocs: |
    protocol: MQTT
    description: |
      6-axis industrial robot arm (FANUC M-20iD) on the main assembly line.
      Physical device speaks Siemens S7 PLC protocol. Shifu translates PLC
      registers into MQTT topics — your app publishes commands and subscribes
      to status topics.

      **SAFETY:** This device controls physical machinery. Command interactions
      (`robot-arm/commands/*`) actuate real joints and the gripper. Always
      validate joint angles are within safe ranges before publishing.
      Subscribe to `robot-arm/status/error` and handle emergency stop.

    connectionInfo: |
      MQTT broker: mqtt://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883
      No authentication required for in-cluster access.

      Your app should connect as an MQTT client to this broker.
      Use QoS 1 for commands (at-least-once delivery).
      Use QoS 0 for status subscriptions (latest value is fine).

      ## Python (paho-mqtt)
      ```python
      import paho.mqtt.client as mqtt
      client = mqtt.Client()
      client.connect("deviceshifu-robot-arm.deviceshifu.svc.cluster.local", 1883)
      ```

      ## Go
      ```go
      opts := mqtt.NewClientOptions().
          AddBroker("tcp://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883")
      client := mqtt.NewClient(opts)
      client.Connect()
      ```

    interactions:
      - name: move_joint
        readWrite: W
        safe: false
        description: |
          Move a specific joint to a target angle.

          ## Topic
          `robot-arm/commands/move_joint`

          ## Message format (JSON)
          ```json
          {"joint": 1, "angle": 45.0, "speed": 50}
          ```
          - `joint`: 1-6 (axis number)
          - `angle`: degrees. Safe ranges: J1 ±170, J2 -100/+75, J3 -70/+200, J4 ±190, J5 ±125, J6 ±360
          - `speed`: 1-100 (% of max speed)

          Ack on `robot-arm/status/move_ack`:
          ```json
          {"joint": 1, "status": "reached", "actual_angle": 45.0}
          ```

          ## Code example
          ```python
          client.publish("robot-arm/commands/move_joint",
              json.dumps({"joint": 1, "angle": 45.0, "speed": 50}), qos=1)
          ```

      - name: gripper
        readWrite: W
        safe: false
        description: |
          Open or close the gripper.

          ## Topic
          `robot-arm/commands/gripper`

          ## Message format
          ```json
          {"action": "close", "force": 80}
          ```
          - `action`: "open" or "close"
          - `force`: 0-100 (% , only for "close")

          Ack on `robot-arm/status/gripper`:
          ```json
          {"action": "close", "status": "done", "gripping": true}
          ```

      - name: joint_positions
        readWrite: R
        safe: true
        description: |
          Real-time joint positions. Subscribe to receive continuous updates.

          ## Topic
          `robot-arm/status/joint_positions`

          ## Message format
          ```json
          {"joints": [0.0, 45.0, 90.0, 0.0, -30.0, 0.0], "timestamp": "2025-01-15T10:30:00.123Z"}
          ```
          Published every 100ms. Array is [J1, J2, J3, J4, J5, J6] in degrees.

          ```python
          def on_message(client, userdata, msg):
              data = json.loads(msg.payload)
              print(f"Positions: {data['joints']}")

          client.subscribe("robot-arm/status/joint_positions")
          client.on_message = on_message
          client.loop_start()
          ```

      - name: error
        readWrite: R
        safe: true
        description: |
          Error and emergency events.

          ## Topic
          `robot-arm/status/error`

          ## Message format
          ```json
          {"code": "E001", "severity": "critical", "message": "Joint 2 overcurrent"}
          ```
          Severity: "info", "warning", "critical".
          On "critical", the arm stops automatically. Do NOT resume without human confirmation.

      - name: emergency_stop
        readWrite: W
        safe: false
        description: |
          Immediately halt all motion.

          ## Topic
          `robot-arm/commands/emergency_stop`

          Publish any message to trigger E-stop. Requires physical reset to resume.
```

#### NATS Device — Sensor Array (serial transformed to NATS)

A distributed sensor array. Shifu translates proprietary RS-485 serial protocol into NATS subjects. The app is a NATS client.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-sensor-array-configmap
  namespace: deviceshifu
data:
  instructions: |
    instructions:
      temperature:
      vibration:
      configure_interval:
  interactionDocs: |
    protocol: NATS
    description: |
      Distributed sensor array across the warehouse floor. 24 sensor nodes
      measuring temperature, humidity, and vibration. Physical devices use
      proprietary RS-485 serial protocol. Shifu translates into NATS subjects.

      Your app connects as a NATS client. Subscribe to subjects for readings,
      use NATS wildcards to subscribe to multiple sensors at once.

    connectionInfo: |
      NATS server: nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222
      No authentication required for in-cluster access.

      ## Python (nats-py)
      ```python
      import nats
      nc = await nats.connect("nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222")
      ```

      ## Go
      ```go
      nc, _ := nats.Connect("nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222")
      ```

    interactions:
      - name: temperature
        readWrite: R
        safe: true
        description: |
          Temperature readings from sensor nodes.

          ## Subject
          `sensors.<node_id>.temperature`
          Wildcard: `sensors.*.temperature` for all nodes.

          ## Message format
          ```json
          {"node": "node-01", "value": 23.5, "unit": "celsius", "timestamp": "2025-01-15T10:30:00Z"}
          ```
          Published every 5 seconds per node.

          ```python
          async def handler(msg):
              data = json.loads(msg.data)
              print(f"{data['node']}: {data['value']}°C")
          await nc.subscribe("sensors.*.temperature", cb=handler)
          ```

      - name: vibration
        readWrite: R
        safe: true
        description: |
          Vibration readings. Used for predictive maintenance.

          ## Subject
          `sensors.<node_id>.vibration`
          Wildcard: `sensors.*.vibration`

          ## Message format
          ```json
          {"node": "node-01", "axis": {"x": 0.02, "y": 0.01, "z": 0.15}, "unit": "g"}
          ```
          Published every 1 second. Values above 0.5g indicate potential failure.

      - name: configure_interval
        readWrite: W
        safe: false
        description: |
          Change reporting interval for a sensor node. Uses NATS request/reply.

          ## Subject
          `sensors.<node_id>.config.interval`

          ```python
          response = await nc.request("sensors.node-01.config.interval",
              json.dumps({"interval_seconds": 10}).encode(), timeout=5)
          # response.data: {"status": "ok", "interval_seconds": 10}
          ```
          Valid intervals: 1-60 seconds. Default is 5.
```

#### HTTP Device — Temperature Sensor (Modbus transformed to HTTP)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-thermometer-configmap
  namespace: deviceshifu
data:
  instructions: |
    instructions:
      get_temperature:
      set_unit:
      status:
  interactionDocs: |
    protocol: HTTP
    description: |
      Industrial temperature sensor on the factory floor. Reads ambient
      temperature via thermocouple. Calibrated for -40°C to 200°C range.
      Read-only sensor.

    connectionInfo: |
      HTTP on port 8080.
      Base URL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local
      No authentication required for in-cluster access.

    interactions:
      - name: get_temperature
        readWrite: R
        safe: true
        description: |
          Read current temperature.

          ```
          GET /get_temperature
          ```

          Response:
          ```json
          {"temperature": 36.5, "unit": "celsius", "timestamp": "2025-01-01T00:00:00Z"}
          ```
          Updates every 3 seconds.

      - name: set_unit
        readWrite: W
        safe: false
        description: |
          Set temperature unit. Accepts "celsius" or "fahrenheit".

          ```
          POST /set_unit
          Content-Type: application/json

          {"unit": "fahrenheit"}
          ```

          Response: `{"status": "ok", "unit": "fahrenheit"}`

      - name: status
        readWrite: R
        safe: true
        description: |
          ```
          GET /status
          ```
          Returns plain text: `running` or `error: <message>`.
```

#### Camera — HTTP + RTSP

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: deviceshifu-camera-configmap
  namespace: deviceshifu
data:
  instructions: |
    instructions:
      capture:
      stream:
      status:
  interactionDocs: |
    protocol: HTTP
    description: |
      IP camera at loading dock B. Supports still capture via HTTP
      and live streaming via RTSP.

    connectionInfo: |
      HTTP base URL: http://deviceshifu-camera.deviceshifu.svc.cluster.local
      RTSP stream: rtsp://10.0.0.30:554/stream1 (credentials in K8s secret `camera-creds`)
      MJPEG stream: http://10.0.0.30:80/mjpeg (no auth, cluster-internal)

    interactions:
      - name: capture
        readWrite: R
        safe: true
        description: |
          Capture still image.

          ```
          GET /capture
          ```
          Returns JPEG image bytes.

      - name: stream
        readWrite: R
        safe: true
        description: |
          Live video. Connect directly — NOT through DeviceShifu HTTP.

          ## RTSP
          `rtsp://10.0.0.30:554/stream1` — H.264, 1920x1080, 30fps
          ```python
          cap = cv2.VideoCapture("rtsp://admin:password@10.0.0.30:554/stream1")
          ```

          ## MJPEG (simpler, no auth)
          `http://10.0.0.30:80/mjpeg`

      - name: status
        readWrite: R
        safe: true
        description: |
          ```
          GET /status
          ```
          Returns plain text: `running` or `error: <message>`.
```

### 4.4 Design Principles

**Why no `httpSpec`, `mqttSpec`, `natsSpec` structs?**

Trying to capture every protocol's specifics in typed fields leads to an ever-growing union type. Every new protocol (or new use case within an existing protocol) requires a schema change.

Instead, each interaction has just **two structured hints** the MCP server needs programmatically:
- `readWrite` — R/W/RW (safety classification)
- `safe` — bool (does this interaction have side effects?)

Everything else goes in **free-form `description`** fields. The AI agent is the consumer — it reads prose, markdown, code examples, and message format samples perfectly. It doesn't need rigid JSON schemas.

**One vocabulary for all protocols:**

| HTTP term | MQTT term | NATS term | interactionDocs term |
|---|---|---|---|
| endpoint | topic | subject | **interaction** |
| request body | message payload | message data | *in description* |
| response | — | reply | *in description* |
| URL path | topic name | subject name | *in description* |

### 4.5 Discovery

The MCP server reads ConfigMaps for DeviceShifu deployments and parses the `interactionDocs` key. It correlates them to EdgeDevice CRs by matching the `EDGEDEVICE_NAME` env var in the DeviceShifu Deployment.

### 4.6 Graceful Degradation

If no `interactionDocs` key exists in a device's ConfigMap, the MCP server falls back to the operational ConfigMap — it returns instruction names (from the `instructions` key) and any existing `argumentPropertyList` / `protocolPropertyList` data. Functional but less descriptive.

### 4.7 Changes to ConfigMap Format

A new optional key `interactionDocs` is added to the existing DeviceShifu ConfigMap `data:` section. This is **backward-compatible** — existing ConfigMaps without this key continue to work unchanged. DeviceShifu ignores keys it doesn't read.

The `interactionDocs` value is a YAML string containing:

```yaml
interactionDocs: |
  protocol: "MQTT"              # optional, app-facing protocol
  description: "..."            # optional, free-form markdown
  connectionInfo: "..."         # optional, free-form markdown
  interactions:                  # optional, array
    - name: "move_joint"
      readWrite: "W"            # optional: R, W, RW
      safe: false               # optional: bool
      description: "..."        # optional, free-form markdown
```

No changes to Go types, CRD definitions, or `controller-gen` output are required. The MCP server parses this key at read time.

## 5. MCP Tools

Two tools. The MCP server is a **knowledge layer** — it tells the AI agent everything it needs to write correct device interaction code using the right protocol. It does not provide a generic "call any device" tool because device protocols have fundamentally different interaction patterns (request-response vs publish-subscribe vs streaming).

Device health is reported via `EdgeDevicePhase` (maintained by DeviceShifu itself) — no separate health-check tool is needed. The `phase` field is included in both `list_devices` and `get_device_desc` responses.

### `list_devices`

Returns all devices in the cluster with a summary, including their current `EdgeDevicePhase` status.

**Parameters:** none

**Returns:** array of device summaries

**Data sources:** EdgeDevice CRDs (all namespaces) + ConfigMap `interactionDocs`

```json
[
  {
    "name": "edgedevice-robot-arm",
    "namespace": "devices",
    "description": "6-axis robot arm, PLC transformed to MQTT topics",
    "protocol": "MQTT",
    "phase": "Running",
    "service": "deviceshifu-robot-arm.deviceshifu.svc.cluster.local"
  },
  {
    "name": "edgedevice-thermometer",
    "namespace": "devices",
    "description": "Industrial temperature sensor",
    "protocol": "HTTP",
    "phase": "Running",
    "service": "deviceshifu-thermometer.deviceshifu.svc.cluster.local"
  },
  {
    "name": "edgedevice-sensor-array",
    "namespace": "devices",
    "description": "24-node sensor array, serial transformed to NATS",
    "protocol": "NATS",
    "phase": "Running",
    "service": "deviceshifu-sensor-array.deviceshifu.svc.cluster.local"
  }
]
```

**Implementation:** List EdgeDevice CRs across namespaces, for each read `EdgeDevicePhase` from status. Find the matching DeviceShifu service by scanning Deployments for `EDGEDEVICE_NAME` env var match. Populate `description` and `protocol` from the matching ConfigMap `interactionDocs` key. If no `interactionDocs` key exists, `protocol` comes from the EdgeDevice CR and `description` is omitted.

---

### `get_device_desc`

Returns the full documentation for a device — what it is, how to connect, and all interactions with usage examples. Everything a coding agent needs to write application code. The `protocol` field tells the agent what kind of client to write (HTTP, MQTT, NATS, etc.). The `connectionInfo` tells it how to connect. Each interaction's `description` tells it the specifics.

**Parameters:**
- `device_name: string` (required) — name of the EdgeDevice

**Returns:** device details + full interaction reference

**Data sources:** EdgeDevice CR + ConfigMap `interactionDocs` + DeviceShifu Service

**Example: MQTT device (robot arm)**
```json
{
  "name": "edgedevice-robot-arm",
  "description": "6-axis industrial robot arm (FANUC M-20iD)...\n\n**SAFETY:** Command interactions actuate real joints...",
  "protocol": "MQTT",
  "phase": "Running",
  "service": "deviceshifu-robot-arm.deviceshifu.svc.cluster.local",
  "connectionInfo": "MQTT broker: mqtt://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883\n...\n## Python\n```python\nclient.connect(\"deviceshifu-robot-arm...\", 1883)\n```",
  "interactions": [
    {
      "name": "move_joint",
      "readWrite": "W",
      "safe": false,
      "description": "Move a specific joint...\n## Topic\n`robot-arm/commands/move_joint`\n## Message format\n```json\n{\"joint\": 1, \"angle\": 45.0, \"speed\": 50}\n```"
    },
    {
      "name": "joint_positions",
      "readWrite": "R",
      "safe": true,
      "description": "Subscribe to real-time joint positions...\n## Topic\n`robot-arm/status/joint_positions`\n..."
    }
  ]
}
```

**Example: HTTP device (thermometer)**
```json
{
  "name": "edgedevice-thermometer",
  "description": "Industrial temperature sensor...",
  "protocol": "HTTP",
  "phase": "Running",
  "service": "deviceshifu-thermometer.deviceshifu.svc.cluster.local",
  "connectionInfo": "HTTP on port 8080.\nBase URL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local",
  "interactions": [
    {
      "name": "get_temperature",
      "readWrite": "R",
      "safe": true,
      "description": "GET /get_temperature\n\nResponse: {\"temperature\": 36.5, \"unit\": \"celsius\"}"
    }
  ]
}
```

**Example: NATS device (sensor array)**
```json
{
  "name": "edgedevice-sensor-array",
  "description": "24-node sensor array...",
  "protocol": "NATS",
  "phase": "Running",
  "service": "deviceshifu-sensor-array.deviceshifu.svc.cluster.local",
  "connectionInfo": "NATS server: nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222",
  "interactions": [
    {
      "name": "temperature",
      "readWrite": "R",
      "safe": true,
      "description": "## Subject\n`sensors.*.temperature`\n## Message\n```json\n{\"node\": \"node-01\", \"value\": 23.5}\n```"
    }
  ]
}
```

**Implementation:**
1. Get EdgeDevice CR → `protocol`, `phase`
2. Find DeviceShifu Service → `service`
3. Read ConfigMap `interactionDocs` key for the matching DeviceShifu
4. If `interactionDocs` exists → use `description`, `connectionInfo`, `protocol`, `interactions`
5. If no `interactionDocs` → fall back to operational ConfigMap: each instruction name becomes an interaction with minimal metadata

---

## 6. Example AI Agent Workflows

### MQTT — Robot Arm Control App

**User prompt:** "Build me a Python app that moves the robot arm to pick up a part from position A and place it at position B"

**Step 1 — Discovery:**
```
Agent calls: list_devices()
Agent sees:  edgedevice-robot-arm — "6-axis robot arm", protocol: MQTT
```

**Step 2 — Learn the interface:**
```
Agent calls: get_device_desc("edgedevice-robot-arm")
Agent learns:
  - protocol: MQTT
  - connectionInfo: broker at mqtt://deviceshifu-robot-arm...:1883
  - interactions:
    - move_joint (W) — publish to robot-arm/commands/move_joint, JSON: {joint, angle, speed}
    - gripper (W) — publish to robot-arm/commands/gripper, JSON: {action, force}
    - joint_positions (R) — subscribe to robot-arm/status/joint_positions
    - error (R) — subscribe to robot-arm/status/error
    - SAFETY: validate angles, handle emergency stop
```

**Step 3 — Write the app (NO MCP involved):**

```python
import paho.mqtt.client as mqtt
import json, time

BROKER = "deviceshifu-robot-arm.deviceshifu.svc.cluster.local"

client = mqtt.Client()
client.connect(BROKER, 1883)

# Subscribe to position feedback
def on_message(client, userdata, msg):
    if "error" in msg.topic:
        print(f"ERROR: {json.loads(msg.payload)}")

client.subscribe("robot-arm/status/#")
client.on_message = on_message
client.loop_start()

# Pick from position A
client.publish("robot-arm/commands/move_joint",
    json.dumps({"joint": 1, "angle": -45.0, "speed": 50}), qos=1)
time.sleep(3)
client.publish("robot-arm/commands/gripper",
    json.dumps({"action": "close", "force": 80}), qos=1)
time.sleep(1)

# Place at position B
client.publish("robot-arm/commands/move_joint",
    json.dumps({"joint": 1, "angle": 45.0, "speed": 50}), qos=1)
time.sleep(3)
client.publish("robot-arm/commands/gripper",
    json.dumps({"action": "open"}), qos=1)
```

**Key point:** The app is a pure MQTT client. No HTTP. MCP was used to learn the topics and message formats; the app uses MQTT directly.

### NATS — Sensor Dashboard

**User prompt:** "Create a real-time dashboard for all warehouse sensors"

```
Agent calls: get_device_desc("edgedevice-sensor-array")
Agent learns: NATS, server address, wildcard pattern sensors.>
Agent writes: NATS client subscribing to sensors.*.temperature, sensors.*.vibration
              Renders live data from 24 nodes
```

### HTTP — Temperature Monitor

**User prompt:** "Build me a Python app that monitors factory temperature and alerts if it goes above 40°C"

```
Agent calls: get_device_desc("edgedevice-thermometer")
Agent learns: HTTP, base URL, GET /get_temperature
Agent writes: HTTP polling app with requests.get()
```

### Mixed-Protocol App

**User prompt:** "Monitor temperature, capture camera image on alert, and stop the robot arm if temperature exceeds 60°C"

```
Agent calls list_devices() → thermometer (HTTP), camera (HTTP+RTSP), robot-arm (MQTT)
Agent calls get_device_desc for each → learns all protocols

Agent writes app that:
  - Polls HTTP thermometer every 30s
  - On alert: HTTP GET /capture from camera
  - On critical: publishes to robot-arm/commands/emergency_stop via MQTT
  - Three different protocols, correctly handled
```

## 7. Repository Structure

The code is organized as an **API layer + adapter** pattern. The core device resolution logic lives in `pkg/deviceapi/` as a reusable Go library. The MCP server is one adapter; other tools (`shifuctl`, dashboards, future APIs) can consume the same library.

```
cmd/
  shifu-mcp-server/
    main.go                    # Entry point, kubeconfig flag, Streamable HTTP transport

pkg/
  deviceapi/                   # Reusable API layer (Go library)
    api.go                     # ListDevices(), GetDeviceDesc() — public interface
    resolver.go                # EdgeDevice CR → DeviceShifu Service / ConfigMap resolution
    configmap.go               # ConfigMap parser for interactionDocs + instruction metadata (fallback)
    types.go                   # DeviceSummary, DeviceDesc, Interaction structs
  mcp/
    server/
      server.go                # MCP adapter — wraps deviceapi into MCP tool handlers
```

## 8. Configuration & Deployment

### Installation

The MCP server is deployed as a **sidecar container** in the existing `shifu-crd-controller-manager` Deployment. It ships with the standard Shifu install (`shifu_install.yml`) — no separate pod, ServiceAccount, or ClusterRole is needed.

Changes to the existing install:

1. **Sidecar container** added to `shifu-crd-controller-manager` Deployment — runs the MCP server binary
2. **`configmaps` read** added to existing `shifu-crd-manager-role` ClusterRole (the controller SA already has access to pods, services, deployments, and edgedevices)
3. **LoadBalancer Service** added to expose the MCP server's HTTP port from the sidecar

A Dockerfile is added at `dockerfiles/Dockerfile.mcpServer`.

### Connecting an AI agent

The MCP server runs in-cluster as a sidecar and exposes a Streamable HTTP endpoint via a LoadBalancer Service. On K3s, ServiceLB maps this to the gateway's host IP automatically — no port-forward needed.

```bash
claude mcp add shifu --transport http http://<gateway-ip>:8443/mcp
```

### RBAC

The MCP server reuses the existing `shifu-crd-controller-manager` ServiceAccount. The only RBAC change is adding `configmaps` read to the existing `shifu-crd-manager-role` ClusterRole:

```yaml
# Added to existing shifu-crd-manager-role ClusterRole rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]
```

The controller SA already has:
- `pods` — get, list, watch
- `services` — get, list, watch, create, delete, patch, update
- `deployments` — get, list, watch, create, delete, patch, update
- `edgedevices` — get, list, watch, create, delete, patch, update

## 9. Device Resolution Logic

| Resource | Namespace | How to find |
|----------|-----------|-------------|
| EdgeDevice CR | `devices` (configurable) | List all EdgeDevice CRs |
| DeviceShifu Deployment | `deviceshifu` | Find Deployment where env `EDGEDEVICE_NAME` matches |
| DeviceShifu Service | `deviceshifu` | From Deployment's label selector |
| ConfigMap (`interactionDocs`) | `deviceshifu` | From Deployment's volume mounts |
| Operational ConfigMap | `deviceshifu` | Same ConfigMap, `instructions` key (fallback only) |

**Resolution flow:**

1. List all EdgeDevice CRs across namespaces → device names + protocol + phase
2. For each device, scan DeviceShifu Deployments for `EDGEDEVICE_NAME` env var match → find Service
3. Read ConfigMap mounted by the Deployment, parse `interactionDocs` key
4. If `interactionDocs` exists → use it for `description`, `connectionInfo`, `interactions`
5. If no `interactionDocs` → fall back to operational ConfigMap for instruction names

This resolution happens in `pkg/mcp/device/resolver.go`.

## 10. Error Handling

```json
{
  "error": "DEVICE_NOT_FOUND",
  "message": "EdgeDevice 'camera1' not found in any namespace"
}
```

| Error | Meaning |
|-------|---------|
| `DEVICE_NOT_FOUND` | No EdgeDevice CR with that name |

## 11. Dependencies

| Dependency | Version | Notes |
|------------|---------|-------|
| Go | 1.25.5 | Match existing project |
| `k8s.io/client-go` | v0.35.1 | Match existing project |
| `github.com/modelcontextprotocol/go-sdk` | v1.4.0+ | Official MCP Go SDK |

Tool input types:

```go
type ListDevicesInput struct{}

type GetDeviceDescInput struct {
    DeviceName string `json:"device_name" jsonschema:"description=Name of the EdgeDevice"`
}
```

## 12. Changes Required to Shifu Core

**Minimal.** No new CRDs, no changes to Go types, no `controller-gen` modifications. The changes are purely additive:

1. **`interactionDocs` ConfigMap key** — new optional key in existing DeviceShifu ConfigMaps. Backward-compatible; existing ConfigMaps work unchanged. DeviceShifu ignores keys it doesn't read.
2. **Sidecar container** — added to `shifu-crd-controller-manager` Deployment in `shifu_install.yml`. Runs the MCP server binary alongside the controller.
3. **`configmaps` read** — added to existing `shifu-crd-manager-role` ClusterRole. The controller SA already has access to pods, services, deployments, and edgedevices.
4. **LoadBalancer Service** — added to expose the MCP server's Streamable HTTP port (8443) from the sidecar.
5. **Dockerfile** — `dockerfiles/Dockerfile.mcpServer`
6. **MCP server binary** — `cmd/shifu-mcp-server/main.go`
7. **Device API library** — `pkg/deviceapi/` (reusable by `shifuctl` and other tools)
8. **MCP adapter** — `pkg/mcp/` (wraps `deviceapi` into MCP tool handlers)
9. **Example ConfigMaps with `interactionDocs`** — MQTT robot arm, NATS sensor array, HTTP thermometer, camera
