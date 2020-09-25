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

package webhook_k8s

import (
	"context"
	"github.com/devfile/devworkspace-operator/internal/images"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func SyncJobToCluster(
	client crclient.Client,
	ctx context.Context,
	serviceAccountName string,
	name string,
	namespace string,
	env map[string]string,
) error {

	specJob, err := getSpecJob(serviceAccountName, name, namespace, env)
	if err != nil {
		return err
	}

	if err := client.Create(ctx, specJob); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		existingCfg, err := getJobInNamespace(client, TLSJobName, specJob.Namespace)
		if err != nil {
			return err
		}
		specJob.ResourceVersion = existingCfg.ResourceVersion
		err = client.Update(ctx, specJob)
		if err != nil {
			return err
		}
		log.Info("Updated Job")
	} else {
		log.Info("Created Job")
	}

	return nil
}

// GetSpecJob creates new job configuration by given parameters.
func getSpecJob(
	serviceAccountName string,
	name string,
	namespace string,
	env map[string]string) (*batchv1.Job, error) {

	backoffLimit := int32(2)
	terminationGracePeriodSeconds := int64(30)
	ttlSecondsAfterFinished := int32(15)
	pullPolicy := corev1.PullIfNotPresent

	labels := make(map[string]string)

	var jobEnvVars []corev1.EnvVar
	for envVarName, envVarValue := range env {
		jobEnvVars = append(jobEnvVars, corev1.EnvVar{Name: envVarName, Value: envVarValue})
	}

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            serviceAccountName,
					RestartPolicy:                 "Never",
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers: []corev1.Container{
						{
							Name:            name + "-container",
							Image:           images.GetWebhookCertJobImage(),
							ImagePullPolicy: pullPolicy,
							Env:             jobEnvVars,
						},
					},
				},
			},
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			BackoffLimit:            &backoffLimit,
		},
	}

	return job, nil
}
