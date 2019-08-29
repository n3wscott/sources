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
func (s *JobSource) SetDefaults(ctx context.Context) {
	s.Spec.BaseSourceSpec.SetDefaults(ctx)

	// Use the documented default for the embedded JobSpec.
	// See k8s.io/api/batch/v1.JobSpec.BackoffLimit.
	if s.Spec.BackoffLimit == nil {
		s.Spec.BackoffLimit = ptr.Int32(6)
	}

	// Kubernetes defaults the template spec RestartPolicy to "Always",
	// which is not valid for jobs.
	// Choose a default that makes sense given the BackoffLimit.
	if tSpec := &s.Spec.Template.Spec; tSpec.RestartPolicy == "" {
		if s.Spec.BackoffLimit == nil || *s.Spec.BackoffLimit <= 0 {
			// No back off limit; don't recreate jobs forever
			tSpec.RestartPolicy = corev1.RestartPolicyNever
		} else {
			tSpec.RestartPolicy = corev1.RestartPolicyOnFailure
		}
	}
}
