/*
Copyright 2018 The Kubernetes Authors.

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
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TKEClusterProviderConfig is the Schema for the gceclusterproviderconfigs API
// +k8s:openapi-gen=true
type TKEClusterProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	ClusterName    string `json:"clusterName"`
	ClusterCIDR    string `json:"clusterCIDR"`
	ClusterVersion string `json:"clusterVersion"`
	VpcId          string `json:"vpcId"`
	Region         string `json:"region"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GCEClusterProviderConfigList contains a list of GCEClusterProviderConfig
type TKEClusterProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TKEClusterProviderConfig `json:"items"`
}

func ClusterProviderFromProviderConfig(providerConfig clusterv1.ProviderConfig) (*TKEClusterProviderConfig, error) {
	var config TKEClusterProviderConfig
	if err := yaml.Unmarshal(providerConfig.Value.Raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func init() {
	SchemeBuilder.Register(&TKEClusterProviderConfig{}, &TKEClusterProviderConfigList{})
}
