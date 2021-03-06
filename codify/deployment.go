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

type Deployment struct {
	KubeObject *appsv1.Deployment
	GoName     string
}

func NewDeployment(obj *appsv1.Deployment) *Deployment {
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	obj.Status = appsv1.DeploymentStatus{}
	return &Deployment{
		KubeObject: obj,
		GoName:     goName(obj.Name),
	}
}

func (k Deployment) Install() (string, []string) {

	// We do not guarantee the NAML is perfect.
	//
	// Because Pod resources are a factory, we cannot literally set the values.
	// This turns out to be somewhat reasonable as we are defining resource parameters
	// that could possibly be looked up at runtime anyway.
	//
	// For now, we ignore the resource requirements.
	for i, _ := range k.KubeObject.Spec.Template.Spec.InitContainers {
		k.KubeObject.Spec.Template.Spec.InitContainers[i].Resources = v1.ResourceRequirements{}
	}
	for i, _ := range k.KubeObject.Spec.Template.Spec.Containers {
		k.KubeObject.Spec.Template.Spec.Containers[i].Resources = v1.ResourceRequirements{}
	}

	c, err := Literal(k.KubeObject)
	if err != nil {
		logger.Critical(err.Error())
	}
	l := c.Source
	packages := c.Packages

	install := fmt.Sprintf(`
	// Adding a deployment: "{{ .KubeObject.Name }}"
	{{ .GoName }}Deployment := %s
	x.objects = append(x.objects, {{ .GoName }}Deployment)

	if client != nil {
		_, err = client.AppsV1().Deployments("{{ .KubeObject.Namespace }}").Create(context.TODO(), {{ .GoName }}Deployment, v1.CreateOptions{})
		if err != nil {
			return err
		}
	}
`, l)

	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl, err = tpl.Parse(install)
	if err != nil {
		logger.Critical(err.Error())
	}
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, k)
	if err != nil {
		logger.Critical(err.Error())
	}
	return alias(buf.String(), "appsv1"), packages
}

func (k Deployment) Uninstall() string {
	uninstall := `
	if client != nil {
		err = client.AppsV1().Deployments("{{ .KubeObject.Namespace }}").Delete(context.TODO(), "{{ .KubeObject.Name }}", metav1.DeleteOptions{})
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
