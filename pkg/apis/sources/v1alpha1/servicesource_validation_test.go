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

	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"

	corev1 "k8s.io/api/core/v1"
)

var (
	minServiceSpec = servingv1beta1.ServiceSpec{
		RouteSpec: servingv1beta1.RouteSpec{
			Traffic: []servingv1beta1.TrafficTarget{{Percent: 100}},
		},
		ConfigurationSpec: servingv1beta1.ConfigurationSpec{
			Template: servingv1beta1.RevisionTemplateSpec{
				Spec: servingv1beta1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name:  "Steve",
								Image: "example.com",
							},
						},
					},
				},
			},
		},
	}
)

func TestServiceSourceValidation(t *testing.T) {
	tests := []struct {
		name string
		js   *ServiceSource
		want string
	}{{
		name: "all perfect",
		js: &ServiceSource{Spec: ServiceSourceSpec{
			BaseSourceSpec: BaseSourceSpec{
				OutputFormat: OutputFormatBinary,
				Sink: &corev1.ObjectReference{
					// None of these fields have to be meaningful
					Name:       "Steve",
					APIVersion: "42",
					Kind:       "Service",
				},
			},
			ServiceSpec: minServiceSpec,
		}},
		want: ``,
	}, {
		name: "missing service spec",
		js: &ServiceSource{Spec: ServiceSourceSpec{
			BaseSourceSpec: BaseSourceSpec{
				OutputFormat: OutputFormatBinary,
				Sink: &corev1.ObjectReference{
					// None of these fields have to be meaningful
					Name:       "Steve",
					APIVersion: "42",
					Kind:       "Service",
				},
			},
		}},
		want: (&servingv1beta1.ServiceSpec{}).Validate(context.Background()).ViaField("spec").Error(),
	}, {
		name: "bad sink shows up in spec field",
		js: &ServiceSource{Spec: ServiceSourceSpec{
			BaseSourceSpec: BaseSourceSpec{
				OutputFormat: OutputFormatBinary,
				Sink: &corev1.ObjectReference{
					APIVersion: "42",
					Kind:       "Service",
				}},
			ServiceSpec: minServiceSpec,
		}},
		want: `missing field(s): spec.sink.name`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := test.js.Validate(context.Background())
			if got := errs.Error(); got != test.want {
				t.Errorf("Validate() = %q, wanted %q", got, test.want)
			}
		})
	}
}
