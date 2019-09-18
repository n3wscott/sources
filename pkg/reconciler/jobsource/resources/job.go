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
	"knative.dev/eventing/pkg/utils"
	"knative.dev/pkg/kmeta"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	labelKey = "sources.knative.dev/jobsource"
)

func MakeJob(js *v1alpha1.JobSource) *batchv1.Job {
	spec := js.Spec.JobSpec.DeepCopy()
	podTemplate := &spec.Template
	podTemplate.Labels = reconciler.Labels(js, labelKey)
	podTemplate.Annotations = reconciler.Annotations(js)

	containers := []corev1.Container{}
	for i, c := range podTemplate.Spec.Containers {
		if c.Name == "" {
			c.Name = fmt.Sprintf("jobsource%d", i)
		}
		c.Env = append(c.Env, corev1.EnvVar{Name: "K_SINK", Value: js.Status.SinkURI})
		c.Env = append(c.Env, corev1.EnvVar{Name: "K_OUTPUT_FORMAT", Value: string(js.Spec.OutputFormat)})
		containers = append(containers, c)
	}
	podTemplate.Spec.Containers = containers

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            JobName(js.GetObjectMeta()),
			Namespace:       js.GetObjectMeta().GetNamespace(),
			Labels:          reconciler.Labels(js, labelKey),
			Annotations:     reconciler.Annotations(js),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(js)},
		},
		Spec: *spec,
	}

	// Extend annotations
	for k, v := range js.GetAnnotations() {
		job.Annotations[k] = v
	}

	return job
}

func JobName(owner metav1.Object) string {
	return utils.GenerateFixedName(owner, owner.GetName()+"-jobsource-")
}
