// Copyright 2017 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package general

import (
	"fmt"
	"strings"

	"github.com/coreos/ignition/v2/tests/register"
	"github.com/coreos/ignition/v2/tests/types"
)

func init() {
	// Base and provider configs are both applied
	register.Register(register.PositiveTest, makePreemptTest("BP"))
	// Base and user configs are applied; provider config is ignored
	register.Register(register.PositiveTest, makePreemptTest("BUp"))
	// No configs provided; Ignition should still run successfully
	register.Register(register.PositiveTest, makePreemptTest(""))
}

// makePreemptTest returns a config preemption test that executes some
// combination of "b"ase, "u"ser (user.ign), and "p"rovider
// (IGNITION_CONFIG_FILE) configs. Capital letters indicate configs that
// Ignition should apply.
func makePreemptTest(components string) types.Test {
	longnames := map[string]string{
		"b": "base",
		"u": "user",
		"p": "provider",
	}
	makeConfig := func(component string) string {
		return fmt.Sprintf(`{
			"ignition": {"version": "3.0.0"},
			"storage": {
				"files": [{
					"path": "/ignition/%s",
					"contents": {"source": "data:,%s"},
					"overwrite": true
				}]}
		}`, longnames[component], component)
	}
	enabled := func(component string) bool {
		return strings.Contains(strings.ToLower(components), component)
	}

	componentsSlice := strings.Split(strings.ToLower(components), "")
	longnameList := make([]string, len(componentsSlice))
	for _, component := range componentsSlice {
		longnameList = append(longnameList, longnames[component])
	}
	if len(longnameList) == 0 {
		longnameList = append(longnameList, "no")
	}
	name := "preemption." + strings.Join(longnameList, ".") + ".config"

	in := types.GetBaseDisk()
	out := types.GetBaseDisk()

	var config string
	if enabled("p") {
		config = makeConfig("p")
	}

	var systemFiles []types.File
	for _, component := range []string{"b", "u"} {
		if enabled(component) {
			var dir string
			if component == "b" {
				dir = "base.d"
			}
			systemFiles = append(systemFiles, types.File{
				Node: types.Node{
					Name:      longnames[component] + ".ign",
					Directory: dir,
				},
				Contents: makeConfig(component),
			})
		}
	}

	for component, longname := range longnames {
		in[0].Partitions.AddFiles("ROOT", []types.File{
			{
				Node: types.Node{
					Name:      longname,
					Directory: "ignition",
				},
				Contents: "unset",
			},
		})
		result := "unset"
		if strings.Contains(components, strings.ToUpper(component)) {
			result = component
		}
		out[0].Partitions.AddFiles("ROOT", []types.File{
			{
				Node: types.Node{
					Name:      longname,
					Directory: "ignition",
				},
				Contents: result,
			},
		})
	}

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		Config:            config,
		SystemDirFiles:    systemFiles,
		ConfigShouldBeBad: true,
	}
}
