/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2024 The KubeVirt Authors.
 *
 */

package hotplug

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/converter"
)

func ValidateMemoryHotplug(vmSpec *v1.VirtualMachineInstanceSpec) []metav1.StatusCause {
	domain := &vmSpec.Domain

	causes := []metav1.StatusCause{}
	field := field.NewPath("spec", "template", "spec")

	if domain.Resources.Limits.Memory() != nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Configuration of Memory limits is not allowed when Memory live update is enabled"),
			Field:   field.Child("domain", "resources").String(),
		})
	}

	if domain.CPU != nil &&
		domain.CPU.Realtime != nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Memory hotplug is not compatible with realtime VMs"),
			Field:   field.Child("domain", "cpu", "realtime").String(),
		})
	}

	if domain.CPU != nil &&
		domain.CPU.NUMA != nil &&
		domain.CPU.NUMA.GuestMappingPassthrough != nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Memory hotplug is not compatible with guest mapping passthrough"),
			Field:   field.Child("domain", "cpu", "numa", "guestMappingPassthrough").String(),
		})
	}

	if domain.LaunchSecurity != nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Memory hotplug is not compatible with encrypted VMs"),
			Field:   field.Child("domain", "launchSecurity").String(),
		})
	}

	if domain.CPU != nil &&
		domain.CPU.DedicatedCPUPlacement {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Memory hotplug is not compatible with dedicated CPUs"),
			Field:   field.Child("domain", "cpu", "dedicatedCpuPlacement").String(),
		})
	}

	if domain.Memory != nil &&
		domain.Memory.Hugepages != nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Memory hotplug is not compatible with hugepages"),
			Field:   field.Child("domain", "memory", "hugepages").String(),
		})
	}

	if domain.Memory == nil ||
		domain.Memory.Guest == nil {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Guest memory must be configured when memory hotplug is enabled"),
			Field:   field.Child("domain", "memory", "guest").String(),
		})
	} else if domain.Memory.MaxGuest != nil &&
		domain.Memory.Guest.Cmp(*domain.Memory.MaxGuest) > 0 {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Guest memory is greater than the configured maxGuest memory"),
			Field:   field.Child("domain", "memory", "guest").String(),
		})
	} else if domain.Memory.Guest.Value()%converter.MemoryHotplugBlockAlignmentBytes != 0 {
		alignment := resource.NewQuantity(converter.MemoryHotplugBlockAlignmentBytes, resource.BinarySI)
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Guest memory must be %s aligned", alignment),
			Field:   field.Child("domain", "memory", "guest").String(),
		})
	}

	if domain.Memory.MaxGuest != nil &&
		domain.Memory.MaxGuest.Value()%converter.MemoryHotplugBlockAlignmentBytes != 0 {
		alignment := resource.NewQuantity(converter.MemoryHotplugBlockAlignmentBytes, resource.BinarySI)
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("MaxGuest must be %s aligned", alignment),
			Field:   field.Child("domain", "memory", "maxGuest").String(),
		})
	}

	if vmSpec.Architecture != "amd64" {
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseTypeFieldValueInvalid,
			Message: fmt.Sprintf("Memory hotplug is only available for x86_64 VMs"),
			Field:   field.Child("architecture").String(),
		})
	}

	return causes
}
