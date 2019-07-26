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
	"testing"
)

func TestJobSourceSucceeded(t *testing.T) {
	tests := []struct {
		name string
		body func(s *JobSourceStatus)
		want bool
	}{
		{
			name: "initialized",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
			},
			want: false,
		},
		{
			name: "mark job succeeded",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkJobSucceeded()
			},
			want: false,
		},
		{
			name: "mark job failed",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkJobFailed("MarkJobFailed", "")
			},
			want: false,
		},
		{
			name: "mark sink",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkSink("example.com")
			},
			want: false,
		},
		{
			name: "mark sink and succeed",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkSink("example.com")
				s.MarkJobSucceeded()
			},
			want: true,
		},
		{
			name: "mark sink and running",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkSink("example.com")
				s.MarkJobRunning("")
			},
			want: false,
		},
		{
			name: "mark sink, running and mark sink again",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkSink("example.com")
				s.MarkJobRunning("")
				s.MarkSink("example2.com")
			},
			want: false,
		},
		{
			name: "mark sink, running and job succeeded",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkSink("example.com")
				s.MarkJobRunning("")
				s.MarkJobSucceeded()
			},
			want: true,
		},
		{
			name: "mark sink, running and job failed",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				s.MarkSink("example.com")
				s.MarkJobRunning("")
				s.MarkJobFailed("MarkJobFailed", "")
			},
			want: false,
		},
		{
			name: "mark running, sink, job succeeded",
			body: func(s *JobSourceStatus) {
				s.InitializeConditions()
				// This exchange should not be possible. Job should not start running until there is a sink.
				s.MarkJobRunning("")
				s.MarkSink("example.com")
				s.MarkJobSucceeded()
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &JobSourceStatus{}
			test.body(s)
			if got := s.Succeeded(); got != test.want {
				t.Errorf("JobSourceStatus %s: from Succeeded() got %t, wanted %t", test.name, got, test.want)
			}
		})
	}
}
