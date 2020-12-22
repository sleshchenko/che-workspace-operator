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

package advdiff

import (
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func IgnoreChildrenExcept(typ interface{}, parent string, names ...string) cmp.Option {
	children := make([]string, len(names))
	for i, child := range names {
		children[i] = parent + "." + child
	}

	return cmp.FilterPath(func(p cmp.Path) bool {
		//parent should be walked through
		if p.String() == parent {
			return false
		}

		for _, child := range children {
			if strings.HasPrefix(p.String(), child) {
				return false
			}
		}

		//fall through to IgnoreFields
		return true
	}, //ignore all parents fields which was not excluded above
		cmpopts.IgnoreFields(typ, parent))
}
