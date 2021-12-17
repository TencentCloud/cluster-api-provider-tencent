package scope

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/spectrolcoud/cluster-api-provider-tencent/api/v1alpha4"
	"k8s.io/klog/v2/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	exp "sigs.k8s.io/cluster-api/exp/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTKEManagedMachinePool(params TKEManagedMachinePoolScopeParams) (*TKEManagedMachinePoolScope, error) {
	if params.Log == nil {
		params.Log = klogr.New()
	}

	scope := &TKEManagedMachinePoolScope{
		Logger:             params.Log,
		Client:             params.Client,
		ControllerName:     params.ControllerName,
		Cluster:            params.Cluster,
		TKECluster:         params.TKECluster,
		MachinePool:        params.MachinePool,
		ManagedMachinePool: params.ManagedMachinePool,
	}

	helper, err := patch.NewHelper(scope.ManagedMachinePool, scope.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	scope.patchHelper = helper

	return scope, nil
}

type TKEManagedMachinePoolScopeParams struct {
	Log                logr.Logger
	Client             client.Client
	ControllerName     string
	Cluster            *clusterv1.Cluster
	TKECluster         *v1alpha4.TKECluster
	MachinePool        *exp.MachinePool
	ManagedMachinePool *v1alpha4.TKEManagedMachinePool
}

type TKEManagedMachinePoolScope struct {
	logr.Logger
	Client             client.Client
	ControllerName     string
	Cluster            *clusterv1.Cluster
	TKECluster         *v1alpha4.TKECluster
	MachinePool        *exp.MachinePool
	ManagedMachinePool *v1alpha4.TKEManagedMachinePool
	patchHelper        *patch.Helper
}

func (t *TKEManagedMachinePoolScope) Close() error {
	return t.PatchObject()
}

func (t *TKEManagedMachinePoolScope) PatchObject() error {
	return t.patchHelper.Patch(context.TODO(), t.ManagedMachinePool)
}

func (t *TKEManagedMachinePoolScope) SetReady() {
	t.ManagedMachinePool.Status.Ready = true
}

func (t *TKEManagedMachinePoolScope) SetNotReady() {
	t.ManagedMachinePool.Status.Ready = false
}
