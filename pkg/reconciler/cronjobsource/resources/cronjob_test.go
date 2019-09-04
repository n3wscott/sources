package resources

import (
	"context"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
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
			Labels:          Labels(in),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(in)},
		},
		Spec: batchv1beta1.CronJobSpec{
			SuccessfulJobsHistoryLimit: ptr.Int32(100),
			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: Labels(in),
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
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
