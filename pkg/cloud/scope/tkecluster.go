package scope

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spectrolcoud/cluster-api-provider-tencent/api/v1alpha4"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TKEClusterScopeParams struct {
	Log        logr.Logger
	Client     client.Client
	Cluster    *clusterv1.Cluster
	TKECluster *v1alpha4.TKECluster
}

func NewTKEClusterScope(params TKEClusterScopeParams) (*TKEClusterScope, error) {
	scope := &TKEClusterScope{}

	if params.Cluster == nil {
		return nil, errors.New("failed to generate scope from nil cluster")
	}

	if params.TKECluster == nil {
		return nil, errors.New("failed to generate scope from nil tkecluster")
	}

	scope.Logger = params.Log
	scope.Cluster = params.Cluster
	scope.Client = params.Client
	scope.TKECluster = params.TKECluster

	helper, err := patch.NewHelper(scope.TKECluster, scope.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	scope.patchHelper = helper

	return scope, nil
}

type TKEClusterScope struct {
	logr.Logger
	Cluster     *clusterv1.Cluster
	TKECluster  *v1alpha4.TKECluster
	Client      client.Client
	patchHelper *patch.Helper
}

func (t *TKEClusterScope) Close() error {
	return t.PatchObject()
}

// PatchObject persists the control plane configuration and status.
func (t *TKEClusterScope) PatchObject() error {
	return t.patchHelper.Patch(
		context.TODO(),
		t.TKECluster)
}
