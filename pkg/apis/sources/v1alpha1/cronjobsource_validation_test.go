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

	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func TestCronJobSourceValidation(t *testing.T) {
	tests := []struct {
		name string
		s    *CronJobSource
		want string
	}{{
		name: "all perfect",
		s: &CronJobSource{Spec: CronJobSourceSpec{BaseSourceSpec: BaseSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				// None of these fields have to be meaningful
				Name:       "Steve",
				APIVersion: "42",
				Kind:       "Service",
			}},
		}}},
		want: ``,
	}, {
		name: "bad sink shows up in spec field",
		s: &CronJobSource{Spec: CronJobSourceSpec{BaseSourceSpec: BaseSourceSpec{
			OutputFormat: OutputFormatBinary,
			Sink: apisv1alpha1.Destination{ObjectReference: &corev1.ObjectReference{
				APIVersion: "42",
				Kind:       "Service",
			}},
		}}},
		want: `missing field(s): spec.sink.name`,
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
