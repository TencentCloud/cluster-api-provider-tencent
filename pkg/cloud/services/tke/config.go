package tke

import (
	"bytes"
	"context"
	"fmt"

	infrastructurev1beta1 "github.com/TencentCloud/cluster-api-provider-tencent/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *Service) reconcileKubeconfig() error {
	kubeconfigRequest := tke.NewDescribeClusterKubeconfigRequest()
	kubeconfigRequest.IsExtranet = pointer.BoolPtr(true)
	kubeconfigRequest.ClusterId = pointer.StringPtr(s.scope.TKECluster.Spec.ClusterID)

	kubeconfigResponse, err := s.tkeClient.DescribeClusterKubeconfig(kubeconfigRequest)
	if err != nil {
		return errors.Wrap(err, "unable to get kubeconfig")
	}

	m := manager{
		Logger:     s.scope.Logger,
		client:     s.scope.Client,
		cluster:    s.scope.Cluster,
		tkeCluster: s.scope.TKECluster,
		kubeconfig: []byte(*kubeconfigResponse.Response.Kubeconfig),
	}

	secretExists, err := m.SecretExists()
	if err != nil {
		return err
	}

	if secretExists != nil && *secretExists {
		return m.UpdateSecret()
	}

	return m.CreateSecret()
}

type KubeconfigSecret interface {
	SecretExists() (*bool, error)
	UpdateSecret() error
	CreateSecret() error
}

type manager struct {
	logr.Logger
	client     client.Client
	cluster    *v1beta1.Cluster
	tkeCluster *infrastructurev1beta1.TKECluster
	secret     *v1.Secret
	kubeconfig []byte
}

func (m *manager) SecretExists() (*bool, error) {
	s := &v1.Secret{}
	err := m.client.Get(context.Background(), types.NamespacedName{
		Namespace: m.cluster.Namespace,
		Name:      m.secretName(m.cluster.Name),
	}, s)
	switch {
	case err != nil && apierrors.IsNotFound(err):
		return pointer.BoolPtr(false), nil
	case err != nil:
		return nil, err
	}

	m.secret = s
	return pointer.BoolPtr(true), nil
}

func (m *manager) UpdateSecret() error {
	m.V(2).Info("Updating TKE kubeconfigs for cluster", "cluster-name", m.cluster.Name)

	data, ok := m.secret.Data[secret.KubeconfigDataName]
	if !ok {
		return errors.Errorf("missing key %q in secret data", secret.KubeconfigDataName)
	}

	if bytes.Equal(data, m.kubeconfig) {
		return nil
	}

	m.secret.Data[secret.KubeconfigDataName] = m.kubeconfig

	err := m.client.Update(context.Background(), m.secret)
	if err != nil {
		return errors.Wrap(err, "unable to update kubeconfig secret")
	}
	return nil
}

func (m *manager) CreateSecret() error {
	controllerOwnerRef := *metav1.NewControllerRef(m.tkeCluster, infrastructurev1beta1.GroupVersion.WithKind("TKECluster"))

	s := kubeconfig.GenerateSecretWithOwner(types.NamespacedName{
		Namespace: m.cluster.Namespace,
		Name:      m.cluster.Name,
	}, m.kubeconfig, controllerOwnerRef)

	err := m.client.Create(context.Background(), s)
	if err != nil {
		return errors.Wrap(err, "unable to create kubeconfig secret")
	}
	return nil
}

func (m *manager) secretName(clusterName string) string {
	return fmt.Sprintf("%s-kubeconfig", clusterName)
}
