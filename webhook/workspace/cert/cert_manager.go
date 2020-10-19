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

package cert

import (
	"fmt"
	"strings"

	"github.com/devfile/devworkspace-operator/webhook/server"
	"github.com/google/martian/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

//holds logic related to webhook server certificates powered by https://cert-manager.io/

const (
	CertManagerAnnotationPrefix    = "cert-manager.io"
	CertManagerInjectKeyAnnotation = CertManagerAnnotationPrefix + "/inject-ca-from"

	//TODO It would be better to hardcode as less things as possible
	//We can only require secret name, and certificate can be evaluated from
	//secret annotation: cert-manager.io/certificate-name: devworkspace-webhook-certificate
	WebhookServerCertManagerCertificateName = "devworkspace-webhook-certificate"
)

func isCertManagerSecret(client client.Client, name, namespace string) (bool, error) {
	secret, err := getSecret(client, namespace, name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	isCertManagerAnnotated := hasCertManagerAnnotation(secret.GetAnnotations())
	return isCertManagerAnnotated, nil
}

func hasCertManagerAnnotation(annotations map[string]string) bool {
	for key, _ := range annotations {
		if strings.HasPrefix(key, CertManagerAnnotationPrefix) {
			return true
		}
	}
	return false
}

func GetWebhookCfgAnnotations(client crclient.Client, namespace string) (map[string]string, error) {
	webhookAnnotations := make(map[string]string)

	certManagerSecret, err := cert.isCertManagerSecret(client, server.WebhookServerTLSSecretName, namespace)
	if err != nil {
		log.Error(err, "Failed when attempting to check if the secret is annotated with cert manager")
		return nil, err
	}

	if certManagerSecret {
		webhookAnnotations[cert.CertManagerInjectKeyAnnotation] = fmt.Sprintf("%s/%s", namespace, cert.WebhookServerCertManagerCertificateName)
	}

	return webhookAnnotations, nil
}
