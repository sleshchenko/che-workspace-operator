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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// cleanJob cleans up a job in a given namespace
func cleanJob(client crclient.Client, namespace string) error {
	job, err := getJobInNamespace(client, TLSJobName, namespace)
	if err != nil {
		return err
	}
	err = deleteJob(client, job)
	if err != nil {
		return err
	}
	return nil
}

// getJobInNamespace finds a job with a given name in a namespace
func getJobInNamespace(client crclient.Client, name string, namespace string) (*batchv1.Job, error){
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, job)
	if err != nil {
		return job, err
	}
	return job, nil
}

// deleteJob deletes a given job and cleans up any pods associated with it
func deleteJob(client crclient.Client, job *batchv1.Job) error {
	err := cleanupPods(client, job, TLSJobName)
	if err != nil {
		return err
	}
	err = client.Delete(context.TODO(), job)
	if err != nil {
		log.Error(err, "Error deleting job: " + job.Name)
		return err
	}
	return nil
}

// Wait for the job to complete. Times out if the job isn't complete after $(timeout) seconds
func waitForJobCompletion(client crclient.Client, name string, namespace string, timeout time.Duration) error {
	const interval = 1 * time.Second
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		job, err := getJobInNamespace(client, name, namespace)
		if err != nil {
			return false, err
		}

		if job.Status.Succeeded > 0 {
			log.Info("Please import public part of DevWorkspace self-signed CA certificate from " + TLSSelfSignedCertificateSecretName + " secret into your browser.", )
			return true, nil
		}
		return false, nil
	})
}

