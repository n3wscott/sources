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

package resources

import (
	"fmt"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"knative.dev/pkg/kmeta"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
)

const (
	labelKey = "sources.knative.dev/servicesource"
)

func MakeService(source *v1alpha1.ServiceSource) *servingv1beta1.Service {
	podTemplate := &source.Spec.Template
	podTemplate.Labels = reconciler.Labels(source, labelKey)
	podTemplate.Annotations = reconciler.Annotations(source)

	containers := []corev1.Container{}
	for i, c := range podTemplate.Spec.Containers {
		if c.Name == "" {
			c.Name = fmt.Sprintf("servicesource%d", i)
		}
		c.Env = append(c.Env, corev1.EnvVar{Name: "K_SINK", Value: source.Status.SinkURI})
		c.Env = append(c.Env, corev1.EnvVar{Name: "K_OUTPUT_FORMAT", Value: string(source.Spec.OutputFormat)})
		containers = append(containers, c)
	}
	podTemplate.Spec.Containers = containers

	service := &servingv1beta1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ServiceName(source.GetObjectMeta()),
			Namespace:       source.GetObjectMeta().GetNamespace(),
			Labels:          reconciler.Labels(source, labelKey),
			Annotations:     reconciler.Annotations(source),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(source)},
		},
		Spec: source.Spec.ServiceSpec,
	}

	// TODO(spencer-p) Bubble up some annotations about creator, etc
	return service
}

func ServiceName(owner metav1.Object) string {
	// For now, this just returns the owner's name.
	return owner.GetName()
}
