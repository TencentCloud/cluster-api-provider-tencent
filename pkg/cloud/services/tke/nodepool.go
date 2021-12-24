package tke

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/cache"
	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/cloud/scope"
	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/util"
	"github.com/pkg/errors"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"k8s.io/utils/pointer"
)

type NodePoolService struct {
	scope     *scope.TKEManagedMachinePoolScope
	tkeClient *tke.Client
	vpcClient *vpc.Client
	asClient  *as.Client
	cvmClient *cvm.Client
}

func NewNodePoolService(scope *scope.TKEManagedMachinePoolScope) (*NodePoolService, error) {
	service := &NodePoolService{
		scope: scope,
	}

	tkeClient, err := tke.NewClient(util.NewCredentialsFromENV().
		WithRegion(scope.TKECluster.Spec.Region).
		AsArgs())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create tke client")
	}

	service.tkeClient = tkeClient

	vpcClient, err := vpc.NewClient(util.NewCredentialsFromENV().
		WithRegion(scope.TKECluster.Spec.Region).
		AsArgs())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vpc client")
	}

	service.vpcClient = vpcClient

	asClient, err := as.NewClient(util.NewCredentialsFromENV().
		WithRegion(scope.TKECluster.Spec.Region).
		AsArgs())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vpc client")
	}

	service.asClient = asClient

	cvmClient, err := cvm.NewClient(util.NewCredentialsFromENV().
		WithRegion(scope.TKECluster.Spec.Region).
		AsArgs())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vpc client")
	}

	service.cvmClient = cvmClient

	return service, nil
}

func (s *NodePoolService) ReconcileNodePool() error {
	s.scope.Info("begin node pool reconcile")

	describeReq := tke.NewDescribeClusterNodePoolsRequest()
	describeReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	describeResp, err := s.tkeClient.DescribeClusterNodePools(describeReq)
	if err != nil {
		return errors.Wrap(err, "unable to describe node pools")
	}

	var existingNodePool *tke.NodePool
	for _, nodepool := range describeResp.Response.NodePoolSet {
		if *nodepool.Name == s.scope.ManagedMachinePool.Name {
			existingNodePool = nodepool
			break
		}
	}

	if existingNodePool == nil || existingNodePool.NodePoolId == nil || *existingNodePool.NodePoolId == "" {
		var subnets []string
		if len(s.scope.ManagedMachinePool.Spec.SubnetIDs) == 0 {
			describeSubnetReq := vpc.NewDescribeSubnetsRequest()
			describeSubnetReq.Filters = append(describeSubnetReq.Filters, &vpc.Filter{
				Name: pointer.StringPtr("vpc-id"),
				Values: []*string{
					pointer.StringPtr(s.scope.TKECluster.Spec.VPCID),
				},
			})

			describeSubnetResp, err := s.vpcClient.DescribeSubnets(describeSubnetReq)
			if err != nil {
				return errors.Wrap(err, "unable to describe subnets")
			}

			for _, subnet := range describeSubnetResp.Response.SubnetSet {
				subnets = append(subnets, *subnet.SubnetId)
			}
			// put first subnet
		} else {
			for _, subnetID := range s.scope.ManagedMachinePool.Spec.SubnetIDs {
				subnets = append(subnets, subnetID)
			}
		}

		req := tke.NewCreateClusterNodePoolRequest()
		req.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
		req.Name = pointer.StringPtr(s.scope.ManagedMachinePool.Name)
		// {"MaxSize":5,"MinSize":3,"DesiredCapacity":3,"VpcId":"vpc-ouipya6w","SubnetIds":["subnet-dzej360x"]}
		asParameter := fmt.Sprintf(`{"MaxSize":%d,"MinSize":%d,"DesiredCapacity":%d,"VpcId":"%s", "SubnetIds":["%s"]}`,
			s.scope.ManagedMachinePool.Spec.MaxSize,
			s.scope.ManagedMachinePool.Spec.MinSize,
			s.DesiredSize(),
			s.scope.TKECluster.Spec.VPCID,
			strings.Join(subnets, `","`))

		s.scope.Info("autoscaling parameter", "asgp", asParameter)

		req.AutoScalingGroupPara = pointer.StringPtr(asParameter)

		req.EnableAutoscale = pointer.BoolPtr(false)

		if s.scope.ManagedMachinePool.Spec.ImageID != "" {
			req.NodePoolOs = pointer.StringPtr(s.scope.ManagedMachinePool.Spec.ImageID)
		}

		req.InstanceAdvancedSettings = &tke.InstanceAdvancedSettings{}
		// {"InstanceType":"S3.SMALL1","SecurityGroupIds":["sg-hnpkqgjk"]}
		req.LaunchConfigurePara = common.StringPtr(fmt.Sprintf(`{"InstanceType":"%s","SecurityGroupIds":["%s"]}`,
			s.scope.ManagedMachinePool.Spec.InstanceType,
			strings.Join(s.scope.ManagedMachinePool.Spec.SecurityGroups, `","`)))

		s.scope.Info("begin creating cluster node pool")
		resp, err := s.tkeClient.CreateClusterNodePool(req)
		if err != nil {
			return errors.Wrap(err, "unable to create cluster node pool")
		}

		describeNodePoolReq := tke.NewDescribeClusterNodePoolDetailRequest()
		describeNodePoolReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
		describeNodePoolReq.NodePoolId = resp.Response.NodePoolId

		nodePoolRes, err := s.tkeClient.DescribeClusterNodePoolDetail(describeNodePoolReq)
		if err != nil {
			return errors.Wrap(err, "unable to get node pool details")
		}

		existingNodePool = nodePoolRes.Response.NodePool
	}

	s.scope.Info("pool status", "image", *existingNodePool.ImageId, "os", *existingNodePool.NodePoolOs, "status", *existingNodePool.LifeState)

	describeInstancesReq := as.NewDescribeAutoScalingInstancesRequest()
	describeInstancesReq.Filters = append(describeInstancesReq.Filters, &as.Filter{
		Name: common.StringPtr("auto-scaling-group-id"),
		Values: []*string{
			existingNodePool.AutoscalingGroupId,
		},
	})

	describeInstancesRes, err := s.asClient.DescribeAutoScalingInstances(describeInstancesReq)
	if err != nil {
		return errors.Wrap(err, "unable to describe instaces from autoscaling group")
	}

	err = cache.InitZoneCache(s.cvmClient)
	if err != nil {
		return errors.Wrap(err, "unable to initialize zone cache")
	}

	providerList := []string{}
	readyInstances := int32(0)
	for _, instance := range describeInstancesRes.Response.AutoScalingInstanceSet {
		if *instance.HealthStatus == "HEALTHY" && *instance.LifeCycleState == "IN_SERVICE" {
			readyInstances += 1
		}
		//  qcloud:///210001/ins-cbki3mw4
		providerList = append(providerList,
			fmt.Sprintf("qcloud:///%s/%s", cache.ZoneNameToID[*instance.Zone], *instance.InstanceId))
	}
	sort.Strings(providerList)
	s.scope.ManagedMachinePool.Spec.ProviderIDList = providerList

	s.scope.ManagedMachinePool.Status.Replicas = readyInstances

	switch *existingNodePool.LifeState {
	case "normal":
		s.scope.SetReady()
	case "creating", "updating", "deleting", "deleted":
		s.scope.SetNotReady()
		s.scope.Info("node pool updating, skip reconcile loop")
		return nil
	}

	if *existingNodePool.MinNodesNum != int64(s.scope.ManagedMachinePool.Spec.MinSize) ||
		*existingNodePool.MaxNodesNum != int64(s.scope.ManagedMachinePool.Spec.MaxSize) {

		s.scope.Info("detected nodepool scale change, updating")

		updateReq := tke.NewModifyClusterNodePoolRequest()
		updateReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
		updateReq.Name = pointer.StringPtr(s.scope.ManagedMachinePool.Name)
		updateReq.NodePoolId = existingNodePool.NodePoolId

		updateReq.MaxNodesNum = pointer.Int64Ptr(int64(s.scope.ManagedMachinePool.Spec.MaxSize))
		updateReq.MinNodesNum = pointer.Int64Ptr(int64(s.scope.ManagedMachinePool.Spec.MinSize))

		_, err := s.tkeClient.ModifyClusterNodePool(updateReq)
		if err != nil {
			return errors.Wrap(err, "unable to upgrade nodepool")
		}
		s.scope.SetNotReady()
	}

	if *existingNodePool.DesiredNodesNum != s.DesiredSize() {
		updateReq := tke.NewModifyNodePoolDesiredCapacityAboutAsgRequest()
		updateReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
		updateReq.NodePoolId = existingNodePool.NodePoolId
		updateReq.DesiredCapacity = pointer.Int64Ptr(s.DesiredSize())

		_, err := s.tkeClient.ModifyNodePoolDesiredCapacityAboutAsg(updateReq)
		if err != nil {
			return errors.Wrap(err, "unable to upgrade nodepool desired capacity")
		}
		s.scope.SetNotReady()
	}

	s.scope.V(2).Info("begin nodepool instance type reconcile")

	launchConfigReq := as.NewDescribeLaunchConfigurationsRequest()
	launchConfigReq.LaunchConfigurationIds = []*string{
		existingNodePool.LaunchConfigurationId,
	}

	res, err := s.asClient.DescribeLaunchConfigurations(launchConfigReq)
	if err != nil {
		return errors.Wrap(err, "unable to describe nodepool launch configuration")
	}

	if *res.Response.LaunchConfigurationSet[0].InstanceType != s.scope.ManagedMachinePool.Spec.InstanceType {
		req := tke.NewModifyNodePoolInstanceTypesRequest()
		req.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
		req.NodePoolId = existingNodePool.NodePoolId
		req.InstanceTypes = []*string{
			pointer.StringPtr(s.scope.ManagedMachinePool.Spec.InstanceType),
		}

		_, err := s.tkeClient.ModifyNodePoolInstanceTypes(req)
		if err != nil {
			return errors.Wrap(err, "unable to modify node pool instanceType")
		}
		s.scope.SetNotReady()
	}
	return nil
}

func (s *NodePoolService) DesiredSize() int64 {
	desiredSize := int64(1)
	if s.scope.MachinePool.Spec.Replicas != nil {
		desiredSize = int64(*s.scope.MachinePool.Spec.Replicas)
	}
	return desiredSize
}

func (s *NodePoolService) DeleteNodePool() error {
	s.scope.Info("begin node pool deletion reconcile")

	describeReq := tke.NewDescribeClusterNodePoolsRequest()
	describeReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	describeResp, err := s.tkeClient.DescribeClusterNodePools(describeReq)
	if err != nil {
		return errors.Wrap(err, "unable to describe node pools")
	}

	var existingNodePool *tke.NodePool
	for _, nodepool := range describeResp.Response.NodePoolSet {
		if *nodepool.Name == s.scope.ManagedMachinePool.Name {
			existingNodePool = nodepool
			break
		}
	}

	if existingNodePool == nil || existingNodePool.NodePoolId == nil || *existingNodePool.NodePoolId == "" {
		s.scope.Info("could not find nodepool maybe deleted already")
		return nil
	}

	deletionRequest := tke.NewDeleteClusterNodePoolRequest()
	deletionRequest.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
	deletionRequest.NodePoolIds = []*string{
		existingNodePool.NodePoolId,
	}
	deletionRequest.KeepInstance = pointer.BoolPtr(false)

	_, err = s.tkeClient.DeleteClusterNodePool(deletionRequest)
	if err != nil {
		return errors.Wrap(err, "unable to delete node pool")
	}

	return nil
}
