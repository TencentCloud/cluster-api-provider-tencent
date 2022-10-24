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
	infrav1beta1 "github.com/TencentCloud/cluster-api-provider-tencent/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (in *TKECluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1beta1.TKECluster)

	return Convert_v1alpha4_TKECluster_To_v1beta1_TKECluster(in, dst, nil)
}

func (in *TKECluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1beta1.TKECluster)

	return Convert_v1beta1_TKECluster_To_v1alpha4_TKECluster(src, in, nil)
}

func (in *TKEClusterList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1beta1.TKEClusterList)

	return Convert_v1alpha4_TKEClusterList_To_v1beta1_TKEClusterList(in, dst, nil)
}

func (in *TKEClusterList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1beta1.TKEClusterList)

	return Convert_v1beta1_TKEClusterList_To_v1alpha4_TKEClusterList(src, in, nil)
}
