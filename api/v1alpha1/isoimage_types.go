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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AttachedISOImage will be used by Server Resource

type AttachedISOImage struct {
	Name string `json:"name"`
}

// ISOImageSpec defines the desired state of ISOImage
type ISOImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	Size string `json:"size"`
	//+kubebuilder:validation:Required
	Repository string `json:"repository"`
}

// ISOImageStatus defines the observed state of ISOImage
type ISOImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State      string `json:"state"`
	Size       string `json:"size"`
	Progress   string `json:"progress"`
	Repository string `json:"repository"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description=""
//+kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress",description=""
//+kubebuilder:printcolumn:name="Size",type="string",JSONPath=".status.size",description=""
//+kubebuilder:printcolumn:name="Repository",type="string",JSONPath=".status.repository",description=""
//+kubebuilder:subresource:status

// ISOImage is the Schema for the isoimages API
type ISOImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ISOImageSpec   `json:"spec,omitempty"`
	Status ISOImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ISOImageList contains a list of ISOImage
type ISOImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ISOImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ISOImage{}, &ISOImageList{})
}
