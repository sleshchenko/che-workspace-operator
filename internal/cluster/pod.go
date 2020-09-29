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
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CleanupPods takes a job and cleans up all the pods associated with that job
func CleanupPods(client crclient.Client, job *batchv1.Job) error {
	selector := fmt.Sprintf("job-name=%s", job.GetName())
	pods, err := GetPodsBySelector(client, job.Namespace, selector)
	for _, pod := range pods.Items {

		// Get the pod from the cluster as a runtime object and then delete it
		clusterPod := &corev1.Pod{}
		err = client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: job.Namespace}, clusterPod)
		if err != nil {
			return err
		}
		err = client.Delete(context.TODO(), clusterPod)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPodsBySelector selects all pods by a selector in a namespace
func GetPodsBySelector(client crclient.Client, namespace string, selector string) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	labelSelector, err := labels.Parse(selector)
	if err != nil {
		return nil, err
	}
	listOptions := &crclient.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	}
	err = client.List(context.TODO(), podList, listOptions)
	if err != nil {
		return nil, err
	}
	return podList, nil
}
