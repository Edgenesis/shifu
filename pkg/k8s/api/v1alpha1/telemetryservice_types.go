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

// HTTPSetting defines HTTP specific settings when connecting to an EdgeDevice
type HTTPSetting struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

// ServiceSettings defines protocol settings when connecting to an EdgeDevice
type ServiceSettings struct {
	HTTPSetting *HTTPSetting `json:"HTTPSetting,omitempty"`
}

// TelemetryServiceSpec defines the desired state of TelemetryService
type TelemetryServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Type            *string            `json:"type,omitempty"`
	Address         *string            `json:"address,omitempty"`
	ServiceSettings *ServiceSettings   `json:"serviceSettings,omitempty"`
	CustomMetadata  *map[string]string `json:"customMetadata,omitempty"`
}

// TelemetryServiceStatus defines the observed state of TelemetryService
type TelemetryServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TelemetryServicePhase *EdgeDevicePhase `json:"telemetryservicephase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TelemetryService is the Schema for the telemetryservices API
type TelemetryService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TelemetryServiceSpec   `json:"spec,omitempty"`
	Status TelemetryServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TelemetryServiceList contains a list of TelemetryService
type TelemetryServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TelemetryService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TelemetryService{}, &TelemetryServiceList{})
}
