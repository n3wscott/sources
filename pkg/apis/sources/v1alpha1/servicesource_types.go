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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/kmeta"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceSource is a Knative abstraction that encapsulates the interface by which Knative
// components express a desire to have a particular image cached.
type ServiceSource struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the ServiceSource (from the client).
	// +required
	Spec ServiceSourceSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the ServiceSource (from the controller).
	// +optional
	Status ServiceSourceStatus `json:"status,omitempty"`
}

// Check that ServiceSource can be validated and defaulted.
var _ apis.Validatable = (*ServiceSource)(nil)
var _ apis.Defaultable = (*ServiceSource)(nil)
var _ kmeta.OwnerRefable = (*ServiceSource)(nil)

// ServiceSourceSpec holds the desired state of the ServiceSource (from the client).
type ServiceSourceSpec struct {
	BaseSourceSpec             `json:",inline"`
	servingv1beta1.ServiceSpec `json:",inline"`
}

// ServiceSourceStatus communicates the observed state of the ServiceSource (from the controller).
type ServiceSourceStatus struct {
	BaseSourceStatus `json:",inline"`

	// ServiceSource is an AddressableType via inheriting its Service's address status.
	// This enables the ServiceSource to also be a sink.
	duckv1beta1.AddressStatus `json:",inline"`

	// URL is the ServiceSource's pretty/public address.
	URL *apis.URL `json:"url,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceSourceList is a list of ServiceSource resources
type ServiceSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ServiceSource `json:"items"`
}

func (s *ServiceSource) GetSink() apisv1alpha1.Destination {
	return s.Spec.Sink
}

func (s *ServiceSource) GetStatus() SourceStatus {
	return &s.Status
}
