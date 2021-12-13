package v1alpha4

import clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"

const (
	// TKENodepoolReadyCondition condition reports on the successful reconciliation of tke node pool.
	TKENodepoolReadyCondition clusterv1.ConditionType = "TKENodepoolReady"
	// WaitingForTKEClusterReason used when the machine pool is waiting for
	// tke cluster infrastructure to be ready before proceeding.
	WaitingForTKEClusterReason = "WaitingForTKECluster"
)
