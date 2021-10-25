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

// EdgeDeviceSpec defines the desired state of EdgeDevice
type EdgeDeviceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of EdgeDevice
	// Important: Run "make" to regenerate code after modifying this file

	// Sku specifies the EdgeDevice's SKU.
	Sku        *string     `json:"sku,omitempty"`
	Connection *Connection `json:"connection,omitempty"`
	Address    *string     `json:"address,omitempty"`
	Protocol   *Protocol   `json:"protocol,omitempty"`

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
	ConnectionEthernet Connection = "Ethernet"
)

// Protocol specifies the EdgeDevice's communication protocol.
type Protocol string

const (
	ProtocolHTTP            Protocol = "HTTP"
	ProtocolUSB             Protocol = "USB"
	ProtocolHTTPCommandline Protocol = "HTTPCommandline"
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
