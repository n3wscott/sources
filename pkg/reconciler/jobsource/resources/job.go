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
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/ptr"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeJob(args Arguments) *batchv1.Job {
	podTemplate := args.Template
	if podTemplate.ObjectMeta.Labels == nil {
		podTemplate.ObjectMeta.Labels = make(map[string]string)
	}
	podTemplate.ObjectMeta.Labels[labelKey] = args.Owner.GetObjectMeta().GetName()
	podTemplate.Spec.RestartPolicy = corev1.RestartPolicyNever

	containers := []corev1.Container{}
	for i, c := range podTemplate.Spec.Containers {
		if c.Name == "" {
			c.Name = fmt.Sprintf("jobsource%d", i)
		}
		c.Env = append(c.Env, corev1.EnvVar{Name: "K_SINK", Value: args.SinkURI})
		c.Env = append(c.Env, corev1.EnvVar{Name: "K_OUTPUT_FORMAT", Value: string(args.OutputFormat)})
		containers = append(containers, c)
	}
	podTemplate.Spec.Containers = containers

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    jobName(args.Owner),
			Namespace:       args.Owner.GetObjectMeta().GetNamespace(),
			Labels:          Labels(args.Owner),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(args.Owner)},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.Int32(0),
			Parallelism:  ptr.Int32(1),
			Completions:  ptr.Int32(1),
			Template:     *podTemplate,
		},
	}

	if args.Annotations != nil {
		if job.Spec.Template.ObjectMeta.Annotations == nil {
			job.Spec.Template.ObjectMeta.Annotations = make(map[string]string, len(args.Annotations))
		}
		for k, v := range args.Annotations {
			job.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}

	if args.Labels != nil {
		for k, v := range args.Labels {
			if k != labelKey {
				job.Spec.Template.ObjectMeta.Labels[k] = v
			}
		}
	}
	return job
}

func jobName(owner kmeta.OwnerRefable) string {
	return strings.ToLower(
		strings.Join(append([]string{owner.GetObjectMeta().GetName(), "jobsource"}), "-") + "-")
}
