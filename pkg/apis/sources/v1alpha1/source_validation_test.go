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
	"testing"

	"knative.dev/pkg/apis/duck"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func TestSourceValidation(t *testing.T) {
	tests := []struct {
		name string
		s    *BaseSourceSpec
		want string
	}{{
		name: "all zeroes",
		s:    &BaseSourceSpec{},
		want: `expected exactly one, got neither: sink.uri, sink[apiVersion, kind, name]
missing field(s): outputFormat`,
	}, {
		name: "all perfect",
		s: &BaseSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				// None of these fields have to be meaningful
				Name:       "Steve",
				APIVersion: "42",
				Kind:       "Service",
			}}},
		want: ``,
	}, {
		name: "no sink name",
		s: &BaseSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				APIVersion: "42",
				Kind:       "Service",
			}}},
		want: `missing field(s): sink.name`,
	}, {
		name: "missing sink api version",
		s: &BaseSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				Name: "Steve",
				Kind: "Service",
			}}},
		want: `missing field(s): sink.apiVersion`,
	}, {
		name: "missing sink kind",
		s: &BaseSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				Name:       "Steve",
				APIVersion: "42",
			}}},
		want: `missing field(s): sink.kind`,
	}, {
		name: "invalid outputformat",
		s: &BaseSourceSpec{
			OutputFormat: "messenger_pigeon",
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				Name:       "Steve",
				APIVersion: "42",
				Kind:       "Service",
			}}},
		want: `invalid value: messenger_pigeon: outputFormat`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := test.s.Validate(context.Background())
			if got := errs.Error(); got != test.want {
				t.Errorf("Validate() = %q, wanted %q", got, test.want)
			}
		})
	}
}

func TestSourceDuckTypes(t *testing.T) {
	tests := []struct {
		name string
		t    interface{}
	}{{
		name: "jobsource",
		t:    &JobSource{},
	}, {
		name: "servicesource",
		t:    &ServiceSource{},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := duck.VerifyType(test.t, &duckv1beta1.Source{})
			if err != nil {
				t.Errorf("VerifyType(%T, duckv1beta1.Source) = %v", test.t, err)

			}

		})

	}
}
