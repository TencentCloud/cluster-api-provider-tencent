package v1beta1

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

const (
	// TKENodepoolReadyCondition condition reports on the successful reconciliation of tke node pool.
	TKENodepoolReadyCondition clusterv1.ConditionType = "TKENodepoolReady"
	// WaitingForTKEClusterReason used when the machine pool is waiting for
	// tke cluster infrastructure to be ready before proceeding.
	WaitingForTKEClusterReason = "WaitingForTKECluster"
)

const (
	// ConditionSeverityError specifies that a condition with `Status=False` is an error.
	ConditionSeverityError clusterv1.ConditionSeverity = "Error"

	// ConditionSeverityWarning specifies that a condition with `Status=False` is a warning.
	ConditionSeverityWarning clusterv1.ConditionSeverity = "Warning"

	// ConditionSeverityInfo specifies that a condition with `Status=False` is informative.
	ConditionSeverityInfo clusterv1.ConditionSeverity = "Info"

	// ConditionSeverityNone should apply only to conditions with `Status=True`.
	ConditionSeverityNone clusterv1.ConditionSeverity = ""
)
