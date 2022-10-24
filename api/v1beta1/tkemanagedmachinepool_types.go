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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// TKEClusterFinalizer allows the controller to clean up resources on delete.
	TKEManagedMachinePoolFinalizer = "tkemanagedmachinepool.infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TKEManagedMachinePoolSpec defines the desired state of TKEManagedMachinePool
type TKEManagedMachinePoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	MaxSize int32 `json:"maxSize"`

	MinSize int32 `json:"minSize"`

	// SubnetIDs specifies which subnets are used for the
	// +optional
	SubnetIDs []string `json:"subnetIDs,omitempty"`

	// SecurityGroups specifies
	// +optional
	SecurityGroups []string `json:"securityGroups,omitempty"`

	// ProviderIDList are the identification IDs of machine instances provided by the provider.
	// This field must match the provider IDs as seen on the node objects corresponding to a machine pool's machine instances.
	// +optional
	ProviderIDList []string `json:"providerIDList,omitempty"`

	// SecurityGroups specifies
	// +optional
	OSName string `json:"osName,omitempty"`

	InstanceType string `json:"instanceType"`

	KeyIDs []string `json:"keyIDs,omitempty"`
}

// TKEManagedMachinePoolStatus defines the observed state of TKEManagedMachinePool
type TKEManagedMachinePoolStatus struct {
	// +optional
	Ready bool `json:"ready"`

	// Replicas is the most recently observed number of replicas
	// +optional
	Replicas int32 `json:"replicas"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Conditions defines current service state of the managed machine pool
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// TKEManagedMachinePool is the Schema for the tkemanagedmachinepools API
type TKEManagedMachinePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TKEManagedMachinePoolSpec   `json:"spec,omitempty"`
	Status TKEManagedMachinePoolStatus `json:"status,omitempty"`
}

func (r *TKEManagedMachinePool) GetConditions() clusterv1.Conditions {
	return r.Status.Conditions
}

func (r *TKEManagedMachinePool) SetConditions(conditions clusterv1.Conditions) {
	r.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// TKEManagedMachinePoolList contains a list of TKEManagedMachinePool
type TKEManagedMachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TKEManagedMachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TKEManagedMachinePool{}, &TKEManagedMachinePoolList{})
}
