module github.com/TencentCloud/cluster-api-provider-tencent

go 1.16

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-logr/logr v1.2.0
	github.com/google/uuid v1.1.2
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as v1.0.498
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.498
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm v1.0.498
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke v1.0.498
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc v1.0.498
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/cluster-api v1.1.3
	sigs.k8s.io/controller-runtime v0.11.1
)
