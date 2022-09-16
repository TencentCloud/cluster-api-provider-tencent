package scope

import (
	"context"
	"fmt"

	"k8s.io/utils/pointer"

	"github.com/TencentCloud/cluster-api-provider-tencent/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	//"k8s.io/klog/v2/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	exp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewTKEManagedMachinePool(params TKEManagedMachinePoolScopeParams) (*TKEManagedMachinePoolScope, error) {
	// if params.Log == nil {
	// 	params.Log = klogr.New()
	// }

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
	TKECluster         *v1beta1.TKECluster
	MachinePool        *exp.MachinePool
	ManagedMachinePool *v1beta1.TKEManagedMachinePool
}

type TKEManagedMachinePoolScope struct {
	logr.Logger
	Client             client.Client
	ControllerName     string
	Cluster            *clusterv1.Cluster
	TKECluster         *v1beta1.TKECluster
	MachinePool        *exp.MachinePool
	ManagedMachinePool *v1beta1.TKEManagedMachinePool
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

func (t *TKEManagedMachinePoolScope) WorkerPoolSecurityGroupName() *string {
	return pointer.String(fmt.Sprintf("%s-%s-worker-sg",
		t.TKECluster.Spec.ClusterID,
		t.ManagedMachinePool.Name))
}
