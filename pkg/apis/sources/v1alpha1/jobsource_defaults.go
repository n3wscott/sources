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

	"knative.dev/pkg/ptr"

	corev1 "k8s.io/api/core/v1"
)

// SetDefaults implements apis.Defaultable
func (js *JobSource) SetDefaults(ctx context.Context) {
	// TODO(spencer-p) Document this default
	if js.Spec.OutputFormat == "" {
		js.Spec.OutputFormat = OutputFormatBinary
	}

	// Use the documented default for the embedded JobSpec.
	// See k8s.io/api/batch/v1.JobSpec.BackoffLimit.
	if js.Spec.BackoffLimit == nil {
		js.Spec.BackoffLimit = ptr.Int32(6)
	}

	// Kubernetes defaults this to "Always", which is not valid for jobs.
	// Set something valid and sane for this job.
	if tSpec := &js.Spec.Template.Spec; tSpec.RestartPolicy == "" {
		if *js.Spec.BackoffLimit > 0 {
			tSpec.RestartPolicy = corev1.RestartPolicyOnFailure
		} else {
			tSpec.RestartPolicy = corev1.RestartPolicyNever
		}
	}
}
