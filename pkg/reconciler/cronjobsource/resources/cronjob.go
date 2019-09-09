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
	"knative.dev/pkg/kmeta"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	labelKey = "sources.knative.dev/jobsource"
)

func MakeCronJob(s *v1alpha1.CronJobSource) *batchv1beta1.CronJob {
	cronjob := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            CronJobName(s.GetObjectMeta()),
			Namespace:       s.GetObjectMeta().GetNamespace(),
			Labels:          Labels(s),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(s)},
		},
	}

	// Copy the Source's spec into the new CronJob object, then make changes
	s.Spec.CronJobSpec.DeepCopyInto(&cronjob.Spec)
	podTemplate := &cronjob.Spec.JobTemplate

	if podTemplate.ObjectMeta.Labels == nil {
		podTemplate.ObjectMeta.Labels = make(map[string]string)
	}
	podTemplate.ObjectMeta.Labels[labelKey] = s.GetObjectMeta().GetName()

	if podTemplate.ObjectMeta.Annotations == nil {
		podTemplate.ObjectMeta.Annotations = make(map[string]string)
	}
	for k, v := range s.GetAnnotations() {
		podTemplate.ObjectMeta.Annotations[k] = v
	}

	// TODO(spencer-p) Eliminate extra copying here
	containers := podTemplate.Spec.Template.Spec.Containers
	for i, _ := range containers {
		if containers[i].Name == "" {
			containers[i].Name = fmt.Sprintf("cronjobsource%d", i)
		}
		containers[i].Env = append(containers[i].Env, corev1.EnvVar{Name: "K_SINK", Value: s.Status.SinkURI})
		containers[i].Env = append(containers[i].Env, corev1.EnvVar{Name: "K_OUTPUT_FORMAT", Value: string(s.Spec.OutputFormat)})
	}

	// TODO(spencer-p) Set .Annotations or .Labels?
	return cronjob
}

func CronJobName(owner metav1.Object) string {
	// Reuse the owner's name.
	return owner.GetName()
}

func Labels(owner kmeta.OwnerRefable) map[string]string {
	return map[string]string{
		labelKey: owner.GetObjectMeta().GetName(),
	}
}
