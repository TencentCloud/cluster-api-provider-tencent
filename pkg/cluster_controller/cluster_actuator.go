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

package cluster_controller

import (
	"errors"
	"fmt"
	"github.com/dbdd4us/qcloudapi-sdk-go/ccs"
	"github.com/dbdd4us/qcloudapi-sdk-go/common"
	"golang.org/x/net/context"
	"log"
	"os"
	tkeconfigv1 "sigs.k8s.io/cluster-api-provider-tencent/pkg/apis/tkeproviderconfig/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"
)



type TKEClusterClient struct {
	clusterClient client.Client
}


func NewClusterActuator(m manager.Manager) (*TKEClusterClient, error) {
	return &TKEClusterClient{
		clusterClient: m.GetClient(),
	}, nil
}

func (tke *TKEClusterClient) Reconcile(cluster *clusterv1.Cluster) error {
	clusterConfig, err := tkeconfigv1.ClusterProviderFromProviderConfig(cluster.Spec.ProviderConfig)
	if err != nil {
		var err error = errors.New("Cannot unmarshal machine's providerConfig field")
		fmt.Println(err.Error())
		return err
	}

	credential := common.Credential{
		SecretId:  os.Getenv("SecretId"),
		SecretKey: os.Getenv("SecretKey"),
	}
	opts := common.Opts{
		Region: clusterConfig.Region,
	}

	client, err := ccs.NewClient(credential, opts)
	if err != nil {
		log.Fatal(err)
	}
	args := ccs.CreateEmptyClusterArgs{
		clusterConfig.ClusterName,
		clusterConfig.ClusterCIDR,
		clusterConfig.ClusterVersion,
		clusterConfig.VpcId,
		clusterConfig.Region,
	}
	if cluster.ObjectMeta.Annotations["status"] == "" {
		CreateEmptyClusterResponse, err := client.CreateEmptyCluster(&args)
		if err != nil {
			log.Fatal(err)
		}
		if cluster.ObjectMeta.Annotations == nil {
			cluster.ObjectMeta.Annotations = make(map[string]string)
		}
		cluster.ObjectMeta.Annotations["cluster-id"] = CreateEmptyClusterResponse.Data.ClusterId
		cluster.ObjectMeta.Annotations["status"] = "creating"
		tke.clusterClient.Update(context.Background(), cluster)
		time.Sleep(2 * time.Second)

		for ; ;  {
			dciargs := ccs.DescribeClusterInstancesArgs{
				cluster.ObjectMeta.Annotations["cluster-id"],
			}
			dciins, _ := client.DescribeClusterInstances(&dciargs)
			log.Println(dciins)
			if dciins.CodeDesc == "Success"{
				cluster.ObjectMeta.Annotations["status"] = "created"
				tke.clusterClient.Update(context.Background(), cluster)
				time.Sleep(2 * time.Second)
				break;
			}
			time.Sleep(3*time.Second)
		}
	}
	return nil
}

func (tke *TKEClusterClient) Delete(cluster *clusterv1.Cluster) error {
	clusterConfig, err := tkeconfigv1.ClusterProviderFromProviderConfig(cluster.Spec.ProviderConfig)
	if err != nil {
		var err  = errors.New("Cannot unmarshal machine's providerConfig field")
		fmt.Println(err.Error())
		return err
	}

	credential := common.Credential{
		SecretId:  os.Getenv("SecretId"),
		SecretKey: os.Getenv("SecretKey"),
	}
	opts := common.Opts{
		Region: clusterConfig.Region,
	}

	client, err := ccs.NewClient(credential, opts)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cluster id to be delete")
	log.Println(cluster.ObjectMeta.Annotations["cluster-id"])
	args := ccs.DeleteClusterArgs{
		cluster.ObjectMeta.Annotations["cluster-id"],
	}

	DeleteClusterResponse, err := client.DeleteCluster(&args)
	log.Println(DeleteClusterResponse)
	tke.clusterClient.Update(context.Background(), cluster)
	return nil
}
