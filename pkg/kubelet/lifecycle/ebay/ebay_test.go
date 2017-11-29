package ebay

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/lifecycle"
)

func TestWaitForHostnameAdmitHandlerAdmit(t *testing.T) {
	var testData = []struct {
		pod          *v1.Pod
		admitExpected bool
	}{
		{ // not pending
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
				},
			},
			admitExpected: true,
		},
		{ // nil annotations
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
			},
			admitExpected: true,
		},
		{ // empty annotations
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			admitExpected: true,
		},
		{ // unparseable value
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{WaitForHostnameAnnotation: "not-a-bool"},
				},
			},
			admitExpected: false,
		},
		{ // value is false
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{WaitForHostnameAnnotation: "false"},
				},
			},
			admitExpected: true,
		},
		{ // no hostname
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{WaitForHostnameAnnotation: "true"},
				},
			},
			admitExpected: false,
		},
		{ // told to wait and value is now set
			pod:          &v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodPending,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{WaitForHostnameAnnotation: "true"},
				},
				Spec: v1.PodSpec{
					Hostname: "host1",
				},
			},
			admitExpected: true,
		},
	}
	h := NewWaitForHostnameAdmitHandler()
	for i, test := range testData {
		t.Logf("processing test data: %v", i)
		res := h.Admit(&lifecycle.PodAdmitAttributes{Pod: test.pod})
		if res.Admit != test.admitExpected {
			t.Fatalf("expected %v but got %v", test.admitExpected, res.Admit)
		} else {
			t.Logf("got res: %v", res)
		}
	}
}