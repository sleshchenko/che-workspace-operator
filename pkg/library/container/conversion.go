//
// Copyright (c) 2019-2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package container

import (
	"fmt"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dwEnv "github.com/devfile/devworkspace-operator/controllers/workspace/env"

	"github.com/devfile/devworkspace-operator/pkg/config"
	"github.com/devfile/devworkspace-operator/pkg/constants"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func convertContainerToK8s(devfileComponent dw.Component) (*v1.Container, error) {
	if devfileComponent.Container == nil {
		return nil, fmt.Errorf("cannot get k8s container from non-container component")
	}
	devfileContainer := devfileComponent.Container

	containerResources, err := devfileResourcesToContainerResources(devfileContainer)
	if err != nil {
		return nil, err
	}

	container := &v1.Container{
		Name:            devfileComponent.Name,
		Image:           devfileContainer.Image,
		Command:         devfileContainer.Command,
		Args:            devfileContainer.Args,
		Resources:       *containerResources,
		Ports:           devfileEndpointsToContainerPorts(devfileContainer.Endpoints),
		Env:             devfileEnvToContainerEnv(devfileComponent.Name, devfileContainer.Env),
		VolumeMounts:    devfileVolumeMountsToContainerVolumeMounts(devfileContainer.VolumeMounts),
		ImagePullPolicy: v1.PullPolicy(config.ControllerCfg.GetSidecarPullPolicy()),
	}

	return container, nil
}

func devfileEndpointsToContainerPorts(endpoints []dw.Endpoint) []v1.ContainerPort {
	var containerPorts []v1.ContainerPort
	exposedPorts := map[int]bool{}
	for _, endpoint := range endpoints {
		if exposedPorts[endpoint.TargetPort] {
			continue
		}
		containerPorts = append(containerPorts, v1.ContainerPort{
			// Use meaningless name for port since endpoint.Name does not match requirements for ContainerPort name
			Name:          fmt.Sprintf("%d-%s", endpoint.TargetPort, endpoint.Protocol),
			ContainerPort: int32(endpoint.TargetPort),
			Protocol:      v1.ProtocolTCP,
		})
		exposedPorts[endpoint.TargetPort] = true
	}
	return containerPorts
}

func devfileResourcesToContainerResources(devfileContainer *dw.ContainerComponent) (*v1.ResourceRequirements, error) {
	// TODO: Handle memory request and CPU when implemented in devfile API
	memLimit := devfileContainer.MemoryLimit
	if memLimit == "" {
		memLimit = constants.SidecarDefaultMemoryLimit
	}
	memLimitQuantity, err := resource.ParseQuantity(memLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory limit %q: %w", memLimit, err)
	}

	memReq := devfileContainer.MemoryRequest
	if memReq == "" {
		memReq = "512Mi"
	}
	memReqQuantity, err := resource.ParseQuantity(memLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory limit %q: %w", memLimit, err)
	}

	cpuLimit := devfileContainer.CpuLimit
	if cpuLimit == "" {
		cpuLimit = "50m"
	}
	cpuLimitQuantity, err := resource.ParseQuantity(cpuLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu limit %q: %w", cpuLimit, err)
	}

	cpuReq := devfileContainer.CpuRequest
	if cpuReq == "" {
		cpuReq = "512Mi"
	}
	cpuReqQuantity, err := resource.ParseQuantity(cpuLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu limit %q: %w", cpuLimit, err)
	}

	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: memLimitQuantity,
			v1.ResourceCPU:    cpuLimitQuantity,
		},
		Requests: v1.ResourceList{
			v1.ResourceMemory: memReqQuantity,
			v1.ResourceCPU:    cpuReqQuantity,
		},
	}, nil
}

func devfileVolumeMountsToContainerVolumeMounts(devfileVolumeMounts []dw.VolumeMount) []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount
	for _, vm := range devfileVolumeMounts {
		path := vm.Path
		if path == "" {
			// Devfile API spec: if path is unspecified, default is to use volume name
			path = fmt.Sprintf("/%s", vm.Name)
		}
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      vm.Name,
			MountPath: path,
		})
	}
	return volumeMounts
}

func devfileEnvToContainerEnv(componentName string, devfileEnvVars []dw.EnvVar) []v1.EnvVar {
	var env = []v1.EnvVar{
		{
			Name:  dwEnv.DevWorkspaceComponentName,
			Value: componentName,
		},
	}

	for _, devfileEnv := range devfileEnvVars {
		env = append(env, v1.EnvVar{
			Name:  devfileEnv.Name,
			Value: devfileEnv.Value,
		})
	}
	return env
}
