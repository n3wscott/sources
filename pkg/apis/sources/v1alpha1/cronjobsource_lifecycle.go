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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
)

const (
	noCronJobReason = "NoCronJob"

	CronJobSourceConditionReady = apis.ConditionReady

	// SinkProvided is inherited from the base status.

	// CronJobSourceConditionCronJobCreated becomes true when the underlying CronJob exists.
	CronJobSourceConditionCronJobCreated apis.ConditionType = "CronJobCreated"
)

var cronJobCondSet = apis.NewBatchConditionSet(
	SourceConditionSinkProvided,
	CronJobSourceConditionCronJobCreated,
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (js *CronJobSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("CronJobSource")
}

func (s *CronJobSourceStatus) InitializeConditions() {
	jobCondSet.Manage(s).InitializeConditions()
}

// Ready returns true if the CronJobSource has a sink and a CronJob.
func (s *CronJobSourceStatus) Ready() bool {
	return jobCondSet.Manage(s).IsHappy()
}

// MarkSink sets the conditions that the source has received a sink URI.
func (s *CronJobSourceStatus) MarkSink(uri string) {
	s.BaseSourceStatus.MarkSink(jobCondSet.Manage(s), uri)
}

func (s *CronJobSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	s.BaseSourceStatus.MarkNoSink(jobCondSet.Manage(s), reason, messageFormat, messageA...)
}

// MarkCronJobCreated sets the condition that the CronJobSource owns a CronJob.
func (s *CronJobSourceStatus) MarkCronJobCreated() {
	cronJobCondSet.Manage(s).MarkTrue(CronJobSourceConditionCronJobCreated)
}

// MarkCronJobCreated sets the condition that the CronJobSource does not yet have a CronJob.
func (s *CronJobSourceStatus) MarkNoCronJob(msgFmt string, messageA ...interface{}) {
	cronJobCondSet.Manage(s).MarkFalse(CronJobSourceConditionCronJobCreated, noCronJobReason, msgFmt, messageA...)
}

func (s *CronJobSourceStatus) PropagateCronJobStatus(from *batchv1beta1.CronJobStatus) {
	from.DeepCopyInto(&s.CronJobStatus)
}
