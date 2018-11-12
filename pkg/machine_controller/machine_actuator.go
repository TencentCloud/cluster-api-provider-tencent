package machine_controller

import (
	"github.com/golang/glog"
	tkeconfigv1 "sigs.k8s.io/cluster-api-provider-tencent/pkg/apis/tkeproviderconfig/v1alpha1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"fmt"
	"errors"
	"github.com/dbdd4us/qcloudapi-sdk-go/ccs"
	"github.com/dbdd4us/qcloudapi-sdk-go/common"
	"github.com/ghodss/yaml"
	"golang.org/x/net/context"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"os"
	"time"
)

type TKEProviderDeployer struct {
	Name string
}

const (
	KubeletVersionAnnotationKey      = "kubelet-version"
)


func (a *TKEProviderDeployer) GetIP(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (string, error) {
	return "", nil
}

func (a *TKEProviderDeployer) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	return "", nil
}


func NewMachineActuator(m manager.Manager) (*TKEClient, error) {
	return &TKEClient{
		machineClient: m.GetClient(),
	}, nil
}

type TKEClient struct {
	machineClient client.Client
}

func (tke *TKEClient) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineConfig, err := machineProviderFromProviderConfig(machine.Spec.ProviderConfig)
	if err != nil {
		var err error = errors.New("Cannot unmarshal machine's providerConfig field create machine")
		fmt.Println(err.Error())
		return err
	}
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


	for ; ; {
		if cluster.ObjectMeta.Annotations["status"] == "created"{
			break
		}
		time.Sleep(2*time.Second)
	}

	log.Println(cluster.ObjectMeta.Annotations["cluster-id"])

	if err != nil {
		log.Fatal(err)
	}

	args := ccs.AddClusterInstancesArgs{
		cluster.ObjectMeta.Annotations["cluster-id"],
		machineConfig.ZoneId,
		machineConfig.Cpu,
		machineConfig.Mem,
		machineConfig.BandwidthType,
		machineConfig.Bandwidth,
		machineConfig.SubnetId,
		machineConfig.StorageSize,
		machineConfig.RootSize,
		1,
		machineConfig.Password,
		machineConfig.IsVpcGateway,
		machineConfig.WanIp,
		machineConfig.OsName,
	}
	AddClusterInstancesResponse, err := client.AddClusterInstances(&args)
	if err != nil {
		log.Fatal(err)
	}

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = make(map[string]string)
	}
	log.Println(AddClusterInstancesResponse)
	machine.ObjectMeta.Annotations["instanceIds"] = AddClusterInstancesResponse.Data.InstanceIds[0]
	machine.ObjectMeta.Annotations[KubeletVersionAnnotationKey] = machine.Spec.Versions.Kubelet
	machine.ObjectMeta.Annotations["created"] = "yes"
	tke.machineClient.Update(context.Background(), machine)
	time.Sleep(2 * time.Second)
	return nil
}

func (tke *TKEClient) Delete(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
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
	log.Println("cluster id to be delete")
	log.Println(cluster.ObjectMeta.Annotations["cluster-id"])
	args := ccs.DeleteClusterInstancesArgs{
		cluster.ObjectMeta.Annotations["cluster-id"],
		[]string{machine.ObjectMeta.Annotations["instanceIds"]},
	}

	DeleteClusterInstancesResponse, err := client.DeleteClusterInstances(&args)
	log.Println(DeleteClusterInstancesResponse)
	return nil
}

func (tke *TKEClient) Update(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	glog.Info("hello,this is tencent tkeclient Update")
	return nil
}
func (tke *TKEClient) Exists(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	if machine.ObjectMeta.Annotations["created"] == "" {
		glog.Error("machine not exists")
		return false, nil
	}
	glog.Info("machine exists")
	return true, nil
}

func machineProviderFromProviderConfig(providerConfig clusterv1.ProviderConfig) (*tkeconfigv1.TKEMachineProviderConfig, error) {
	var config tkeconfigv1.TKEMachineProviderConfig
	if err := yaml.Unmarshal(providerConfig.Value.Raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
