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

package v1alpha4

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var tkeclusterlog = logf.Log.WithName("tkecluster-resource")

func (r *TKECluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1alpha4-tkecluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=tkeclusters,verbs=create;update,versions=v1alpha4,name=mtkecluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &TKECluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *TKECluster) Default() {
	tkeclusterlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.ClusterUUID == "" {
		r.Spec.ClusterUUID = uuid.New().String()
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1alpha4-tkecluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=tkeclusters,verbs=create;update,versions=v1alpha4,name=vtkecluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &TKECluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *TKECluster) ValidateCreate() error {
	tkeclusterlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *TKECluster) ValidateUpdate(old runtime.Object) error {
	tkeclusterlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	oldCluster, ok := old.(*TKECluster)
	if !ok {
		return errors.New(fmt.Sprintf("expected TKECluster object got %T", old))
	}

	if r.Spec.ClusterUUID != oldCluster.Spec.ClusterUUID {
		return errors.New(fmt.Sprintf("TKECluster ClusterUUID field is immutable"))
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *TKECluster) ValidateDelete() error {
	tkeclusterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
