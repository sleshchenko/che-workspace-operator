//
// Copyright (c) 2012-2019 Red Hat, Inc.
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
	"github.com/devfile/devworkspace-operator/pkg/webhook/common"
	"github.com/devfile/devworkspace-operator/webhook/server"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"

	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logf.Log.WithName("webhook-k8s")

// TLS related constants
const (
	TLSJobName                         = "devworkspace-tls-job"
	TLSSelfSignedCertificateSecretName = "devworkspace-self-signed-certificate"
	TLSDomain                          = "devworkspace-webhookserver.devworkspace-controller.svc"
)

// SetupKubernetesWebhookCerts handles TLS secrets required for deployment on Kubernetes.
func SetupKubernetesWebhookCerts(client crclient.Client, ctx context.Context, namespace string) error {
	devworkspaceSecret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: server.WebhookServerTLSSecretName}, devworkspaceSecret)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "Error getting Che TLS secret " + server.WebhookServerTLSSecretName)
			return err
		}

		// TLS secret doesn't exist so we need to generate a new one

		// Remove CA certificate secret if any
		err = removeCACertificate(client, namespace)
		if err != nil {
			return err
		}

		jobEnvVars := map[string]string{
			"DOMAIN":                         TLSDomain,
			"CHE_NAMESPACE":                  namespace,
			"CHE_SERVER_TLS_SECRET_NAME":     server.WebhookServerTLSSecretName,
			"CHE_CA_CERTIFICATE_SECRET_NAME": TLSSelfSignedCertificateSecretName,
		}
		err := SyncJobToCluster(client, ctx, server.WebhookServerSAName, TLSJobName, namespace, jobEnvVars)
		if err != nil {
			return err
		}

		err = webhook_common.CreateSecureService(client, ctx, namespace, map[string]string{})
		if err != nil {
			log.Info("Failed creating the secure service")
			return err
		}

		// Wait a maximum of 60 seconds for the job to be completed
		err = waitForJobCompletion(client, TLSJobName, namespace, 60 * time.Second)
		if err != nil {
			return err
		}

		// Clean up everything related to the job now that it should be finished
		err = cleanJob(client, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}


