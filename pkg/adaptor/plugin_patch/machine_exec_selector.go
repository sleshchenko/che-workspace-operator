package plugin_patch

import (
	"fmt"
	"github.com/eclipse/che-plugin-broker/model"
)

const (
	podSelectorCmdLineArg = "--pod-selector"
	podSelectorFmt        = "controller.devfile.io/workspace_id=%s"
)

// PatchMachineSelector overrides the pod selector used to find the workspace pod on the cluster, in order to
// allow opening a terminal into the pod. This is required to enable compatability with existing plugin registries,
// where the default value results in che-machine-exec attempting to use the `che.workspace_id` label.
//
// Function is a no-op if plugin is not named "che-machine-exec-plugin"
func PatchMachineSelector(plugin *model.ChePlugin, workspaceId string) {
	if plugin.Name != "che-machine-exec-plugin" {
		return
	}
	for i, container := range plugin.Containers {
		plugin.Containers[i].Command = append(container.Command, podSelectorCmdLineArg, fmt.Sprintf(podSelectorFmt, workspaceId))
	}
}
