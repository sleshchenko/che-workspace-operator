package server

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	secureServiceName = "workspace-controller"
	certConfigMapName = "che-workspace-controller-secure-service"
	certSecretName    = "workspace-controller"
	certVolumeName    = "webhook-tls-certs"
)

func getSecureServiceSpec(namespace string) *corev1.Service {
	label := map[string]string{"app": "che-workspace-controller"}

	port := int32(443)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secureServiceName,
			Namespace: namespace,
			Labels:    label,
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": certSecretName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
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

func getSecureConfigMapSpec(namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certConfigMapName,
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

func getCertVolume() corev1.Volume {
	return corev1.Volume{
		Name: certVolumeName,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: certConfigMapName,
							},
							Items: []corev1.KeyToPath{
								{
									Key:  "service-ca.crt",
									Path: "./ca.crt",
								},
							},
						},
					},
					{
						Secret: &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: certSecretName,
							},
						},
					},
				},
			},
		},
	}
}

func getCertVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      certVolumeName,
		MountPath: webhookServerCertDir,
		ReadOnly:  true,
	}
}
