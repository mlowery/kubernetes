/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hostnamesync

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	"strings"
)

func TestAdmissionUseName(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations[Annotation] = "true"
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if pod.Name != "abc" {
		t.Errorf("Unexpected value for name: %q", pod.Name)
	}
	if pod.Spec.Hostname != "abc" {
		t.Errorf("Unexpected value for hostname: %q", pod.Spec.Hostname)
	}
}

func TestAdmissionUseGenerateName(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "abc-", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations[Annotation] = "true"
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if !strings.HasPrefix(pod.Name, "abc-") {
		t.Errorf("Unexpected value for name: %q", pod.Name)
	}
	if pod.Spec.Hostname != pod.Name {
		t.Errorf("Unexpected value for hostname: %q", pod.Spec.Hostname)
	}
}

func TestAdmissionErrOnNoNameAndNoGenerateName(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "", GenerateName: "", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations[Annotation] = "true"
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err == nil {
		t.Errorf("Expected error but gone none")
	}
}

func TestAdmissionIgnoreOnNoAnnotations(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if len(pod.Spec.Hostname) != 0 {
		t.Errorf("Unexpected hostname: %q", pod.Spec.Hostname)
	}
}

func TestAdmissionIgnoreOnMissingAnnotation(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations["foo"] = "bar"
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if len(pod.Spec.Hostname) != 0 {
		t.Errorf("Unexpected hostname: %q", pod.Spec.Hostname)
	}
}

func TestAdmissionErrorOnAnnotationGarbage(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations[Annotation] = "not true"
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err == nil {
		t.Errorf("Expected error but got none")
	}
}

func TestAdmissionIgnoreOnAnnotationNotTrue(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations[Annotation] = "false"
	pod.Spec.Subdomain = "sub"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if len(pod.Spec.Hostname) != 0 {
		t.Errorf("Unexpected hostname: %q", pod.Spec.Hostname)
	}
}

func TestAdmissionIgnoreOnMissingSubdomain(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Annotations = make(map[string]string)
	pod.Annotations["foo"] = "bar"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if len(pod.Spec.Hostname) != 0 {
		t.Errorf("Unexpected hostname: %q", pod.Spec.Hostname)
	}
}

func TestAdmissionIgnoreOnSetHostname(t *testing.T) {
	namespace := "test"
	handler := &hostnameSync{}
	pod := api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "abc", Namespace: namespace},
	}
	pod.Spec.Hostname = "foo"
	err := handler.Admit(admission.NewAttributesRecord(&pod, nil, api.Kind("Pod").WithVersion("version"), pod.Namespace, pod.Name, api.Resource("pods").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler")
	}
	if pod.Spec.Hostname != "foo" {
		t.Errorf("Unexpected hostname: %q", pod.Spec.Hostname)
	}
	if pod.Name != "abc" {
		t.Errorf("Unexpected pod name: %q", pod.Name)
	}
}

// TestOtherResources ensures that this admission controller is a no-op for other resources,
// subresources, and non-pods.
func TestOtherResources(t *testing.T) {
	namespace := "testnamespace"
	name := "testname"
	pod := &api.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "ctr2", Image: "image", ImagePullPolicy: api.PullNever},
			},
		},
	}
	tests := []struct {
		name        string
		kind        string
		resource    string
		subresource string
		object      runtime.Object
		expectError bool
	}{
		{
			name:     "non-pod resource",
			kind:     "Foo",
			resource: "foos",
			object:   pod,
		},
		{
			name:        "pod subresource",
			kind:        "Pod",
			resource:    "pods",
			subresource: "exec",
			object:      pod,
		},
		{
			name:        "non-pod object",
			kind:        "Pod",
			resource:    "pods",
			object:      &api.Service{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		handler := &hostnameSync{}

		err := handler.Admit(admission.NewAttributesRecord(tc.object, nil, api.Kind(tc.kind).WithVersion("version"), namespace, name, api.Resource(tc.resource).WithVersion("version"), tc.subresource, admission.Create, nil))

		if tc.expectError {
			if err == nil {
				t.Errorf("%s: unexpected nil error", tc.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		if e, a := api.PullNever, pod.Spec.Containers[0].ImagePullPolicy; e != a {
			t.Errorf("%s: image pull policy was changed to %s", tc.name, a)
		}
	}
}
