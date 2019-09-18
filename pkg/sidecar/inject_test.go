package sidecar

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetArgsMissingAll(t *testing.T) {
	pod := corev1.Pod{}
	_, errs := constructArgs(&pod)
	want := `missing field(s): $K_SOURCE_ADAPTER_IMAGE, annotations.cloudevents.io/source, annotations.cloudevents.io/type, spec.containers[i].Env.K_OUTPUT_FORMAT, spec.containers[i].Env.K_SINK`
	if got := errs.Error(); got != want {
		t.Errorf("wanted %q, got %q", want, got)
	}
}

// TODO(spencer-p):
// - Test happy case
// - Test no ports available (start with 65535 with a container already mapped)
