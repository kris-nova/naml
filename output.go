//
// Copyright © 2021 Kris Nóva <kris@nivenly.com>
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
//
//   ███╗   ██╗ █████╗ ███╗   ███╗██╗
//   ████╗  ██║██╔══██╗████╗ ████║██║
//   ██╔██╗ ██║███████║██╔████╔██║██║
//   ██║╚██╗██║██╔══██║██║╚██╔╝██║██║
//   ██║ ╚████║██║  ██║██║ ╚═╝ ██║███████╗
//   ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝
//

package naml

import (
	"encoding/json"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

const (
	OutputYAML OutputEncoding = 0
	OutputJSON OutputEncoding = 1
)

type OutputEncoding int

func RunOutput(appName string, o OutputEncoding) error {
	app := Find(appName)
	if app == nil {
		return fmt.Errorf("unable to find app: %s", appName)
	}

	// Install the application "nowhere" to register the components in memory
	app.Install(nil)

	switch o {

	// ---- [ JSON ] ----
	case OutputJSON:
		return PrintJSON(app)

	// ---- [ YAML ] ----
	case OutputYAML:
		return PrintKubeYAML(app)

	// ---- [ DEFAULT ] ----
	default:
		return PrintKubeYAML(app)
	}
	return nil
}

func PrintKubeYAML(app Deployable) error {
	for i, obj := range app.Objects() {
		raw, err := yaml.Marshal(obj)
		if err != nil {
			return fmt.Errorf("unable to YAML marshal: %v", err)
		}
		lines := strings.Split(string(raw), `
`)
		for _, line := range lines {

			// And we are back to the fucking alias hackery again
			// I am going to open an issue to clean this up once
			// we have figured out all the crap we need to do
			line = strings.ReplaceAll(line, "corev1", "v1")
			line = strings.ReplaceAll(line, "rbacv1", "v1")
			line = strings.ReplaceAll(line, "metav1", "v1")
			line = strings.ReplaceAll(line, "appsv1", "v1")
			line = strings.ReplaceAll(line, "corev1", "v1")

			// Remove creationTimestamp
			if strings.Contains(line, "creationTimestamp") {
				continue
			}
			// Ignore status for each object
			if strings.Contains(line, "status") {
				break
			}
			fmt.Println(line)
		}
		if i < len(app.Objects())-1 {
			fmt.Println(YAMLDelimiter)
			fmt.Println()
		}
	}
	return nil
}

func PrintJSON(app Deployable) error {
	raw, err := json.MarshalIndent(app.Objects(), " ", "	")
	if err != nil {
		return fmt.Errorf("unable to JSON marshal: %v", err)
	}
	fmt.Println(string(raw))
	return nil
}
