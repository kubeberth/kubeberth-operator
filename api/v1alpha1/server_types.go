/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerSpec defines the desired state of Server
type ServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	Running *bool `json:"running,omitempty" optional:"true"`
	//+kubebuilder:validation:Required
	CPU *resource.Quantity `json:"cpu"`
	//+kubebuilder:validation:Required
	Memory *resource.Quantity `json:"memory"`
	//+optional
	MACAddress string `json:"macAddress"`
	//+kubebuilder:validation:Required
	Hostname string `json:"hostname"`
	//+optional
	Hosting string `json:"hosting"`
	//+kubebuilder:validation:Required
	Disk *AttachedDisk `json:"disk"`
	//+optional
	CloudInit *AttachedCloudInit `json:"cloudinit"`
}

// ServerStatus defines the observed state of Server
type ServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State    string `json:"state"`
	CPU      string `json:"cpu"`
	Memory   string `json:"memory"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Hosting  string `json:"hosting"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description=""
//+kubebuilder:printcolumn:name="CPU",type="string",JSONPath=".status.cpu",description=""
//+kubebuilder:printcolumn:name="Memory",type="string",JSONPath=".status.memory",description=""
//+kubebuilder:printcolumn:name="Hostname",type="string",JSONPath=".status.hostname",description=""
//+kubebuilder:printcolumn:name="IP",type="string",JSONPath=".status.ip",description=""
//+kubebuilder:printcolumn:name="Hosting",type="string",JSONPath=".status.hosting",description=""
//+kubebuilder:subresource:status

// Server is the Schema for the servers API
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec   `json:"spec,omitempty"`
	Status ServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServerList contains a list of Server
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
