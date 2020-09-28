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

package cluster

import (
	"context"
	"time"

	webhook_k8s "github.com/devfile/devworkspace-operator/pkg/webhook/kubernetes"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CleanJob cleans up a job in a given namespace
func CleanJob(client client.Client, name, namespace string) error {
	job, err := GetJobInNamespace(client, name, namespace)
	if err != nil {
		return err
	}

	err = DeleteJob(client, job)
	if err != nil {
		return err
	}
	return nil
}

// GetJobInNamespace finds a job with a given name in a namespace
func GetJobInNamespace(client client.Client, name string, namespace string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, job)
	if err != nil {
		return job, err
	}
	return job, nil
}

// DeleteJob deletes a given job and cleans up any pods associated with it
func DeleteJob(client client.Client, job *batchv1.Job) error {
	err := CleanupPods(client, job)
	if err != nil {
		return err
	}
	err = client.Delete(context.TODO(), job)
	if err != nil {
		log.Error(err, "Error deleting job: "+job.Name)
		return err
	}
	return nil
}

// Wait for the job to complete. Times out if the job isn't complete after $(timeout) seconds
func WaitForJobCompletion(client client.Client, name string, namespace string, timeout time.Duration) error {
	const interval = 1 * time.Second
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		job, err := GetJobInNamespace(client, name, namespace)
		if err != nil {
			return false, err
		}

		if job.Status.Succeeded > 0 {
			log.Info("Please import public part of DevWorkspace self-signed CA certificate from " + webhook_k8s.TLSSelfSignedCertificateSecretName + " secret into your browser.")
			return true, nil
		}
		return false, nil
	})
}

func SyncJobToCluster(
	client client.Client,
	ctx context.Context,
	specJob *batchv1.Job,
) error {
	if err := client.Create(ctx, specJob); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		existingCfg, err := GetJobInNamespace(client, specJob.GetName(), specJob.Namespace)
		if err != nil {
			return err
		}
		specJob.ResourceVersion = existingCfg.ResourceVersion
		err = client.Update(ctx, specJob)
		if err != nil {
			return err
		}
		log.Info("Updated Job " + specJob.GetName())
	} else {
		log.Info("Created Job" + specJob.GetName())
	}

	return nil
}
