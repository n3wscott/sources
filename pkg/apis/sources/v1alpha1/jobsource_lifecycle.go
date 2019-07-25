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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

var condSet = apis.NewLivingConditionSet()

// GetGroupVersionKind implements kmeta.OwnerRefable
func (js *JobSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("JobSource")
}

func (jss *JobSourceStatus) InitializeConditions() {
	condSet.Manage(jss).InitializeConditions()
}

// TODO(spencer-p) We might not need ConditionReady
func (jss *JobSourceStatus) MarkServiceUnavailable(name string) {
	condSet.Manage(jss).MarkFalse(
		JobSourceConditionReady,
		"ServiceUnavailable",
		"Service %q wasn't found.", name)
}

func (jss *JobSourceStatus) MarkServiceAvailable() {
	condSet.Manage(jss).MarkTrue(JobSourceConditionReady)
}
