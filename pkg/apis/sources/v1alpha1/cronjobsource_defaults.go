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

	corev1 "k8s.io/api/core/v1"
)

// SetDefaults implements apis.Defaultable
func (s *CronJobSource) SetDefaults(ctx context.Context) {
	s.Spec.BaseSourceSpec.SetDefaults(ctx)

	// TODO(spencer-p) Is this a good default? Look at k8s docs
	restartPolicy := &s.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy
	if *restartPolicy == "" || *restartPolicy == corev1.RestartPolicyAlways {
		*restartPolicy = corev1.RestartPolicyNever
	}
}
