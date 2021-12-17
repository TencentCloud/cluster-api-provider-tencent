package tke

import (
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	"k8s.io/utils/pointer"
)

type KubernetesVersion interface {
	GetCandidateVersion()
	CanPerformUpgrade()
	PerformClusterUpgrade()
}

func (s *Service) GetCandidateVersion() (string, error) {
	availableVersionsReq := tke.NewDescribeVersionsRequest()

	availableVersions, err := s.tkeClient.DescribeVersions(availableVersionsReq)
	if err != nil {
		return "", errors.Wrap(err, "unable to get available kubernetes version")
	}

	selectedVersion := ""
	if s.scope.TKECluster.Spec.ClusterVersion == nil {
		for _, version := range availableVersions.Response.VersionInstanceSet {
			currentVersion := semver.MustParse(*version.Version)
			if selectedVersion == "" {
				selectedVersion = currentVersion.String()
			} else {
				selectedVersionParsed := semver.MustParse(selectedVersion)
				if currentVersion.GTE(selectedVersionParsed) {
					selectedVersion = currentVersion.String()
				}
			}
		}
	} else {
		s.scope.Info("reuqested kuberntes version", "selectedVersion", *s.scope.TKECluster.Spec.ClusterVersion)
		for _, version := range availableVersions.Response.VersionInstanceSet {
			requestedVersion := semver.MustParse(*s.scope.TKECluster.Spec.ClusterVersion)
			currentVersion := semver.MustParse(*version.Version)

			s.scope.V(2).Info("processing kuberntes version", "selectedVersion", *s.scope.TKECluster.Spec.ClusterVersion)
			if currentVersion.GTE(requestedVersion) {
				if selectedVersion != "" {
					selectedVersionParsed := semver.MustParse(selectedVersion)
					if currentVersion.LTE(selectedVersionParsed) {
						selectedVersion = currentVersion.String()
					}
				} else {
					selectedVersion = currentVersion.String()
				}
			}
		}
	}

	s.scope.Info("selected kubernetes version for cluster", "selectedVersion", selectedVersion)

	if selectedVersion == "" {
		s.scope.Info("could not select candidate version", "selectedVersion", selectedVersion)
		return "", errors.New("failed select kubernetes version")
	}

	return selectedVersion, nil
}

func (s *Service) CanPerformUpgrade(requestedVersion string) (*bool, error) {
	req := tke.NewDescribeAvailableClusterVersionRequest()
	req.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	res, err := s.tkeClient.DescribeAvailableClusterVersion(req)
	if err != nil {
		return nil, err
	}

	for _, cluster := range res.Response.Clusters {
		if *cluster.ClusterId == s.scope.TKECluster.Spec.ClusterID {
			for _, version := range cluster.Versions {
				if *version == requestedVersion {
					return pointer.BoolPtr(true), nil
				}
			}
		}
	}

	return pointer.BoolPtr(false), nil
}

func (s *Service) CanPerformNodePoolUpgrade(requestedVersion string) (*bool, error) {
	req := tke.NewCheckInstancesUpgradeAbleRequest()
	req.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	res, err := s.tkeClient.CheckInstancesUpgradeAble(req)
	if err != nil {
		return nil, err
	}

	if len(res.Response.UpgradeAbleInstances) > 0 {
		return pointer.BoolPtr(true), nil
	}

	return pointer.BoolPtr(false), nil
}

func (s *Service) PerformClusterUpgrade(version string) error {
	req := tke.NewUpdateClusterVersionRequest()
	req.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
	req.DstVersion = pointer.StringPtr(version)

	_, err := s.tkeClient.UpdateClusterVersion(req)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) PerformNodePoolUpgrade() error {
	req := tke.NewCheckInstancesUpgradeAbleRequest()
	req.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	res, err := s.tkeClient.CheckInstancesUpgradeAble(req)
	if err != nil {
		return err
	}

	for _, instance := range res.Response.UpgradeAbleInstances {
		s.scope.Info("upgrading node pool instance", "name", *instance.InstanceId)
		tke.NewCheckInstancesUpgradeAbleRequest()
		upgradeReq := tke.NewUpgradeClusterInstancesRequest()
		upgradeReq.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)
		upgradeReq.InstanceIds = []*string{
			instance.InstanceId,
		}
		upgradeReq.ResetParam = &tke.UpgradeNodeResetParam{}
		upgradeReq.Operation = common.StringPtr("create")
		upgradeReq.UpgradeType = common.StringPtr("reset")

		_, err := s.tkeClient.UpgradeClusterInstances(upgradeReq)
		if err != nil {
			return errors.Wrap(err, "unable to upgrade node pool")
		}
	}

	return nil
}
