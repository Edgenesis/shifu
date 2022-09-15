/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MQTTSetting defines MQTT specific settings when connecting to an EdgeDevice
type MQTTSetting struct {
	MQTTTopic         *string `json:"MQTTTopic,omitempty"`
	MQTTServerAddress *string `json:"MQTTServerAddress,omitempty"`
	MQTTServerSecret  *string `json:"MQTTServerSecret,omitempty"`
}

// OPCUASetting defines OPC UA specific settings when connecting to an OPC UA endpoint
type OPCUASetting struct {
	OPCUAEndpoint                   *string `json:"OPCUAEndpoint,omitempty"`
	SecurityMode                    *string `json:"SecurityMode,omitempty"`
	AuthenticationMode              *string `json:"AuthenticationMode,omitempty"`
	CertificateFileName             *string `json:"CertificateFileName,omitempty"`
	PrivateKeyFileName              *string `json:"PrivateKeyFileName,omitempty"`
	ConfigmapName                   *string `json:"ConfigmapName,omitempty"`
	IssuedToken                     *string `json:"IssuedToken,omitempty"`
	SecurityPolicy                  *string `json:"SecurityPolicy,omitempty"`
	Username                        *string `json:"Username,omitempty"`
	Password                        *string `json:"Password,omitempty"`
	ConnectionTimeoutInMilliseconds *int64  `json:"ConnectionTimeoutInMilliseconds,omitempty"`
}

// SocketSetting defines Socket specific settings when connecting to an EdgeDevice
type SocketSetting struct {
	Encoding    *string `json:"encoding,omitempty"`
	NetworkType *string `json:"networkType,omitempty"`
}

// ProtocolSettings defines protocol settings when connecting to an EdgeDevice
type ProtocolSettings struct {
	MQTTSetting   *MQTTSetting   `json:"MQTTSetting,omitempty"`
	OPCUASetting  *OPCUASetting  `json:"OPCUASetting,omitempty"`
	SocketSetting *SocketSetting `json:"SocketSetting,omitempty"`
}

// EdgeDeviceSpec defines the desired state of EdgeDevice
type EdgeDeviceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of EdgeDevice
	// Important: Run "make" to regenerate code after modifying this file

	// Sku specifies the EdgeDevice's SKU.
	Sku              *string            `json:"sku,omitempty"`
	Connection       *Connection        `json:"connection,omitempty"`
	Address          *string            `json:"address,omitempty"`
	Protocol         *Protocol          `json:"protocol,omitempty"`
	ProtocolSettings *ProtocolSettings  `json:"protocolSettings,omitempty"`
	CustomMetadata   *map[string]string `json:"customMetadata,omitempty"`

	// TODO: add other fields like disconnectTimemoutInSeconds
}

// EdgeDeviceStatus defines the observed state of EdgeDevice
type EdgeDeviceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of EdgeDevice
	// Important: Run "make" to regenerate code after modifying this file

	// TODO: EdgeDeiveIP
	// EdgeDeviceIP is the IP address of the EdgeDevice, if it has native IP support.
	// For non-IP connections, EdgeDeviceIP is the connected EdgeNode's IP address.
	// EdgeDeviceIP *string `json:"edgedeviceip"`

	EdgeDevicePhase *EdgeDevicePhase `json:"edgedevicephase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Connection specifies the EdgeDevice-EdgeNode connection type.
type Connection string

const (
	// ConnectionEthernet String
	ConnectionEthernet Connection = "Ethernet"
)

// Protocol specifies the EdgeDevice's communication protocol.
type Protocol string

// Protocol String
const (
	ProtocolHTTP            Protocol = "HTTP"
	ProtocolHTTPCommandline Protocol = "HTTPCommandline"
	ProtocolMQTT            Protocol = "MQTT"
	ProtocolOPCUA           Protocol = "OPCUA"
	ProtocolSocket          Protocol = "Socket"
	ProtocolUSB             Protocol = "USB"
)

// EdgeDevicePhase is a simple, high-level summary of where the EdgeDevice is in its lifecycle.
type EdgeDevicePhase string

const (
	// EdgeDevicePending means the EdgeDevice has been accepted by the system but not ready yet.
	EdgeDevicePending EdgeDevicePhase = "Pending"
	// EdgeDeviceRunning means the EdgeDevice is able to interact with the system and user applications.
	EdgeDeviceRunning EdgeDevicePhase = "Running"
	// EdgeDeviceFailed means the EdgeDevice is failed state.
	EdgeDeviceFailed EdgeDevicePhase = "Failed"
	// EdgeDeviceUnknown means the EdgeDevice's info could not be obtained.
	// This is typically due to communication failures.
	EdgeDeviceUnknown EdgeDevicePhase = "Unknown"
)

//+kubebuilder:object:root=true

// EdgeDevice is the Schema for the edgedevices API
type EdgeDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EdgeDeviceSpec   `json:"spec,omitempty"`
	Status EdgeDeviceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EdgeDeviceList contains a list of EdgeDevice
type EdgeDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EdgeDevice{}, &EdgeDeviceList{})
}
