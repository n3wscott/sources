/*
Copyright 2019 The Knative Authors

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

package testing

import (
	"context"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler/servicesource/resources"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
)

type ServiceSourceOption func(*v1alpha1.ServiceSource)

func NewServiceSource(name string, options ...ServiceSourceOption) *v1alpha1.ServiceSource {
	s := &v1alpha1.ServiceSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}

	for _, option := range options {
		option(s)
	}

	s.SetDefaults(context.Background())
	return s
}

func WithMinServiceSpec(s *v1alpha1.ServiceSource) {
	s.Spec.ServiceSpec = servingv1beta1.ServiceSpec{
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
}

type ServiceOption func(*servingv1beta1.Service)

func NewService(s *v1alpha1.ServiceSource, options ...ServiceOption) *servingv1beta1.Service {
	svc := resources.MakeService(s)

	for _, option := range options {
		option(svc)
	}

	return svc
}
