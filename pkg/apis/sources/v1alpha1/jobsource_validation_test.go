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

	corev1 "k8s.io/api/core/v1"
)

func TestJobSourceValidation(t *testing.T) {
	tests := []struct {
		name string
		js   *JobSource
		want string
	}{{
		name: "all zeroes",
		js:   &JobSource{},
		want: `missing field(s): spec.outputFormat, spec.sink`,
	}, {
		name: "all perfect",
		js: &JobSource{Spec: JobSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: &corev1.ObjectReference{
				// None of these fields have to be meaningful
				Name:       "Steve",
				APIVersion: "42",
				Kind:       "Service",
			},
		}},
		want: ``,
	}, {
		name: "no sink name",
		js: &JobSource{Spec: JobSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: &corev1.ObjectReference{
				APIVersion: "42",
				Kind:       "Service",
			},
		}},
		want: `missing field(s): spec.sink.name`,
	}, {
		name: "missing sink api version",
		js: &JobSource{Spec: JobSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: &corev1.ObjectReference{
				Name: "Steve",
				Kind: "Service",
			},
		}},
		want: `missing field(s): spec.sink.apiVersion`,
	}, {
		name: "missing sink kind",
		js: &JobSource{Spec: JobSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: &corev1.ObjectReference{
				Name:       "Steve",
				APIVersion: "42",
			},
		}},
		want: `missing field(s): spec.sink.kind`,
	}, {
		name: "invalid outputformat",
		js: &JobSource{Spec: JobSourceSpec{
			OutputFormat: "messenger_pigeon",
			Sink: &corev1.ObjectReference{
				Name:       "Steve",
				APIVersion: "42",
				Kind:       "Service",
			},
		}},
		want: `invalid value: messenger_pigeon: spec.outputFormat`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := test.js.Validate(context.Background())
			if got := errs.Error(); got != test.want {
				t.Errorf("%s: JobSource.Validate() = '%s', wanted '%s'", test.name, got, test.want)
			}
		})
	}
}
