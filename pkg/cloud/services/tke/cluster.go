package tke

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/TencentCloud/cluster-api-provider-tencent/api/v1beta1"
	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/cloud/scope"
)

const (
	ClusterUUIDTag = "cluster-api-provider-tencent/uuid"
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

func (s *Service) ReconcileCluster(clusterScope *scope.TKEClusterScope) (*ctrl.Result, error) {
	describerequest := tke.NewDescribeClustersRequest()
	describerequest.Filters = append(describerequest.Filters, &tke.Filter{
		Name: common.StringPtr("tag-key"),
		Values: []*string{
			common.StringPtr(ClusterUUIDTag),
		},
	}, &tke.Filter{
		Name: common.StringPtr("tag-value"),
		Values: []*string{
			common.StringPtr(s.scope.TKECluster.Spec.ClusterUUID),
		},
	})

	describeResponse, err := s.tkeClient.DescribeClusters(describerequest)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to describe tke cluser using filters: %v", describerequest.Filters)
	}

	if *describeResponse.Response.TotalCount > 0 {
		existingCluster := describeResponse.Response.Clusters[0]
		s.scope.Info("cluster already creating skip", "status", existingCluster.ClusterStatus)

		if s.scope.TKECluster.Spec.ClusterID == "" {
			s.scope.TKECluster.Spec.ClusterID = *existingCluster.ClusterId

			if err := s.scope.PatchObject(); err != nil {
				return nil, errors.Wrap(err, "unable to update object")
			}
		}

		switch *existingCluster.ClusterStatus {
		case "Upgrading":
			s.scope.Info("cluster upgrade in process, requeue")
			s.scope.TKECluster.Status.Ready = false
			return nil, nil
		case "Running":
			s.scope.TKECluster.Status.Ready = true
			s.scope.TKECluster.Status.Initialized = true

			describeEndpoints := tke.NewDescribeClusterEndpointsRequest()
			describeEndpoints.ClusterId = existingCluster.ClusterId

			describeEndpointsResult, err := s.tkeClient.DescribeClusterEndpoints(describeEndpoints)
			if err != nil {
				return nil, err
			}

			tmmlist := &v1beta1.TKEManagedMachinePoolList{}
			err = s.scope.Client.List(context.TODO(), tmmlist, client.InNamespace(s.scope.TKECluster.Namespace))
			if err != nil {
				return nil, err
			}

			onePoolReady := false

			for _, tkeManagedMachinePool := range tmmlist.Items {
				if tkeManagedMachinePool.Status.Ready {
					onePoolReady = true
					break
				}
			}

			if !onePoolReady {
				s.scope.V(0).Info("pool not ready wait")
				return &ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
			}

			if s.isPublicEndPointEnabled() && !s.isPublicEndPointAvailable(*describeEndpointsResult.Response.ClusterExternalEndpoint) {
				if s.scope.TKECluster.Spec.EndpointAccess.SecurityGroup == nil || *s.scope.TKECluster.Spec.EndpointAccess.SecurityGroup == "" {
					return nil, errors.New("missing security group in endpoint access")
				}

				endPoint := tke.NewCreateClusterEndpointRequest()
				endPoint.ClusterId = pointer.String(s.scope.TKECluster.Spec.ClusterID)
				endPoint.SecurityGroup = s.scope.TKECluster.Spec.EndpointAccess.SecurityGroup
				endPoint.IsExtranet = pointer.Bool(true)

				_, err := s.tkeClient.CreateClusterEndpoint(endPoint)
				if err != nil {
					return nil, err
				}

				return &ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
			}

			if s.isPrivateEndPointEnabled() && !s.isPrivateEndPointAvailable(*describeEndpointsResult.Response.ClusterIntranetEndpoint) {
				if s.scope.TKECluster.Spec.EndpointAccess.SubnetID == nil || *s.scope.TKECluster.Spec.EndpointAccess.SubnetID == "" {
					return nil, errors.New("missing subnetID in endpoint access")
				}

				endPoint := tke.NewCreateClusterEndpointRequest()
				endPoint.ClusterId = pointer.String(s.scope.TKECluster.Spec.ClusterID)
				endPoint.SubnetId = s.scope.TKECluster.Spec.EndpointAccess.SubnetID

				_, err := s.tkeClient.CreateClusterEndpoint(endPoint)
				if err != nil {
					return nil, err
				}

				return &ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
			}

			if s.isPrivateEndPointAvailable(*describeEndpointsResult.Response.ClusterIntranetEndpoint) {
				s.scope.TKECluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
					// https://10.0.0.6
					Host: fmt.Sprintf("https://%s", *describeEndpointsResult.Response.ClusterIntranetEndpoint),
					Port: 443,
				}
			}

			if s.isPublicEndPointAvailable(*describeEndpointsResult.Response.ClusterIntranetEndpoint) {
				s.scope.TKECluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
					//	// https://cls-ekxxyfh6.ccs.tencent-cloud.com
					Host: fmt.Sprintf("https://%s.ccs.tencent-cloud.com", *existingCluster.ClusterId),
					Port: 443,
				}
			}

			err = s.reconcileKubeconfig()
			if err != nil {
				return nil, errors.Wrap(err, "unable to reconcile kubeconfig")
			}

			selectedVersion, err := s.GetCandidateVersion()
			if err != nil {
				return nil, errors.Wrap(err, "unable to get candidate cluster version")
			}

			performMasterUpgrade, err := s.CanPerformUpgrade(selectedVersion)
			if err != nil {
				return nil, errors.Wrap(err, "unable to check cluster upgrade")
			}

			if performMasterUpgrade != nil && *performMasterUpgrade {
				s.scope.Info("performing cluster upgrade")
				err := s.PerformClusterUpgrade(selectedVersion)
				if err != nil {
					return nil, errors.Wrap(err, "unable to perform cluster upgrade")
				}
			}

			performNodePoolUpgrade, err := s.CanPerformNodePoolUpgrade(selectedVersion)
			if err != nil {
				return nil, errors.Wrap(err, "unable to check node upgrade")
			}

			if performNodePoolUpgrade != nil && *performNodePoolUpgrade {
				s.scope.Info("performing nodepool upgrade")
				err := s.PerformNodePoolUpgrade()
				if err != nil {
					return nil, errors.Wrap(err, "unable to perform nodepool upgrade")
				}
			}
		}

		return nil, nil
	}

	selectedVersion, err := s.GetCandidateVersion()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get candidate cluster version")
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
						Key:   common.StringPtr(ClusterUUIDTag),
						Value: common.StringPtr(s.scope.TKECluster.Spec.ClusterUUID),
					},
				},
			},
		},
		//OsCustomizeType:       nil,
		//NeedWorkSecurityGroup: nil,
	}

	if s.scope.TKECluster.Spec.OSName != "" {
		req.ClusterBasicSettings.ClusterOs = pointer.StringPtr(s.scope.TKECluster.Spec.OSName)
	}

	_, err = s.tkeClient.CreateCluster(req)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create cluster")
	}
	return nil, nil
}

func (s *Service) isPublicEndPointEnabled() bool {
	return s.scope.TKECluster.Spec.EndpointAccess.Public != nil && *s.scope.TKECluster.Spec.EndpointAccess.Public
}

func (s *Service) isPublicEndPointAvailable(clusterExternalEndpoint string) bool {
	return clusterExternalEndpoint != ""
}

func (s *Service) isPrivateEndPointEnabled() bool {
	return s.scope.TKECluster.Spec.EndpointAccess.Public != nil && *s.scope.TKECluster.Spec.EndpointAccess.Public
}

func (s *Service) isPrivateEndPointAvailable(clusterIntranetEndpoint string) bool {
	return clusterIntranetEndpoint != ""
}

func (s *Service) DeleteCluster(clusterScope *scope.TKEClusterScope) error {
	describerequest := tke.NewDescribeClustersRequest()
	describerequest.Filters = append(describerequest.Filters, &tke.Filter{
		Name: common.StringPtr("tag-key"),
		Values: []*string{
			common.StringPtr(ClusterUUIDTag),
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

func (s *Service) isDeleteRequired(status string) bool {
	if status != "Deleted" && status != "Deleting" && status != "Creating" && status != "NotFound" {
		return true
	}

	return false
}
