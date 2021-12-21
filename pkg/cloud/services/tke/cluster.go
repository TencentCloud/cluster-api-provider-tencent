package tke

import (
	"fmt"
	"os"

	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/cloud/scope"
	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/cluster-api/api/v1alpha4"
)

type Service struct {
	scope     *scope.TKEClusterScope
	tkeClient *tke.Client
}

func NewService(scope *scope.TKEClusterScope) (*Service, error) {
	service := &Service{
		scope: scope,
	}

	tkeClient, err := tke.NewClient(common.NewCredential(
		os.Getenv("TENCENTCLOUD_SECRET_ID"),
		os.Getenv("TENCENTCLOUD_SECRET_KEY"),
	), scope.TKECluster.Spec.Region, profile.NewClientProfile())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create tke client")
	}

	service.tkeClient = tkeClient
	return service, nil
}

func (s *Service) ReconcileCluster(clusterScope *scope.TKEClusterScope) error {
	describerequest := tke.NewDescribeClustersRequest()
	describerequest.Filters = append(describerequest.Filters, &tke.Filter{
		Name: common.StringPtr("tag-key"),
		Values: []*string{
			common.StringPtr("cluster-api-provider-tencet/uuid"),
		},
	}, &tke.Filter{
		Name: common.StringPtr("tag-value"),
		Values: []*string{
			common.StringPtr(s.scope.TKECluster.Spec.ClusterUUID),
		},
	})

	describeResponse, err := s.tkeClient.DescribeClusters(describerequest)
	if err != nil {
		return errors.Wrapf(err, "unable to describe tke cluser using filters: %v", describerequest.Filters)
	}

	if *describeResponse.Response.TotalCount > 0 {
		existingCluster := describeResponse.Response.Clusters[0]
		s.scope.Info("cluster already creating skip", "status", existingCluster.ClusterStatus)

		if s.scope.TKECluster.Spec.ClusterID == "" {
			s.scope.TKECluster.Spec.ClusterID = *existingCluster.ClusterId

			if err := s.scope.PatchObject(); err != nil {
				return errors.Wrap(err, "unable to update object")
			}
		}

		switch *existingCluster.ClusterStatus {
		case "Upgrading":
			s.scope.Info("cluster upgrade in process, requeue")
			return nil
		case "Running":
			s.scope.TKECluster.Status.Ready = true
			s.scope.TKECluster.Status.Initialized = true

			publicEndpointRequest := tke.NewCreateClusterEndpointVipRequest()
			publicEndpointRequest.ClusterId = existingCluster.ClusterId
			publicEndpointRequest.SecurityPolicies = []*string{pointer.StringPtr("0.0.0.0/0")}

			endpointReq := tke.NewDescribeClusterEndpointVipStatusRequest()
			endpointReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

			endpointResp, err := s.tkeClient.DescribeClusterEndpointVipStatus(endpointReq)
			if err != nil {
				return errors.Wrap(err, "unable to describe endpoint")
			}

			if *endpointResp.Response.Status == "NotFound" {
				_, err = s.tkeClient.CreateClusterEndpointVip(publicEndpointRequest)
				if err != nil {
					return errors.Wrap(err, "unable to make endpoint public")
				}
			}

			s.scope.TKECluster.Spec.ControlPlaneEndpoint = v1alpha4.APIEndpoint{
				// https://cls-ekxxyfh6.ccs.tencent-cloud.com
				Host: fmt.Sprintf("https://%s.ccs.tencent-cloud.com", *existingCluster.ClusterId),
				Port: 80,
			}

			err = s.reconcileKubeconfig()
			if err != nil {
				return errors.Wrap(err, "unable to reconcile kubeconfig")
			}

			selectedVersion, err := s.GetCandidateVersion()
			if err != nil {
				return errors.Wrap(err, "unable to get candidate cluster version")
			}

			performMasterUpgrade, err := s.CanPerformUpgrade(selectedVersion)
			if err != nil {
				return errors.Wrap(err, "unable to check cluster upgrade")
			}

			if performMasterUpgrade != nil && *performMasterUpgrade {
				s.scope.Info("performing cluster upgrade")
				err := s.PerformClusterUpgrade(selectedVersion)
				if err != nil {
					return errors.Wrap(err, "unable to perform cluster upgrade")
				}
			}

			performNodePoolUpgrade, err := s.CanPerformNodePoolUpgrade(selectedVersion)
			if err != nil {
				return errors.Wrap(err, "unable to check node upgrade")
			}

			if performNodePoolUpgrade != nil && *performNodePoolUpgrade {
				s.scope.Info("performing nodepool upgrade")
				err := s.PerformNodePoolUpgrade()
				if err != nil {
					return errors.Wrap(err, "unable to perform nodepool upgrade")
				}
			}
		}

		return nil
	}

	selectedVersion, err := s.GetCandidateVersion()
	if err != nil {
		return errors.Wrap(err, "unable to get candidate cluster version")
	}

	req := tke.NewCreateClusterRequest()

	req.ClusterType = pointer.String("MANAGED_CLUSTER")
	req.ClusterCIDRSettings = &tke.ClusterCIDRSettings{
		ClusterCIDR: common.StringPtr("192.168.0.0/16"),
		//IgnoreClusterCIDRConflict: nil,
		//MaxNodePodNum:             nil,
		//MaxClusterServiceNum:      nil,
		//ServiceCIDR:               nil,
		//EniSubnetIds:              nil,
		//ClaimExpiredSeconds:       nil,
	}

	req.ClusterBasicSettings = &tke.ClusterBasicSettings{
		ClusterVersion: pointer.StringPtr(selectedVersion),
		ClusterName:    common.StringPtr(s.scope.TKECluster.Spec.ClusterName),
		//ClusterDescription:    nil,
		VpcId: common.StringPtr(s.scope.TKECluster.Spec.VPCID),
		//ProjectId:             nil,
		TagSpecification: []*tke.TagSpecification{
			{
				ResourceType: common.StringPtr("cluster"),
				Tags: []*tke.Tag{
					{
						Key:   common.StringPtr("cluster-api-provider-tencet/uuid"),
						Value: common.StringPtr(s.scope.TKECluster.Spec.ClusterUUID),
					},
				},
			},
		},
		//OsCustomizeType:       nil,
		//NeedWorkSecurityGroup: nil,
	}

	if s.scope.TKECluster.Spec.ImageID != "" {
		req.ClusterBasicSettings.ClusterOs = pointer.StringPtr(s.scope.TKECluster.Spec.ImageID)
	}

	_, err = s.tkeClient.CreateCluster(req)
	if err != nil {
		return errors.Wrap(err, "unable to create cluster")
	}
	return nil
}

func (s *Service) DeleteCluster(clusterScope *scope.TKEClusterScope) error {
	describerequest := tke.NewDescribeClustersRequest()
	describerequest.Filters = append(describerequest.Filters, &tke.Filter{
		Name: common.StringPtr("tag-key"),
		Values: []*string{
			common.StringPtr("cluster-api-provider-tencet/uuid"),
		},
	}, &tke.Filter{
		Name: common.StringPtr("tag-value"),
		Values: []*string{
			common.StringPtr(s.scope.TKECluster.Spec.ClusterUUID),
		},
	})

	describeResponse, err := s.tkeClient.DescribeClusters(describerequest)
	if err != nil {
		return errors.Wrapf(err, "unable to describe tke cluser using filters: %v", describerequest.Filters)
	}

	if *describeResponse.Response.TotalCount == 0 {
		s.scope.Info("cluster not found, must be deleted already")
		return nil
	}

	deleteRequest := tke.NewDeleteClusterRequest()
	deleteRequest.ClusterId = describeResponse.Response.Clusters[0].ClusterId
	deleteRequest.InstanceDeleteMode = common.StringPtr("terminate")

	_, err = s.tkeClient.DeleteCluster(deleteRequest)
	if err != nil {
		return errors.Wrap(err, "unable to delete cluster")
	}
	return nil
}
