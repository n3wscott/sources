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
	corev1 "k8s.io/api/core/v1"

	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// BaseSourceSpec provides spec components that all sources share. It does not include any
// PodSpecable helpers.
type BaseSourceSpec struct {
	// Sink is a reference to an object that will resolve to a URI to send events to.
	// +required
	Sink *corev1.ObjectReference `json:"sink,omitempty"`

	// OutputFormat describes the CloudEvent output format the source should send events in.
	// All formats are over HTTP.
	// Defaults to binary.
	// +optional
	OutputFormat OutputFormatType `json:"outputFormat,omitempty"`
}

// BaseSourceStatus holds status information that sources need. This base will not necessarily need
// to be extended.
type BaseSourceStatus struct {
	duckv1beta1.Status `json:",inline"`

	// SinkURI is the current sink URI configured for the source.
	// +optional
	SinkURI string `json:"sinkUri,omitempty"`
}

const (
	// SourceConditionSinkProvided represents the condition that a sink with a URI has been
	// provided to the source. All sources will use this condition and set it true when the
	// source is configured with a sink.
	SourceConditionSinkProvided apis.ConditionType = "SinkProvided"
)
