module github.com/spectrolcoud/cluster-api-provider-tencent

go 1.16

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-logr/logr v0.4.0
	github.com/google/uuid v1.1.2
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as v1.0.298
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.289
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm v1.0.299
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke v1.0.289
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc v1.0.292
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	sigs.k8s.io/cluster-api v0.4.3
	sigs.k8s.io/controller-runtime v0.10.2
)

replace github.com/spectrolcoud/cluster-api-provider-tencent/pkg/cloud/scope => ./pkg/cloud/scope
