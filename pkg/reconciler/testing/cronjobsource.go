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
	"github.com/n3wscott/sources/pkg/reconciler/cronjobsource/resources"
	"knative.dev/pkg/ptr"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobSourceOption func(*v1alpha1.CronJobSource)

func NewCronJobSource(name string, options ...CronJobSourceOption) *v1alpha1.CronJobSource {
	s := &v1alpha1.CronJobSource{
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

func WithFakeCronJobSpec(s *v1alpha1.CronJobSource) {
	if s.Spec.CronJobSpec.SuccessfulJobsHistoryLimit == nil {
		s.Spec.CronJobSpec.SuccessfulJobsHistoryLimit = ptr.Int32(100)
	}

	s.Spec.CronJobSpec.JobTemplate.Spec.Template.Spec.Containers = append(s.Spec.CronJobSpec.JobTemplate.Spec.Template.Spec.Containers, corev1.Container{
		Name:  "Steve",
		Image: "grc.io/fakeimage",
	})
}

type CronJobOption func(*batchv1beta1.CronJob)

func NewCronJob(s *v1alpha1.CronJobSource, options ...CronJobOption) *batchv1beta1.CronJob {
	cronjob := resources.MakeCronJob(s)

	for _, option := range options {
		option(cronjob)
	}

	return cronjob
}
