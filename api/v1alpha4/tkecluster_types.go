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

package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
)

const (
	// TKEClusterFinalizer allows the controller to clean up resources on delete.
	TKEClusterFinalizer = "tkecluster.infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TKEClusterSpec defines the desired state of TKECluster
type TKEClusterSpec struct {
	// Name of TKECluster
	ClusterName string `json:"clusterName"`

	ClusterID string `json:"clusterID,omitempty"`

	Region string `json:"region"`

	ClusterUUID string `json:"clusterUUID,omitempty"`

	// +optional
	ClusterVersion *string `json:"clusterVersion,omitempty"`

	VPCID string `json:"vpcID"`

	// SecurityGroups specifies
	// +optional
	ImageID string `json:"imageID,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
}

// TKEClusterStatus defines the observed state of TKECluster
type TKEClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// ExternalManagedControlPlane indicates to cluster-api that the control plane
	// is managed by an external service such as AKS, EKS, GKE, etc.
	// +kubebuilder:default=true
	ExternalManagedControlPlane *bool `json:"externalManagedControlPlane,omitempty"`
	// Ready denotes that the  API Server is ready to
	// receive requests and that the VPC infra is ready.
	// +kubebuilder:default=false
	Ready bool `json:"ready"`

	// Initialized denotes whether or not the control plane has the
	// uploaded kubernetes config-map.
	// +optional
	Initialized bool `json:"initialized"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TKECluster is the Schema for the tkeclusters API
type TKECluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TKEClusterSpec   `json:"spec,omitempty"`
	Status TKEClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TKEClusterList contains a list of TKECluster
type TKEClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TKECluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TKECluster{}, &TKEClusterList{})
}
