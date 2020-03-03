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
package creator

import (
	"context"
	"github.com/che-incubator/che-workspace-operator/pkg/controller/workspace/model"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// WorkspaceAnnotator annotates Workspaces
type WorkspaceValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

func (a *WorkspaceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	client, err := createClient()
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	p := v1.Pod{}
	err = client.Get(ctx, types.NamespacedName{
		Name:      req.Name,
		Namespace: req.Namespace,
	}, &p)

	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if p.Annotations == nil {
		p.Annotations = map[string]string{}
	}
	creator := p.Annotations[model.WorkspaceCreatorAnnotation]
	if creator != req.UserInfo.UID {
		return admission.Denied("The only workspace creator has exec access")
	}

	return admission.Allowed("Did nothing")
}

// WorkspaceAnnotator implements inject.Client.
// A client will be automatically injected.

// InjectClient injects the client.
func (a *WorkspaceValidator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// WorkspaceAnnotator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *WorkspaceValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}