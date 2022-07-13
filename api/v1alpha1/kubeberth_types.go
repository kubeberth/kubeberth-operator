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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubeBerthSpec defines the desired state of KubeBerth
type KubeBerthSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Format=string
	StorageClassName string `json:"storageClassName"`
	//+optional
	VolumeMode *corev1.PersistentVolumeMode `json:"volumeMode"`
	//+optional
	ExternalDNSDomainName string `json:"externalDNSDomainName"`
}

// KubeBerthStatus defines the observed state of KubeBerth
type KubeBerthStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	StorageClass      string                       `json:"storageClass"`
	VolumeMode        *corev1.PersistentVolumeMode `json:"volumeMode"`
	ExternalDNSDomain string                       `json:"externalDNSDomain"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="StorageClass",type="string",JSONPath=".status.storageClass",description=""
//+kubebuilder:printcolumn:name="VolumeMode",type="string",JSONPath=".status.volumeMode",description=""
//+kubebuilder:printcolumn:name="ExternalDNSDomain",type="string",JSONPath=".status.externalDNSDomain",description=""
//+kubebuilder:subresource:status

// KubeBerth is the Schema for the kubeberths API
type KubeBerth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeBerthSpec   `json:"spec,omitempty"`
	Status KubeBerthStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubeBerthList contains a list of KubeBerth
type KubeBerthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeBerth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeBerth{}, &KubeBerthList{})
}
