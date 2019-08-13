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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JobSource is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
type JobSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the JobSource (from the client).
	// +required
	Spec JobSourceSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the JobSource (from the controller).
	// +optional
	Status JobSourceStatus `json:"status,omitempty"`
}

// Check that JobSource can be validated and defaulted.
var _ apis.Validatable = (*JobSource)(nil)
var _ apis.Defaultable = (*JobSource)(nil)
var _ kmeta.OwnerRefable = (*JobSource)(nil)

// JobSourceSpec holds the desired state of the JobSource (from the client).
type JobSourceSpec struct {
	BaseSourceSpec  `json:",inline"`
	batchv1.JobSpec `json:",inline"`
}

// JobSourceStatus communicates the observed state of the JobSource (from the controller).
type JobSourceStatus struct {
	BaseSourceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JobSourceList is a list of JobSource resources
type JobSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []JobSource `json:"items"`
}
