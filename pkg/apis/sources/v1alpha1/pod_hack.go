/*
Copyright 2019 The Knative Authors.

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

package v1alpha1

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"

	"github.com/n3wscott/sources/pkg/sidecar"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SourcePod is an alias for an actual Pod which allows us to put methods on it.
type SourcePod corev1.Pod

// Assert that it satisfies the webhook.GenericCRD interface
var _ apis.Defaultable = &SourcePod{}
var _ apis.Validatable = &SourcePod{}
var _ runtime.Object = &SourcePod{}

func (s *SourcePod) SetDefaults(ctx context.Context) {
	pod := (*corev1.Pod)(s)
	if sidecar.ShouldInjectAdapter(pod) {
		log.Println("Adding adapter")
		sidecar.Inject(pod)
	}
}

func (s *SourcePod) Validate(context.Context) *apis.FieldError {
	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Pods have Binding subresources that we have to pretend to handle.
type NOPBinding corev1.Binding

func (_ *NOPBinding) SetDefaults(ctx context.Context) {
	// nop
}

func (_ *NOPBinding) Validate(context.Context) *apis.FieldError {
	return nil
}
