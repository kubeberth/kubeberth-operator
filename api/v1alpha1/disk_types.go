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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DiskSourceArchive struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type DiskSourceDisk struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type DiskSource struct {
	Archive *DiskSourceArchive `json:"archive,omitempty"`
	Disk    *DiskSourceDisk    `json:"disk,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DiskSpec defines the desired state of Disk
type DiskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Format=string
	Size string `json:"size"`

	//+optional
	Source *DiskSource `json:"source,omitempty"`
}

// DiskStatus defines the observed state of Disk
type DiskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State    string `json:"state"`
	Size     string `json:"size"`
	Phase    string `json:"phase"`
	Progress string `json:"progress"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description=""
//+kubebuilder:printcolumn:name="Size",type="string",JSONPath=".status.size",description=""
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description=""
//+kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress",description=""
//+kubebuilder:subresource:status

// Disk is the Schema for the disks API
type Disk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DiskSpec   `json:"spec,omitempty"`
	Status DiskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DiskList contains a list of Disk
type DiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Disk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Disk{}, &DiskList{})
}
