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
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hexops/valast"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	KubernetesImportPackageMap = map[string]string{
		"k8s.io/api/apps/v1":                                       "appsv1",
		"k8s.io/api/batch/v1":                                      "batchv1",
		"k8s.io/api/core/v1":                                       "corev1",
		"k8s.io/apimachinery/pkg/apis/meta/v1":                     "metav1",
		"k8s.io/api/rbac/v1":                                       "rbacv1",
		"k8s.io/api/networking/v1":                                 "networkingv1",
		"k8s.io/api/admissionregistration/v1":                      "admissionregistrationv1",
		"k8s.io/api/policy/v1beta1":                                "policyv1beta1",
		"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1": "apiextensionsv1",
	}

	PolicyV1Types = []string{
		"PolicyV1Interface",
	}

	AppsV1Types = []string{""}

	MetaV1Types = []string{
		"APIGroup",
		"ObjectMeta",
		"Time",
		"TypeMeta",
		"ManagedFieldsEntry",
		"OwnerReference",
		"CreateOptions",
		"DeleteOptions",
		"LabelSelector",
	}

	CoreV1Types = []string{
		"WeightedPodAffinityTerm",
		"TCPSocketAction",
		"Volume",
		"SecretEnvSource",
		"SecretKeySelector",
		"SecretProjection",
		"ConfigMapKeySelector",
		"ConfigMapProjection",
		"ConfigMapEnvSource",
		"HTTPHeader",
		"PodAntiAffinity",
		"PodAffinityTerm",
		"KeyToPath",
		"SecretVolumeSource",
		"EmptyDirVolumeSource",
		"Handler",
		"TaintEffect",
		"HTTPGetAction",
		"URIScheme",
		"PodTemplateSpec",
		"PodSpec",
		"Protocol",
		"ResourceRequirements",
		"ResourceList",
		"VolumeDevice",
		"Probe",
		"Container",
		"EnvFromSource",
		"EnvVar",
		"VolumeMount",
		"Lifecycle",
		"SecurityContext",
		"EphemeralContainer",
		"LocalObjectReference",
		"Affinity",
		"Toleration",
		"HostAlias",
		"PodDNSConfig",
		"PodReadinessGate",
		"PreemptionPolicy",
		"TopologySpreadConstraint",
		"TerminationMessagePolicy",
		"PullPolicy",
		"RestartPolicy",
		"DNSPolicy",
		"ObjectFieldSelector",
		"PodSecurityContext",
		"ResourceName",
		"Capabilities",
		"Capability",
		"ExecAction",
		"HostPathVolumeSource",
		"HostPathType",
		"ProjectedVolumeSource",
		"MountPropagationMode",
		"ConfigMapVolumeSource",
		"ClaimName",
		"PersistentVolumeClaimVolumeSource",
	}
)

// cleanObjectMeta helps us get rid of things like timestamps
// by only "opting in" to certain fields.
func cleanObjectMeta(m metav1.ObjectMeta) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:                       m.Name,
		Namespace:                  m.Namespace,
		Labels:                     m.Labels,
		Annotations:                m.Annotations,
		ClusterName:                m.ClusterName,
		ResourceVersion:            m.ResourceVersion,
		Finalizers:                 m.Finalizers,
		Generation:                 m.Generation,
		GenerateName:               m.GenerateName,
		UID:                        m.UID,
		ManagedFields:              m.ManagedFields,
		OwnerReferences:            m.OwnerReferences,
		DeletionGracePeriodSeconds: m.DeletionGracePeriodSeconds,
	}
}

// alias will do it's best to manage package aliases in the source code
func alias(generated, defaultalias string) string {
	aliased := generated

	// Each object can pass in a "default" to use if we do not have it defined above.
	aliased = strings.ReplaceAll(aliased, "v1", defaultalias)
	for _, t := range AppsV1Types {
		if t == "" {
			continue
		}
		aliased = strings.ReplaceAll(aliased,
			fmt.Sprintf("%s.%s", defaultalias, t),
			fmt.Sprintf("appsv1.%s", t))
	}
	for _, t := range MetaV1Types {
		if t == "" {
			continue
		}
		aliased = strings.ReplaceAll(aliased,
			fmt.Sprintf("%s.%s", defaultalias, t),
			fmt.Sprintf("metav1.%s", t))
	}
	for _, t := range CoreV1Types {
		if t == "" {
			continue
		}
		aliased = strings.ReplaceAll(aliased,
			fmt.Sprintf("%s.%s", defaultalias, t),
			fmt.Sprintf("corev1.%s", t))
	}
	for _, t := range PolicyV1Types {
		if t == "" {
			continue
		}
		aliased = strings.ReplaceAll(aliased,
			fmt.Sprintf("%s.%s", defaultalias, t),
			fmt.Sprintf("policyv1beta1.%s", t)) // Note this is different from the others!

	}
	return aliased
}

func sanitizeK8sObjectName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9 \\-]+")
	return reg.ReplaceAllString(name, "")
}

func goName(name string) string {
	name = strings.ReplaceAll(name, ".", "")
	return strings.ReplaceAll(name, "-", "_")
}

// Literal will convert an abstract kubeobject interface{} to Go code.
//
// The various components relevant to each conversion are returned in a *Codified
// object.
func Literal(kubeobject interface{}) (*Codified, error) {
	l := valast.String(kubeobject)
	l = cleanValast20(l)
	r, err := valast.AST(reflect.ValueOf(kubeobject), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to source code: %v", err)
	}
	return &Codified{
		Source:   l,
		Packages: r.Packages,
		R:        r,
		Object:   kubeobject,
	}, nil
}

// Codified is the source code associated with a given Kubernetes object.
type Codified struct {

	// The unknown original object
	Object interface{}

	// The packages required to code each object
	Packages []string

	// The Go code to define each object
	Source string

	// The full result from valast
	R valast.Result
}

// cleanValast20 is a determinstic string mutation function
// that will address the bug described in https://github.com/hexops/valast/issues/20
func cleanValast20(input string) string {
	var output string
	output = strings.ReplaceAll(input, "{{", "{\n{")
	output = strings.ReplaceAll(output, "}}", "},\n}")
	return output
}
