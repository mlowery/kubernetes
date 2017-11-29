package ebay

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/lifecycle"
)

const (
	WaitForHostnameAnnotation = "naming.tess.io/wait-for-hostname"
	WaitForHostnameReason     = "WaitForHostname"
)

func NewWaitForHostnameAdmitHandler() lifecycle.PodAdmitHandler {
	return &waitForHostnameAdmitHandler{}
}

type waitForHostnameAdmitHandler struct {}

func (a *waitForHostnameAdmitHandler) Admit(attrs *lifecycle.PodAdmitAttributes) lifecycle.PodAdmitResult {
	glog.V(10).Infof("waitForHostname: begin")
	// If the pod is already running or terminated, no need to recheck.
	if attrs.Pod.Status.Phase != v1.PodPending {
		glog.V(10).Infof("waitForHostname: pod not pending")
		return lifecycle.PodAdmitResult{Admit: true}
	}

	// no annotations at all means we shouldn't wait on spec.hostname
	annotations := attrs.Pod.Annotations
	if annotations == nil {
		glog.V(10).Infof("waitForHostname: annotations nil")
		return lifecycle.PodAdmitResult{Admit: true}
	}

	// missing annotation means we shouldn't wait on spec.hostname
	if _, ok := annotations[WaitForHostnameAnnotation]; !ok {
		glog.V(10).Infof("waitForHostname: no wait-for-hostname annotation")
		return lifecycle.PodAdmitResult{Admit: true}
	}

	if wait, err := strconv.ParseBool(annotations[WaitForHostnameAnnotation]); err != nil {
		glog.V(10).Infof("waitForHostname: annotation cannot be parsed")
		return lifecycle.PodAdmitResult{
			Admit:   false,
			Reason:  WaitForHostnameReason,
			Message: fmt.Sprintf("Error parsing value for annotation %q: %v", WaitForHostnameAnnotation, err),
		}
	} else if !wait {
		glog.V(10).Infof("waitForHostname: annotation is false")
		return lifecycle.PodAdmitResult{Admit: true}
	}

	// no spec.hostname so wait
	if attrs.Pod.Spec.Hostname == "" {
		glog.V(10).Infof("waitForHostname: spec.hostname not available yet")
		return lifecycle.PodAdmitResult{
			Admit:   false,
			Reason:  WaitForHostnameReason,
			Message: fmt.Sprintf("Pod spec.hostname not available yet"),
		}
	}

	glog.V(10).Infof("waitForHostname: end")
	// pod is ready to admit according to this handler
	return lifecycle.PodAdmitResult{Admit: true}
}