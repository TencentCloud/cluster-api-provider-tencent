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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1beta1 "github.com/TencentCloud/cluster-api-provider-tencent/api/v1beta1"
)

// TKEClusterReconciler reconciles a TKECluster object
type TKEClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tkeclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tkeclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=tkeclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TKECluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *TKEClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, reterr error) {
	log := logger.FromContext(ctx)

	// TODO(user): your logic here
	tkeCluster := &infrastructurev1beta1.TKECluster{}

	err := r.Get(ctx, req.NamespacedName, tkeCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, tkeCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}

	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	if annotations.IsPaused(cluster, tkeCluster) {
		log.Info("TKECluster or linked Cluster is marked as paused. Won't reconcile")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	clusterScope, err := scope.NewTKEClusterScope(scope.TKEClusterScopeParams{
		Log:        log,
		Client:     r.Client,
		Cluster:    cluster,
		TKECluster: tkeCluster,
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to create scope")
	}

	defer func() {
		if err := clusterScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	if tkeCluster.GetDeletionTimestamp() != nil {
		return reconcileDelete(ctx, clusterScope)
	}

	return reconcileNormal(ctx, clusterScope)
}

func reconcileNormal(ctx context.Context, clusterScope *scope.TKEClusterScope) (ctrl.Result, error) {
	controllerutil.AddFinalizer(clusterScope.TKECluster, infrastructurev1beta1.TKEClusterFinalizer)
	if err := clusterScope.PatchObject(); err != nil {
		return ctrl.Result{}, err
	}

	tkeService, err := tke.NewService(clusterScope)
	if err != nil {
		return ctrl.Result{}, err
	}

	res, err := tkeService.ReconcileCluster(clusterScope)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile tke cluster")
	}

	if res != nil {
		return *res, errors.Wrap(err, "unable to reconcile tke cluster")
	}

	return ctrl.Result{}, nil
}

func reconcileDelete(ctx context.Context, clusterScope *scope.TKEClusterScope) (ctrl.Result, error) {
	tkeService, err := tke.NewService(clusterScope)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = tkeService.DeleteCluster(clusterScope)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to reconcile tke cluster")
	}

	controllerutil.RemoveFinalizer(clusterScope.TKECluster, infrastructurev1beta1.TKEClusterFinalizer)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TKEClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.TKECluster{}).
		Complete(r)
}
