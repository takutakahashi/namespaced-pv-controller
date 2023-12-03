/*
Copyright 2023.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NamespacedPvSpec defines the desired state of NamespacedPv
type NamespacedPvSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	VolumeName       string                               `json:"volumeName"`
	AccessModes      []corev1.PersistentVolumeAccessMode  `json:"accessModes"`
	StorageClassName string                               `json:"storageClassName"`
	ReclaimPolicy    corev1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy"`
	MountOptions     string                               `json:"mountOptions"`
	Nfs              NFS                                  `json:"nfs,omitempty"`
	Capacity         corev1.ResourceList                  `json:"capacity"`
	// +kubebuilder:default=Filesystem
	VolumeMode corev1.PersistentVolumeMode `json:"volumeMode,omitempty"`
}

// NamespacedPvStatus defines the observed state of NamespacedPv
type NamespacedPvStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NamespacedPv is the Schema for the namespacedpvs API
type NamespacedPv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespacedPvSpec   `json:"spec,omitempty"`
	Status NamespacedPvStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NamespacedPvList contains a list of NamespacedPv
type NamespacedPvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespacedPv `json:"items"`
}

type NFS struct {
	Server   string `json:"server"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

type Capacity struct {
	Storage string `json:"storage"`
}

func init() {
	SchemeBuilder.Register(&NamespacedPv{}, &NamespacedPvList{})
}
