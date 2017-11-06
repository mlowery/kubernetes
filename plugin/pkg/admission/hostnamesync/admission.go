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

// Package hostnamesync contains an admission controller that modifies every new Pod to set the spec.hostname
// if an annotation instructs it to do so. spec.hostname matches metadata.name at the end of this admission controller.
package hostnamesync

import (
	"fmt"
	"io"
	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	podregistry "k8s.io/kubernetes/pkg/registry/core/pod"
)

const (
	Annotation = "pod.tess.io/hostname-sync"
)

func init() {
	admission.RegisterPlugin("HostnameSync", func(config io.Reader) (admission.Interface, error) {
		return NewHostnameSync(), nil
	})
}

// hostnameSync is an implementation of admission.Interface.
type hostnameSync struct {
	*admission.Handler
}

func (a *hostnameSync) Admit(attributes admission.Attributes) (err error) {
	// Ignore all calls to subresources or resources other than pods.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != api.Resource("pods") {
		return nil
	}
	pod, ok := attributes.GetObject().(*api.Pod)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}

	metadata, err := meta.Accessor(pod)
	if err != nil {
		return fmt.Errorf("error creating metadata accessor for pod %v: %v", pod, err)
	}

	if pod.Spec.Subdomain == "" || pod.Spec.Hostname != "" {
		// In the case of specified hostname, there is nothing to do--it's already specified.
		// In the case of missing subdomain, there is no work to do since this pod won't get a DNS name anyway and hostname
		// will be set to pod name for /etc/hosts.
		return nil
	}

	ann := metadata.GetAnnotations()
	if ann == nil {
		// no annotations at all
		return nil
	}

	if _, ok := ann[Annotation]; !ok {
		// annotation that we care about is missing
		return nil
	}

	b, err := strconv.ParseBool(ann[Annotation])
	if err != nil {
		return fmt.Errorf("could not parse annotation: %q", Annotation)
	}

	if !b {
		// annotation set to false
		return nil
	}

	name := metadata.GetName()
	if len(name) == 0 {
		generateName := metadata.GetGenerateName()
		// can't have empty name and empty generateName and ask to generate a hostname
		if len(generateName) == 0 {
			return fmt.Errorf("no name from which to generate hostname for pod %v", pod)
		}
		// set name now since it wouldn't usually get a name until later and we need the name now
		name = podregistry.Strategy.GenerateName(generateName)
		metadata.SetName(name)
	}

	pod.Spec.Hostname = name

	return nil
}

func NewHostnameSync() admission.Interface {
	// pod names cannot change so no need to handle anything other than create
	return &hostnameSync{
		Handler: admission.NewHandler(admission.Create),
	}
}
