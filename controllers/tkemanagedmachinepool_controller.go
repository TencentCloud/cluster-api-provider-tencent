/*
Copyright 2021.

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

package controllers

import (
	"context"

	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/cloud/scope"
	"github.com/TencentCloud/cluster-api-provider-tencent/pkg/cloud/services/tke"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/cluster-api/exp/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1alpha4 "github.com/TencentCloud/cluster-api-provider-tencent/api/v1alpha4"
)

// TKEManagedMachinePoolReconciler reconciles a TKEManagedMachinePool object
type TKEManagedMachinePoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tkemanagedmachinepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tkemanagedmachinepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tkemanagedmachinepools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TKEManagedMachinePool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TKEManagedMachinePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, reterr error) {
	log := logger.FromContext(ctx)

	// TODO(user): your logic here
	tkePool := &infrastructurev1alpha4.TKEManagedMachinePool{}
	if err := r.Get(ctx, req.NamespacedName, tkePool); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}

	machinePool, err := getOwnerMachinePool(ctx, r.Client, tkePool.ObjectMeta)
	if err != nil {
		log.Error(err, "Failed to retrieve owner MachinePool from the API Server")
		return ctrl.Result{}, err
	}
	if machinePool == nil {
		log.Info("MachinePool Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	log = log.WithValues("MachinePool", machinePool.Name)

	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machinePool.ObjectMeta)
	if err != nil {
		log.Info("Failed to retrieve Cluster from MachinePool")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("Cluster", cluster.Name)

	controlPlaneKey := client.ObjectKey{
		Namespace: tkePool.Namespace,
		Name:      cluster.Spec.ControlPlaneRef.Name,
	}
	tkeCluster := &infrastructurev1alpha4.TKECluster{}
	if err := r.Client.Get(ctx, controlPlaneKey, tkeCluster); err != nil {
		log.Info("Failed to retrieve ControlPlane from MachinePool")
		return reconcile.Result{}, nil
	}

	if !tkeCluster.Status.Ready {
		log.Info("cluster is not ready yet")
		conditions.MarkFalse(tkePool, infrastructurev1alpha4.TKENodepoolReadyCondition, infrastructurev1alpha4.WaitingForTKEClusterReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	machinePoolScope, err := scope.NewTKEManagedMachinePool(scope.TKEManagedMachinePoolScopeParams{
		Log:                log,
		Client:             r.Client,
		ControllerName:     "tkemanagedmachinepool",
		Cluster:            cluster,
		TKECluster:         tkeCluster,
		MachinePool:        machinePool,
		ManagedMachinePool: tkePool,
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to create scope")
	}

	defer func() {
		applicableConditions := []clusterv1.ConditionType{}

		conditions.SetSummary(machinePoolScope.ManagedMachinePool, conditions.WithConditions(applicableConditions...), conditions.WithStepCounter())

		if err := machinePoolScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	if !tkePool.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, machinePoolScope)
	}

	return r.reconcileNormal(ctx, machinePoolScope)
}

// getOwnerMachinePool returns the MachinePool object owning the current resource.
func getOwnerMachinePool(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*v1alpha4.MachinePool, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "MachinePool" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if gv.Group == v1alpha4.GroupVersion.Group {
			return getMachinePoolByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// getMachinePoolByName finds and return a Machine object using the specified params.
func getMachinePoolByName(ctx context.Context, c client.Client, namespace, name string) (*v1alpha4.MachinePool, error) {
	m := &v1alpha4.MachinePool{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(ctx, key, m); err != nil {
		return nil, err
	}
	return m, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TKEManagedMachinePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha4.TKEManagedMachinePool{}).
		Complete(r)
}

func (r *TKEManagedMachinePoolReconciler) reconcileNormal(ctx context.Context, scope *scope.TKEManagedMachinePoolScope) (ctrl.Result, error) {
	scope.Info("Reconciling TKEManagedMachinePool")

	controllerutil.AddFinalizer(scope.ManagedMachinePool, infrastructurev1alpha4.TKEManagedMachinePoolFinalizer)
	if err := scope.PatchObject(); err != nil {
		return ctrl.Result{}, err
	}

	tkeService, err := tke.NewNodePoolService(scope)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to create tke service")
	}

	err = tkeService.ReconcileNodePool()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile node pool")
	}

	return ctrl.Result{}, nil
}

func (r *TKEManagedMachinePoolReconciler) reconcileDelete(ctx context.Context, scope *scope.TKEManagedMachinePoolScope) (ctrl.Result, error) {
	scope.Info("Reconciling TKEManagedMachinePool deletion")

	tkeService, err := tke.NewNodePoolService(scope)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to create tke service")
	}

	err = tkeService.DeleteNodePool()
	if err != nil {
		return ctrl.Result{}, err
	}

	controllerutil.RemoveFinalizer(scope.ManagedMachinePool, infrastructurev1alpha4.TKEManagedMachinePoolFinalizer)

	return ctrl.Result{}, nil
}
