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

package config

import (
	"errors"
	"fmt"
	"os"
)

const (
	webTerminalToolingImage = "IMAGE_web_terminal_tooling"
)

func GetWebTerminalToolingImage() (string, error) {
	return GetImage(webTerminalToolingImage)
}

// get image returns the value for the image with the specified name
// it's expected to be in format [$registry_]$name_$tag
func GetImage(image string) (string, error) { //TODO Maybe the argument may be the full env vars, or just as it's done here
	// the images are expected to be configured as IMAGE_$image_name
	val, ok := os.LookupEnv("IMAGE_" + image)
	if !ok {
		return "", errors.New(fmt.Sprintf("The requested image %s is not configured", image))
	}
	return val, nil
}
