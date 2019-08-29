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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CronJobSource is a CronJob with a Sink.
type CronJobSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the CronJobSource (from the client).
	// +required
	Spec CronJobSourceSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the CronJobSource (from the controller).
	// +optional
	Status CronJobSourceStatus `json:"status,omitempty"`
}

// Check that CronJobSource can be validated and defaulted.
var _ apis.Validatable = (*CronJobSource)(nil)
var _ apis.Defaultable = (*CronJobSource)(nil)
var _ kmeta.OwnerRefable = (*CronJobSource)(nil)

// CronJobSourceSpec holds the desired state of the CronJobSource (from the client).
type CronJobSourceSpec struct {
	BaseSourceSpec           `json:",inline"`
	batchv1beta1.CronJobSpec `json:",inline"`
}

// CronJobSourceStatus communicates the observed state of the CronJobSource (from the controller).
type CronJobSourceStatus struct {
	BaseSourceStatus           `json:",inline"`
	batchv1beta1.CronJobStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CronJobSourceList is a list of CronJobSource resources
type CronJobSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CronJobSource `json:"items"`
}

func (s *CronJobSource) GetSink() apisv1alpha1.Destination {
	return s.Spec.Sink
}

func (s *CronJobSource) GetStatus() SourceStatus {
	return &s.Status
}
