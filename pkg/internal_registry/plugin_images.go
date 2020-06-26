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

package registry

import (
	"regexp"

	"github.com/devfile/devworkspace-operator/pkg/config"
	"github.com/eclipse/che-plugin-broker/model"
)

var imagePlaceHolder = regexp.MustCompile(`$\${(.*)}^`)

// fillPluginMeta fills in the configured image for plugin from internal plugin registry
func fillPluginMeta(pluginMeta model.PluginMeta) (model.PluginMeta, error) {
	for idx, container := range pluginMeta.Spec.Containers {
		img, err := resolvePlaceholders(container.Image)
		if err != nil {
			return model.PluginMeta{}, err
		}
		pluginMeta.Spec.Containers[idx].Image = img
	}
	for idx, initContainer := range pluginMeta.Spec.InitContainers {
		img, err := resolvePlaceholders(initContainer.Image)
		if err != nil {
			return model.PluginMeta{}, err
		}
		pluginMeta.Spec.InitContainers[idx].Image = img
	}
	return pluginMeta, nil
}

func isImagePlaceHolder(query string) bool {
	return imagePlaceHolder.MatchString(query)
}

func resolvePlaceholders(pluginImage string) (string, error) {
	if !isImagePlaceHolder(pluginImage) {
		// Value passed in is not env var, return unmodified
		return pluginImage, nil
	}
	matches := imagePlaceHolder.FindStringSubmatch(pluginImage)
	env := matches[1]
	return config.GetImage(env)
}
