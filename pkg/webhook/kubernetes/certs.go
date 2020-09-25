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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// removeCACertificate removes a CA Cert. Used to clear out the old CACert when we are creating a new one
func removeCACertificate(client crclient.Client, namespace string) error {
	caSelfSignedCertificateSecret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: TLSSelfSignedCertificateSecretName}, caSelfSignedCertificateSecret)
	if err != nil && !errors.IsNotFound(err) {
		log.Error(err, "Error getting self-signed certificate secret " + TLSSelfSignedCertificateSecretName)
		return err
	} else if err != nil && errors.IsNotFound(err) {
		// We don't have anything to remove in this case since its already not found
		return nil
	}

	// Remove CA secret because TLS secret is missing (they should be generated together).
	if err = client.Delete(context.TODO(), caSelfSignedCertificateSecret); err != nil {
		log.Error(err, "Error deleting self-signed certificate secret " + TLSSelfSignedCertificateSecretName, )
		return err
	}

	return nil
}
