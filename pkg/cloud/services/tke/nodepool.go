package tke

import (
	"fmt"
	"sort"
	"strings"
	"time"

	infrastructurev1beta1 "github.com/TencentCloud/cluster-api-provider-tencent/api/v1beta1"
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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
)

type NodePoolService struct {
	scope     *scope.TKEManagedMachinePoolScope
	tkeClient *tke.Client
	vpcClient *vpc.Client
	asClient  *as.Client
	cvmClient *cvm.Client
}

const (
	Init                = "INIT"
	Running             = "RUNNING"
	Successful          = "SUCCESSFUL"
	PartiallySuccessful = "PARTIALLY_SUCCESSFUL"
	Failed              = "FAILED"
	Cancelled           = "CANCELLED"
)

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

	s.scope.Info("set nodepool explicitly not ready")
	s.scope.SetNotReady()

	describeReq := tke.NewDescribeClusterNodePoolsRequest()
	describeReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	describeResp, err := s.tkeClient.DescribeClusterNodePools(describeReq)
	if err != nil {
		s.scope.Error(err, "unable to describe node pools")
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

		if s.scope.ManagedMachinePool.Spec.OSName != "" {
			req.NodePoolOs = pointer.StringPtr(s.scope.ManagedMachinePool.Spec.OSName)
		}

		if req.NodePoolOs == nil {
			req.NodePoolOs = pointer.StringPtr(s.scope.TKECluster.Spec.OSName)
		}

		var securityGroups []string

		describeSecurityGroupReq := vpc.NewDescribeSecurityGroupsRequest()
		describeSecurityGroupReq.Filters = []*vpc.Filter{
			{
				Name:   pointer.String("security-group-name"),
				Values: []*string{s.scope.WorkerPoolSecurityGroupName()},
			},
		}

		describeSecurityGroupsResp, err := s.vpcClient.DescribeSecurityGroups(describeSecurityGroupReq)
		if err != nil {
			return errors.Wrap(err, "unable to describe security groups")
		}

		if len(describeSecurityGroupsResp.Response.SecurityGroupSet) == 0 {
			s.scope.V(0).Info("security group not found creating")
			describeVPCReq := vpc.NewDescribeVpcsRequest()
			describeVPCReq.VpcIds = []*string{
				pointer.String(s.scope.TKECluster.Spec.VPCID),
			}

			describeVPCResp, err := s.vpcClient.DescribeVpcs(describeVPCReq)
			if err != nil {
				return err
			}

			if len(describeVPCResp.Response.VpcSet) == 0 {
				return errors.New(fmt.Sprintf("unable to get vpc: %s", s.scope.TKECluster.Spec.VPCID))
			}

			createSecurityGroupReq := vpc.NewCreateSecurityGroupWithPoliciesRequest()
			createSecurityGroupReq.GroupName = s.scope.WorkerPoolSecurityGroupName()
			createSecurityGroupReq.GroupDescription = pointer.String(fmt.Sprintf("worker pool security group"))
			createSecurityGroupReq.SecurityGroupPolicySet = &vpc.SecurityGroupPolicySet{
				Egress: []*vpc.SecurityGroupPolicy{
					{
						PolicyIndex:       pointer.Int64(1),
						Protocol:          pointer.String("ALL"),
						Port:              pointer.String("all"),
						CidrBlock:         pointer.String("0.0.0.0/0"),
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("allow all egress"),
					},
				},
				Ingress: []*vpc.SecurityGroupPolicy{
					{
						PolicyIndex:       pointer.Int64(1),
						Protocol:          pointer.String("ICMP"),
						Port:              pointer.String("all"),
						CidrBlock:         pointer.String("0.0.0.0/0"),
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("allow ICMP"),
					},
					{
						PolicyIndex:       pointer.Int64(2),
						Protocol:          pointer.String("TCP"),
						Port:              pointer.String("22"),
						CidrBlock:         pointer.String("0.0.0.0/0"),
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("allow ssh on worker instances"),
					},
					{
						PolicyIndex:       pointer.Int64(3),
						Protocol:          pointer.String("TCP"),
						Port:              pointer.String("30000-32768"),
						CidrBlock:         pointer.String("0.0.0.0/0"),
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("nodeport services"),
					},
					{
						PolicyIndex:       pointer.Int64(4),
						Protocol:          pointer.String("UDP"),
						Port:              pointer.String("30000-32768"),
						CidrBlock:         pointer.String("0.0.0.0/0"),
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("nodeport services"),
					},
					{
						PolicyIndex:       pointer.Int64(5),
						Protocol:          pointer.String("TCP"),
						Port:              pointer.String("all"),
						CidrBlock:         describeVPCResp.Response.VpcSet[0].CidrBlock,
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("vpc cidr"),
					},
					{
						PolicyIndex:       pointer.Int64(6),
						Protocol:          pointer.String("ALL"),
						Port:              pointer.String("all"),
						CidrBlock:         pointer.String("192.168.0.0/16"),
						Action:            pointer.String("ACCEPT"),
						PolicyDescription: pointer.String("service cidr"),
					},
				},
			}

			createSecurityGroupResp, err := s.vpcClient.CreateSecurityGroupWithPolicies(createSecurityGroupReq)
			if err != nil {
				return errors.Wrap(err, "unable to create worker security group")
			}

			securityGroups = append(securityGroups, *createSecurityGroupResp.Response.SecurityGroup.SecurityGroupId)
		}

		for _, securityGroup := range s.scope.ManagedMachinePool.Spec.SecurityGroups {
			securityGroups = append(securityGroups, securityGroup)
		}

		req.InstanceAdvancedSettings = &tke.InstanceAdvancedSettings{}
		if len(s.scope.ManagedMachinePool.Spec.KeyIDs) == 0 {
			// {"InstanceType":"S3.SMALL1","SecurityGroupIds":["sg-hnpkqgjk"]}
			req.LaunchConfigurePara = common.StringPtr(fmt.Sprintf(`{"InstanceType":"%s","SecurityGroupIds":["%s"]}`,
				s.scope.ManagedMachinePool.Spec.InstanceType,
				strings.Join(securityGroups, `","`)))
		} else {
			// {"InstanceType":"S3.SMALL1","SecurityGroupIds":["sg-hnpkqgjk"]}
			req.LaunchConfigurePara = common.StringPtr(fmt.Sprintf(`{"InstanceType":"%s","SecurityGroupIds":["%s"],"LoginSettings":{"KeyIds":["%s"]}}`,
				s.scope.ManagedMachinePool.Spec.InstanceType,
				strings.Join(securityGroups, `","`),
				strings.Join(s.scope.ManagedMachinePool.Spec.KeyIDs, `","`)))
		}

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

	// Describe Auto Scaling Activity logs
	describeAutoScalingReq := as.NewDescribeAutoScalingActivitiesRequest()
	describeAutoScalingReq.Filters = append(describeAutoScalingReq.Filters, &as.Filter{
		Name: common.StringPtr("auto-scaling-group-id"),
		Values: []*string{
			existingNodePool.AutoscalingGroupId,
		},
	})

	describeAutoScalingRes, err := s.asClient.DescribeAutoScalingActivities(describeAutoScalingReq)
	if err != nil {
		return errors.Wrap(err, "unable to describe activity from autoscaling group")
	}

	activityMap := make(map[clusterv1.ConditionType]as.Activity, 0)
	for _, activity := range describeAutoScalingRes.Response.ActivitySet {

		if a, ok := activityMap[clusterv1.ConditionType(*activity.ActivityType)]; ok {

			t, _ := time.Parse("2006-01-02 15:04", *activity.EndTime)
			t2, _ := time.Parse("2006-01-02 15:04", *a.EndTime)
			if t.After(t2) {
				activityMap[clusterv1.ConditionType(*activity.ActivityType)] = *activity
			}
		} else {
			activityMap[clusterv1.ConditionType(*activity.ActivityType)] = *activity
		}
	}

	for _, activity := range activityMap {
		if *activity.StatusCode == Failed || *activity.StatusCode == PartiallySuccessful {
			s.scope.Error(errors.New(*activity.StatusMessage), "auto-scaling")

			conditions.MarkFalse(s.scope.ManagedMachinePool, clusterv1.ConditionType(*activity.ActivityType), *activity.Cause, infrastructurev1beta1.ConditionSeverityError, *activity.StatusMessage)
		} else {
			conditions.MarkTrue(s.scope.ManagedMachinePool, clusterv1.ConditionType(*activity.ActivityType))
		}
	}

	for _, condition := range s.scope.ManagedMachinePool.Status.Conditions {
		if condition.Status == Failed || condition.Status == PartiallySuccessful {
			s.scope.Error(errors.New(condition.Message), "auto-scaling")
		}
	}

	// Describe Auto Scaling Instances
	describeInstancesReq := as.NewDescribeAutoScalingInstancesRequest()
	describeInstancesReq.Filters = append(describeInstancesReq.Filters, &as.Filter{
		Name: common.StringPtr("auto-scaling-group-id"),
		Values: []*string{
			existingNodePool.AutoscalingGroupId,
		},
	})

	describeInstancesRes, err := s.asClient.DescribeAutoScalingInstances(describeInstancesReq)
	if err != nil {
		return errors.Wrap(err, "unable to describe instances from autoscaling group")
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

	s.scope.ManagedMachinePool.Status.Replicas = int32(*existingNodePool.NodeCountSummary.AutoscalingAdded.Normal)

	switch *existingNodePool.LifeState {
	case "normal":
		if *existingNodePool.NodeCountSummary.AutoscalingAdded.Normal != *existingNodePool.NodeCountSummary.AutoscalingAdded.Total ||
			*existingNodePool.NodeCountSummary.AutoscalingAdded.Total != *existingNodePool.DesiredNodesNum {
			s.scope.SetNotReady()
			return nil
		}
		s.scope.Info("here", "summary", existingNodePool.NodeCountSummary)
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

func timeCompare(i, j *as.Activity) bool {
	return *i.EndTime < *j.EndTime
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
