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

package codify

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/kris-nova/logger"
	appsv1 "k8s.io/api/apps/v1"
)

type StatefulSet struct {
	KubeObject *appsv1.StatefulSet
	GoName     string
}

func NewStatefulSet(obj *appsv1.StatefulSet) *StatefulSet {
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	obj.Status = appsv1.StatefulSetStatus{}
	return &StatefulSet{
		KubeObject: obj,
		GoName:     goName(obj.Name),
	}
}

func (k StatefulSet) Install() (string, []string) {

	// Ignore resource requirements
	for i, _ := range k.KubeObject.Spec.Template.Spec.InitContainers {
		k.KubeObject.Spec.Template.Spec.InitContainers[i].Resources = v1.ResourceRequirements{}
	}
	for i, _ := range k.KubeObject.Spec.Template.Spec.Containers {
		k.KubeObject.Spec.Template.Spec.Containers[i].Resources = v1.ResourceRequirements{}
	}

	c, err := Literal(k.KubeObject)
	if err != nil {
		logger.Debug(err.Error())
	}
	l := c.Source
	packages := c.Packages
	install := fmt.Sprintf(`
	{{ .GoName }}StatefulSet := %s
	x.objects = append(x.objects, {{ .GoName }}StatefulSet)

	if client != nil {
		_, err = client.AppsV1().StatefulSets("{{ .KubeObject.Namespace }}").Create(context.TODO(), {{ .GoName }}StatefulSet, v1.CreateOptions{})
		if err != nil {
			return err
		}
	}
`, l)
	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl.Parse(install)
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, k)
	if err != nil {
		logger.Debug(err.Error())
	}
	return alias(buf.String(), "appsv1"), packages
}

func (k StatefulSet) Uninstall() string {
	uninstall := `
	if client != nil {
		err = client.AppsV1().StatefulSets("{{ .KubeObject.Namespace }}").Delete(context.TODO(), "{{ .KubeObject.Name }}", metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
 `
	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl.Parse(uninstall)
	buf := &bytes.Buffer{}
	k.KubeObject.Name = sanitizeK8sObjectName(k.KubeObject.Name)
	err := tpl.Execute(buf, k)
	if err != nil {
		logger.Debug(err.Error())
	}
	return buf.String()
}
