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
│  │ (instruc-    │                             ▼                 │
│  │  tions)      │                     DeviceShifu Pods          │
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

The MCP server is **stateless**. All device information is read live from Kubernetes (EdgeDevice CRDs + ConfigMaps). It never caches or stores device state.

## 4. Device Metadata Model

The MCP server reads device metadata from two sources — both are **extensions to existing Shifu resources** (no new resource types):

1. **EdgeDevice CRD** (existing, extended) — already has `protocol`, `address`, `sku`. New optional fields: `description` (free-form markdown about the device) and `connectionInfo` (free-form markdown about how apps connect).
2. **DeviceShifu ConfigMap `instructions` key** (existing, extended) — already defines instruction names and `argumentPropertyList`. New optional fields per instruction: `description` (free-form markdown), `readWrite` (R/W/RW), `safe` (bool).

### 4.1 Why Extend Existing Resources

Rather than introducing a new CRD or a new ConfigMap key, the interaction documentation is added to resources that already exist:

- **No new resource type** — no additional CRD, no extra ConfigMap key, no new RBAC rules
- **Natural home** — device-level info (`description`, `connectionInfo`) belongs on the EdgeDevice CRD; per-instruction info (`description`, `readWrite`, `safe`) belongs alongside existing instruction config
- **Simple RBAC** — the controller SA already has access to EdgeDevice CRDs and ConfigMaps; no RBAC changes needed
- **Backward-compatible** — all new fields are optional; existing EdgeDevice CRs and ConfigMaps work unchanged
- **Operator-friendly** — `kubectl edit edgedevice <name>` to add device docs; `kubectl edit configmap <name>` to add instruction docs

### 4.2 EdgeDevice CRD Extensions

Two new optional fields on `EdgeDeviceSpec`:

```go
type EdgeDeviceSpec struct {
    Sku              *string            `json:"sku,omitempty"`
    Connection       *Connection        `json:"connection,omitempty"`
    Address          *string            `json:"address,omitempty"`
    Protocol         *Protocol          `json:"protocol,omitempty"`
    ProtocolSettings *ProtocolSettings  `json:"protocolSettings,omitempty"`
    GatewaySettings  *GatewaySettings   `json:"gatewaySettings,omitempty"`
    CustomMetadata   *map[string]string `json:"customMetadata,omitempty"`

    // New fields for MCP / AI agent documentation
    // Description is a free-form markdown description of the device.
    // +optional
    Description    *string `json:"description,omitempty"`
    // ConnectionInfo is free-form markdown describing how apps connect.
    // +optional
    ConnectionInfo *string `json:"connectionInfo,omitempty"`
}
```

| Field | Description |
|-------|-------------|
| `description` | Free-form markdown. What the device is, where it's deployed, safety notes. |
| `connectionInfo` | Free-form markdown. App-facing connection details: broker URLs, base URLs, auth, code examples. This is how the **application** connects to DeviceShifu — distinct from `address` which is how Shifu connects to the **physical device**. |

Example EdgeDevice CR:

```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-robot-arm
  namespace: devices
spec:
  sku: "FANUC-M20iD"
  connection: Ethernet
  address: "192.168.1.50"
  protocol: MQTT
  description: |
    6-axis industrial robot arm (FANUC M-20iD) on the main assembly line.
    Physical device speaks Siemens S7 PLC protocol. Shifu translates PLC
    registers into MQTT topics — your app publishes commands and subscribes
    to status topics.

    **SAFETY:** This device controls physical machinery. Command interactions
    (`robot-arm/commands/*`) actuate real joints and the gripper. Always
    validate joint angles are within safe ranges before publishing.
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
```

### 4.3 Instruction Extensions

Three new optional fields per instruction in the `instructions` ConfigMap key, alongside the existing `argumentPropertyList`:

| Field | Description |
|-------|-------------|
| `description` | Free-form markdown. Protocol details, message formats, code examples. |
| `readWrite` | `R` = read/subscribe, `W` = write/publish, `RW` = both. |
| `safe` | `true` if this interaction has no side effects. |

These extend `DeviceShifuInstruction`:

```go
type DeviceShifuInstruction struct {
    DeviceShifuInstructionProperties []DeviceShifuInstructionProperty `yaml:"argumentPropertyList,omitempty"`
    DeviceShifuProtocolProperties    map[string]string                `yaml:"protocolPropertyList,omitempty"`
    DeviceShifuGatewayProperties     map[string]string                `yaml:"gatewayPropertyList,omitempty"`

    // New fields for MCP / AI agent documentation
    Description string `yaml:"description,omitempty"`
    ReadWrite   string `yaml:"readWrite,omitempty"` // R, W, RW
    Safe        *bool  `yaml:"safe,omitempty"`
}
```

### 4.4 Examples

Each device needs two things: an EdgeDevice CR (with `description` and `connectionInfo`) and a ConfigMap (with per-instruction `description`, `readWrite`, `safe`).

#### MQTT Device — Robot Arm

EdgeDevice CR:
```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-robot-arm
  namespace: devices
spec:
  sku: "FANUC-M20iD"
  connection: Ethernet
  address: "192.168.1.50"
  protocol: MQTT
  description: |
    6-axis industrial robot arm (FANUC M-20iD) on the main assembly line.
    Shifu translates PLC registers into MQTT topics.

    **SAFETY:** Command interactions (`robot-arm/commands/*`) actuate real
    joints and the gripper. Validate joint angles before publishing.
  connectionInfo: |
    MQTT broker: mqtt://deviceshifu-robot-arm.deviceshifu.svc.cluster.local:1883
    No authentication required. Use QoS 1 for commands, QoS 0 for status.

    ```python
    import paho.mqtt.client as mqtt
    client = mqtt.Client()
    client.connect("deviceshifu-robot-arm.deviceshifu.svc.cluster.local", 1883)
    ```
```

ConfigMap (instructions key):
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
      gripper:
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
      joint_positions:
        readWrite: R
        safe: true
        description: |
          Real-time joint positions. Subscribe to receive continuous updates.
          ## Topic
          `robot-arm/status/joint_positions`
          Published every 100ms. Array is [J1..J6] in degrees.
      emergency_stop:
        readWrite: W
        safe: false
        description: |
          Immediately halt all motion.
          ## Topic
          `robot-arm/commands/emergency_stop`
          Publish any message to trigger E-stop.
```

#### NATS Device — Sensor Array

EdgeDevice CR:
```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-sensor-array
  namespace: devices
spec:
  protocol: NATS
  address: "/dev/ttyUSB0"
  description: |
    Distributed sensor array across the warehouse floor. 24 sensor nodes.
    Shifu translates proprietary RS-485 serial protocol into NATS subjects.
  connectionInfo: |
    NATS server: nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222
    No authentication required. Use NATS wildcards for multiple sensors.

    ```python
    import nats
    nc = await nats.connect("nats://deviceshifu-sensor-array.deviceshifu.svc.cluster.local:4222")
    ```
```

ConfigMap:
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
        readWrite: R
        safe: true
        description: |
          Temperature readings. Subject: `sensors.<node_id>.temperature`
          Wildcard: `sensors.*.temperature` for all nodes.
          Published every 5 seconds per node.
      vibration:
        readWrite: R
        safe: true
        description: |
          Vibration readings. Subject: `sensors.<node_id>.vibration`
          Values above 0.5g indicate potential failure.
      configure_interval:
        readWrite: W
        safe: false
        description: |
          Change reporting interval. Uses NATS request/reply.
          Subject: `sensors.<node_id>.config.interval`
          Valid intervals: 1-60 seconds. Default is 5.
```

#### HTTP Device — Temperature Sensor

EdgeDevice CR:
```yaml
apiVersion: shifu.edgenesis.io/v1alpha1
kind: EdgeDevice
metadata:
  name: edgedevice-thermometer
  namespace: devices
spec:
  protocol: HTTP
  address: "192.168.1.100:502"
  description: |
    Industrial temperature sensor. Calibrated for -40°C to 200°C range.
  connectionInfo: |
    Base URL: http://deviceshifu-thermometer.deviceshifu.svc.cluster.local
    No authentication required.
```

ConfigMap:
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
        readWrite: R
        safe: true
        description: |
          GET /get_temperature
          Response: {"temperature": 36.5, "unit": "celsius"}
          Updates every 3 seconds.
      set_unit:
        readWrite: W
        safe: false
        description: |
          POST /set_unit {"unit": "fahrenheit"}
          Response: {"status": "ok", "unit": "fahrenheit"}
      status:
        readWrite: R
        safe: true
        description: |
          GET /status — returns plain text: `running` or `error: <message>`.
```

### 4.5 Design Principles

**Why no `httpSpec`, `mqttSpec`, `natsSpec` structs?**

Trying to capture every protocol's specifics in typed fields leads to an ever-growing union type. Every new protocol (or new use case within an existing protocol) requires a schema change.

Instead, each interaction has just **two structured hints** the MCP server needs programmatically:
- `readWrite` — R/W/RW (safety classification)
- `safe` — bool (does this interaction have side effects?)

Everything else goes in **free-form `description`** fields. The AI agent is the consumer — it reads prose, markdown, code examples, and message format samples perfectly. It doesn't need rigid JSON schemas.

**One vocabulary for all protocols:**

| HTTP term | MQTT term | NATS term | Shifu term |
|---|---|---|---|
| endpoint | topic | subject | **interaction** |
| request body | message payload | message data | *in description* |
| response | — | reply | *in description* |
| URL path | topic name | subject name | *in description* |

### 4.6 Discovery

The MCP server reads EdgeDevice CRs for device-level metadata (`description`, `connectionInfo`, `protocol`, `phase`) and ConfigMaps for per-interaction documentation (extended `instructions` key). It correlates them by matching the `EDGEDEVICE_NAME` env var in the DeviceShifu Deployment.

### 4.7 Graceful Degradation

If an EdgeDevice CR has no `description` or `connectionInfo`, those fields are omitted from MCP responses. If the `instructions` key uses the existing format (instruction names without `description`, `readWrite`, or `safe`), the MCP server returns instruction names with minimal metadata. Fully backward-compatible.

### 4.8 Changes to Shifu Resources

**EdgeDevice CRD** — two new optional fields in `EdgeDeviceSpec`:

| Field | Type | Description |
|-------|------|-------------|
| `description` | `*string` | Free-form markdown describing the device |
| `connectionInfo` | `*string` | Free-form markdown on how to connect |

Backward-compatible: existing EdgeDevice CRs without these fields work unchanged.

**DeviceShifu ConfigMap (`instructions` key)** — three new optional fields per instruction in `DeviceShifuInstruction`:

| Field | Type | Description |
|-------|------|-------------|
| `description` | `string` | Free-form markdown describing the interaction |
| `readWrite` | `string` | R, W, or RW |
| `safe` | `*bool` | Whether the interaction has side effects |

Backward-compatible: existing ConfigMaps without these fields work unchanged. The existing `argumentPropertyList` and `protocolPropertyList` continue to work as before.

Requires `controller-gen` for the CRD changes; no schema changes for ConfigMap (parsed at runtime).

## 5. MCP Tools

Two tools. The MCP server is a **knowledge layer** — it tells the AI agent everything it needs to write correct device interaction code using the right protocol. It does not provide a generic "call any device" tool because device protocols have fundamentally different interaction patterns (request-response vs publish-subscribe vs streaming).

Device health is reported via `EdgeDevicePhase` (maintained by DeviceShifu itself) — no separate health-check tool is needed. The `phase` field is included in both `list_devices` and `get_device_desc` responses.

### `list_devices`

Returns all devices in the cluster with a summary, including their current `EdgeDevicePhase` status.

**Parameters:** none

**Returns:** array of device summaries

**Data sources:** EdgeDevice CRDs (all namespaces) + DeviceShifu ConfigMaps (`instructions` key)

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

**Implementation:** List EdgeDevice CRs across namespaces, for each read `EdgeDevicePhase` from status and `description`/`protocol` from spec. Find the matching DeviceShifu service by scanning Deployments for `EDGEDEVICE_NAME` env var match. If the EdgeDevice CR has no `description`, the field is omitted.

---

### `get_device_desc`

Returns the full documentation for a device — what it is, how to connect, and all interactions with usage examples. Everything a coding agent needs to write application code. The `protocol` field tells the agent what kind of client to write (HTTP, MQTT, NATS, etc.). The `connectionInfo` tells it how to connect. Each interaction's `description` tells it the specifics.

**Parameters:**
- `device_name: string` (required) — name of the EdgeDevice

**Returns:** device details + full interaction reference

**Data sources:** EdgeDevice CR (`description`, `connectionInfo`, `protocol`, `phase`) + DeviceShifu ConfigMap (`instructions` key) + DeviceShifu Service

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
1. Get EdgeDevice CR → `protocol`, `phase`, `description`, `connectionInfo`
2. Find DeviceShifu Service → `service`
3. Read ConfigMap `instructions` key for the matching DeviceShifu
4. For each instruction, read `description`, `readWrite`, `safe` → populate `interactions`
5. If instructions lack extended fields → return instruction names with minimal metadata (graceful degradation)

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
    configmap.go               # ConfigMap parser for extended instructions metadata
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

| Resource | Namespace | How to find | Data |
|----------|-----------|-------------|------|
| EdgeDevice CR | `devices` (configurable) | List all EdgeDevice CRs | `protocol`, `phase`, `description`, `connectionInfo` |
| DeviceShifu Deployment | `deviceshifu` | Find Deployment where env `EDGEDEVICE_NAME` matches | Links device to service/configmap |
| DeviceShifu Service | `deviceshifu` | From Deployment's label selector | Cluster DNS endpoint |
| DeviceShifu ConfigMap | `deviceshifu` | From Deployment's volume mounts | `instructions` key → per-interaction docs |

**Resolution flow:**

1. List all EdgeDevice CRs across namespaces → device names, `protocol`, `phase`, `description`, `connectionInfo`
2. For each device, scan DeviceShifu Deployments for `EDGEDEVICE_NAME` env var match → find Service
3. Read ConfigMap mounted by the Deployment, parse `instructions` key for per-interaction metadata (`description`, `readWrite`, `safe`)
4. If instructions have extended fields → use them for rich interaction docs
5. If instructions lack extended fields → return instruction names with minimal metadata (graceful degradation)

This resolution happens in `pkg/deviceapi/resolver.go`.

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

**Minimal.** Changes are additive and backward-compatible:

1. **EdgeDevice CRD** — two new optional fields (`description`, `connectionInfo`) in `EdgeDeviceSpec`. Requires `controller-gen` to regenerate CRD manifests. Existing EdgeDevice CRs without these fields work unchanged.
2. **`DeviceShifuInstruction` Go type** — three new optional fields (`description`, `readWrite`, `safe`). Parsed at runtime from the existing `instructions` ConfigMap key. No schema migration needed.
3. **Sidecar container** — added to `shifu-crd-controller-manager` Deployment in `shifu_install.yml`. Runs the MCP server binary alongside the controller.
4. **`configmaps` read** — added to existing `shifu-crd-manager-role` ClusterRole. The controller SA already has access to pods, services, deployments, and edgedevices.
5. **LoadBalancer Service** — added to expose the MCP server's Streamable HTTP port (8443) from the sidecar.
6. **Dockerfile** — `dockerfiles/Dockerfile.mcpServer`
7. **MCP server binary** — `cmd/shifu-mcp-server/main.go`
8. **Device API library** — `pkg/deviceapi/` (reusable by `shifuctl` and other tools)
9. **MCP adapter** — `pkg/mcp/` (wraps `deviceapi` into MCP tool handlers)
10. **Example EdgeDevice CRs and ConfigMaps** — MQTT robot arm, NATS sensor array, HTTP thermometer with extended fields
