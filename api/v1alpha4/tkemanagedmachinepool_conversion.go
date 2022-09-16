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

func (in *TKEManagedMachinePool) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1beta1.TKEManagedMachinePool)

	return Convert_v1alpha4_TKEManagedMachinePool_To_v1beta1_TKEManagedMachinePool(in, dst, nil)
}

func (in *TKEManagedMachinePool) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1beta1.TKEManagedMachinePool)

	return Convert_v1beta1_TKEManagedMachinePool_To_v1alpha4_TKEManagedMachinePool(src, in, nil)
}

func (in *TKEManagedMachinePoolList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*infrav1beta1.TKEManagedMachinePoolList)

	return Convert_v1alpha4_TKEManagedMachinePoolList_To_v1beta1_TKEManagedMachinePoolList(in, dst, nil)
}

func (in *TKEManagedMachinePoolList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*infrav1beta1.TKEManagedMachinePoolList)

	return Convert_v1beta1_TKEManagedMachinePoolList_To_v1alpha4_TKEManagedMachinePoolList(src, in, nil)
}
