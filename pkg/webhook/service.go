//
// Copyright (c) 2019-2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package webhook

import (
	"context"

	"github.com/devfile/devworkspace-operator/webhook/server"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateSecureService(client client.Client, ctx context.Context, namespace string, annotations map[string]string) error {
	port := int32(443)
	service := &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        server.WebhookServerServiceName,
			Namespace:   namespace,
			Labels:      server.WebhookServerAppLabels(),
			Annotations: annotations,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       port,
					Protocol:   "TCP",
					TargetPort: intstr.FromString(server.WebhookServerPortName),
				},
			},
			Selector: server.WebhookServerAppLabels(),
		},
	}

	if err := client.Create(ctx, service); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		existingCfg, err := getClusterService(ctx, namespace, client)
		if err != nil {
			return err
		}

		// Cannot naively copy spec, as clusterIP is unmodifiable
		clusterIP := existingCfg.Spec.ClusterIP
		service.Spec = existingCfg.Spec
		service.Spec.ClusterIP = clusterIP
		service.ResourceVersion = existingCfg.ResourceVersion

		err = client.Update(ctx, service)
		if err != nil {
			return err
		}
		log.Info("Updated webhook server service")
	} else {
		log.Info("Created webhook server service")
	}
	return nil
}

func getClusterService(ctx context.Context, namespace string, client client.Client) (*v1.Service, error) {
	service := &v1.Service{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      server.WebhookServerServiceName,
	}
	err := client.Get(ctx, namespacedName, service)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return service, nil
}
