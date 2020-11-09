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

package library

import (
	"fmt"

	"github.com/devfile/api/pkg/apis/workspaces/v1alpha1"
)

func getCommandType(command v1alpha1.Command) (v1alpha1.CommandType, error) {
	err := command.Normalize()
	if err != nil {
		return "", err
	}
	return command.CommandType, nil
}

func getCommandsForIDs(ids []string, commands []v1alpha1.Command) ([]v1alpha1.Command, error) {
	var resolvedCommands []v1alpha1.Command

	for _, id := range ids {
		resolvedCommand, err := getCommandByID(id, commands)
		if err != nil {
			return nil, err
		}
		resolvedCommands = append(resolvedCommands, *resolvedCommand)
	}

	return resolvedCommands, nil
}

func getCommandByID(id string, commands []v1alpha1.Command) (*v1alpha1.Command, error) {
	for _, command := range commands {
		commandID, err := command.Key()
		if err != nil {
			return nil, err
		}
		if commandID == id {
			return &command, nil
		}
	}
	return nil, fmt.Errorf("no command with key %s is defined", id)
}

func commandListToComponentID(commands []v1alpha1.Command) (map[string]bool, error) {
	componentIDs := map[string]bool{}
	for _, command := range commands {
		commandType, err := getCommandType(command)
		if err != nil {
			return nil, err
		}
		switch commandType {
		case v1alpha1.ApplyCommandType:
			componentIDs[command.Apply.Component] = true
		case v1alpha1.ExecCommandType:
			// TODO: This will require special handling (how do we handle prestart exec?)
			componentIDs[command.Exec.Component] = true
		case v1alpha1.CompositeCommandType:
			// TODO: Handle composite commands: what if an init command is composite and refers to other commands
		default: // Ignore
		}
	}
	return componentIDs, nil
}

func removeCommandsByIDs(IDs []string, commands []v1alpha1.Command) ([]v1alpha1.Command, error) {
	toRemove := stringListToMap(IDs)
	var filtered []v1alpha1.Command
	for _, command := range commands {
		key, err := command.Key()
		if err != nil {
			return nil, err
		}
		if !toRemove[key] {
			filtered = append(filtered, command)
		}
	}
	return filtered, nil
}
