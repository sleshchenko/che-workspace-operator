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
package server

import (
	"context"
	"errors"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	"github.com/che-incubator/che-workspace-operator/internal/controller"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secureServiceName = "che-workspace-controller-secure-service"
)

// InitWebhookServer setups TLS for the webhook server
func InitWebhookServer(ctx context.Context) error {
	crclient, err := controller.CreateClient()
	if err != nil {
		return err
	}

	ns, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}

	err = syncService(ctx, crclient, ns)
	if err != nil {
		return err
	}

	err = syncConfigMap(ctx, crclient, ns)
	if err != nil {
		return err
	}

	err = updateDeployment(ctx, crclient, ns)
	if err != nil {
		return err
	}

	return errors.New("TLS is setup. Controller needs to restart to apply changes")
}

func syncService(ctx context.Context, crclient client.Client, namespace string) error {
	secureService := getSecureServiceSpec(namespace)
	if err := crclient.Create(ctx, secureService); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		existingCfg := &v1.Service{}
		err := crclient.Get(ctx, types.NamespacedName{
			Name:      secureService.Name,
			Namespace: secureService.Namespace,
		}, existingCfg)

		err = crclient.Update(ctx, secureService)
		if err != nil {
			return err
		}
		log.Info("Updated secure service")
	} else {
		log.Info("Created secure service")
	}
	return nil
}

func getSecureServiceSpec(namespace string) *v1.Service {
	label := map[string]string{"app": "che-workspace-controller"}

	port := int32(443)
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-controller",
			Namespace: namespace,
			Labels:    label,
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": "workspace-controller",
			},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:       port,
					Protocol:   "TCP",
					TargetPort: intstr.FromString("webhook-server"),
				},
			},
			Selector: label,
		},
	}

	return service
}

func syncConfigMap(ctx context.Context, crclient client.Client, namespace string) error {
	secureConfigMap := getSecureConfigMapSpec(namespace)
	if err := crclient.Create(ctx, secureConfigMap); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		existingCfg := &v1.Service{}
		err := crclient.Get(ctx, types.NamespacedName{
			Name:      secureConfigMap.Name,
			Namespace: secureConfigMap.Namespace,
		}, existingCfg)

		err = crclient.Update(ctx, secureConfigMap)
		if err != nil {
			return err
		}
		log.Info("Updated secure configmap")
	} else {
		log.Info("Created secure configmap")
	}
	return nil
}

func getSecureConfigMapSpec(namespace string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secureServiceName,
			Namespace: namespace,
			Annotations: map[string]string{
				"service.beta.openshift.io/inject-cabundle": "true",
			},
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
	}
}

// Update the deployment with the volumes needed for webhook server if they aren't already present
func updateDeployment(ctx context.Context, crclient client.Client, namespace string) error {
	deployment, err := controller.FindControllerDeployment(ctx, crclient)
	if err != nil {
		return err
	}

	isVolumeMissing := appendVolumeIfMissing(&deployment.Spec.Template.Spec.Volumes,
		v1.Volume{
			Name: "webhook-tls-certs",
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: []v1.VolumeProjection{
						{
							ConfigMap: &v1.ConfigMapProjection{
								LocalObjectReference: v1.LocalObjectReference{
									Name: secureServiceName,
								},
								Items: []v1.KeyToPath{
									{
										Key:  "service-ca.crt",
										Path: "./ca.crt",
									},
								},
							},
						},
						{
							Secret: &v1.SecretProjection{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "workspace-controller",
								},
							},
						},
					},
				},
			},
		})

	isVMMissing := appendVolumeMountIfMissing(&deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
		*&v1.VolumeMount{
			Name:      "webhook-tls-certs",
			MountPath: webhookServerCertDir,
			ReadOnly:  true,
		})

	// Only bother updating if the volume or volume mount are missing
	if isVolumeMissing || isVMMissing {
		if err = crclient.Update(ctx, deployment); err != nil {
			return err
		}
	}

	return nil
}

// appendVolumeMountIfMissing appends the volume mount if it is missing. Indicates if the volume mount is missing with the return value
func appendVolumeMountIfMissing(volumeMounts *[]v1.VolumeMount, volumeMount v1.VolumeMount) bool {
	for _, vm := range *volumeMounts {
		if vm.Name == volumeMount.Name {
			return false
		}
	}
	*volumeMounts = append(*volumeMounts, volumeMount)
	return true
}

// appendVolumeIfMissing appends the volume if it is missing. Indicates if the volume is missing with the return value
func appendVolumeIfMissing(volumes *[]v1.Volume, volume v1.Volume) bool {
	for _, v := range *volumes {
		if v.Name == volume.Name {
			return true
		}
	}
	*volumes = append(*volumes, volume)
	return true
}
