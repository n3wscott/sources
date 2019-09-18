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
	"context"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"knative.dev/pkg/ptr"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	"github.com/google/go-cmp/cmp"
)

func TestMakeCronJob(t *testing.T) {
	in := &v1alpha1.CronJobSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Steve",
			Namespace: "default",
		},
		Spec: v1alpha1.CronJobSourceSpec{
			CronJobSpec: batchv1beta1.CronJobSpec{
				SuccessfulJobsHistoryLimit: ptr.Int32(100),
				JobTemplate: batchv1beta1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									corev1.Container{
										Name:  "SteveImage",
										Image: "example-img",
									},
								},
							},
						},
					},
				},
			},
		},
		Status: v1alpha1.CronJobSourceStatus{
			BaseSourceStatus: v1alpha1.BaseSourceStatus{
				SinkURI: "http://example.com/",
			},
		},
	}

	in.SetDefaults(context.TODO())

	want := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "Steve",
			Namespace:       "default",
			Labels:          reconciler.Labels(in, labelKey),
			Annotations:     reconciler.Annotations(in),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(in)},
		},
		Spec: batchv1beta1.CronJobSpec{
			SuccessfulJobsHistoryLimit: ptr.Int32(100),
			Suspend:                    ptr.Bool(false),
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      reconciler.Labels(in, labelKey),
							Annotations: reconciler.Annotations(in),
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Name:  "SteveImage",
									Image: "example-img",
									Env: []corev1.EnvVar{
										corev1.EnvVar{Name: "K_SINK", Value: in.Status.SinkURI},
										corev1.EnvVar{Name: "K_OUTPUT_FORMAT", Value: string(in.Spec.OutputFormat)},
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	got := MakeCronJob(in)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want, +got): %s", diff)
	}
}
