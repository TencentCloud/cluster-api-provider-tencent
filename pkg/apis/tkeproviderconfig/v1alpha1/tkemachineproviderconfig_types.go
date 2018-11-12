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

// TKEMachineProviderConfig is the Schema for the gcemachineproviderconfigs API
// +k8s:openapi-gen=true
// The MachineRole indicates the purpose of the Machine, and will determine
// what software and configuration will be used when provisioning and managing
// the Machine. A single Machine may have more than one role, and the list and
// definitions of supported roles is expected to evolve over time.
//
// Currently, only two roles are supported: Master and Node. In the future, we
// expect user needs to drive the evolution and granularity of these roles,
// with new additions accommodating common cluster patterns, like dedicated
// etcd Machines.
//
//                 +-----------------------+------------------------+
//                 | Master present        | Master absent          |
// +---------------+-----------------------+------------------------|
// | Node present: | Install control plane | Join the cluster as    |
// |               | and be schedulable    | just a node            |
// |---------------+-----------------------+------------------------|
// | Node absent:  | Install control plane | Invalid configuration  |
// |               | and be unschedulable  |                        |
// +---------------+-----------------------+------------------------+

type TKEMachineProviderConfig struct {
	metav1.TypeMeta `json:",inline"`

	//ClusterId     string `json:"clusterId"`
	ZoneId        string `json:"zoneId"`
	Cpu           int    `json:"cpu"`
	Mem           int    `json:"mem"`
	BandwidthType string `json:"bandwidthType"`
	Bandwidth     int    `json:"bandwidth"`
	SubnetId      string `json:"subnetId"`
	StorageSize   int    `json:"storageSize"`
	RootSize      int    `json:"rootSize"`
	//GoodsNum      int    `json:"goodsNum"`
	Password      string `json:"password"`
	IsVpcGateway  int    `json:"isVpcGateway"`
	WanIp         int    `json:"wanIp"`
	OsName        string `json:"osName"`
}

type TKEMachineProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TKEMachineProviderConfig `json:"items"`
}

func MachineProviderFromProviderConfig(providerConfig clusterv1.ProviderConfig) (*TKEMachineProviderConfig, error) {
	var config TKEMachineProviderConfig
	if err := yaml.Unmarshal(providerConfig.Value.Raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func init() {
	SchemeBuilder.Register(&TKEMachineProviderConfig{}, &TKEMachineProviderConfigList{})
}
