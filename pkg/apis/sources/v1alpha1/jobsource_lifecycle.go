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

const (
	jobRunningReason = "Running"

	// JobSourceConditionSucceeded is set when the revision starts to
	// materialize runtime resources and becomes true when the Job finishes
	// successfully.
	JobSourceConditionSucceeded = apis.ConditionSucceeded

	// SinkProvided is inherited from the base status.

	// JobSourceConditionJobSucceeded becomes true when the underlying Job
	// succeeds.
	JobSourceConditionJobSucceeded apis.ConditionType = "JobSucceeded"
)

var jobCondSet = apis.NewBatchConditionSet(
	SourceConditionSinkProvided,
	JobSourceConditionJobSucceeded,
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (js *JobSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("JobSource")
}

func (s *JobSourceStatus) InitializeConditions() {
	jobCondSet.Manage(s).InitializeConditions()
}

// Succeeded returns true if the JobSource has succeeded.
func (s *JobSourceStatus) Succeeded() bool {
	// This condition set should use apis.ConditionSucceeded (JobSourceConditionSucceeded).
	return jobCondSet.Manage(s).IsHappy()
}

// MarkSink sets the conditions that the source has received a sink URI.
func (s *JobSourceStatus) MarkSink(uri string) {
	if s.IsJobRunning() {
		// If the sink changes while the job is already running, we are choosing to let the job finish
		// using the outdated sink.
		// TODO(spencer-p) Log this somewhere.
		// TODO(spencer-p,n3wscott) This is opinionated, discuss further. Other options include:
		//  - Proxying the sink and resolving it only when the cloud event is generated
		//  - Killing the job and restarting it
		return
	}

	s.BaseSourceStatus.MarkSink(jobCondSet.Manage(s), uri)
}

func (s *JobSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	s.BaseSourceStatus.MarkNoSink(jobCondSet.Manage(s), reason, messageFormat, messageA...)
}

// JobSucceeded returns true if the underlying Job has succeeded.
func (s *JobSourceStatus) JobSucceeded() bool {
	return jobCondSet.Manage(s).GetCondition(JobSourceConditionJobSucceeded).IsTrue()
}

// IsJobRunning returns true if the job is currently running.
func (s *JobSourceStatus) IsJobRunning() bool {
	jobsucceeded := jobCondSet.Manage(s).GetCondition(JobSourceConditionJobSucceeded)
	if !jobsucceeded.IsUnknown() {
		// The job's success is known iff if it finished running.
		return false
	}

	// If the success is unknown because it is jobRunningReason, then the job is running.
	// TODO(spencer-p) Better way to do this?
	return jobsucceeded.Reason == jobRunningReason
}

// MarkJobSucceeded sets the condition that the underlying Job succeeded.
func (s *JobSourceStatus) MarkJobSucceeded() {
	jobCondSet.Manage(s).MarkTrue(JobSourceConditionJobSucceeded)
}

// MarkJobRunning sets the condition of the underlying Job's success to unknown.
func (s *JobSourceStatus) MarkJobRunning(messageFormat string, messageA ...interface{}) {
	jobCondSet.Manage(s).MarkUnknown(JobSourceConditionJobSucceeded, jobRunningReason, messageFormat, messageA...)
}

// MarkJobFailed sets the condition that the underlying Job failed.
func (s *JobSourceStatus) MarkJobFailed(reason, messageFormat string, messageA ...interface{}) {
	jobCondSet.Manage(s).MarkFalse(JobSourceConditionJobSucceeded, reason, messageFormat, messageA...)
}
