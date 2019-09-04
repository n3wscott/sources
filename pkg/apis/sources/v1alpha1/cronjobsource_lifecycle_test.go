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

func TestCronJobSourceReady(t *testing.T) {
	tests := []struct {
		name string
		body func(s *CronJobSourceStatus)
		want bool
	}{{
		name: "initialized",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
		},
		want: false,
	}, {
		name: "mark cron job created",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			s.MarkCronJobCreated()
		},
		want: false,
	}, {
		name: "mark job failed",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			s.MarkNoCronJob("")
		},
		want: false,
	}, {
		name: "mark sink",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
		},
		want: false,
	}, {
		name: "mark sink and cron job created",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkCronJobCreated()
		},
		want: true,
	}, {
		name: "mark sink and no cron job",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkNoCronJob("")
		},
		want: false,
	}, {
		name: "mark sink, no cron job, cron job created",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkNoCronJob("")
			s.MarkCronJobCreated()
		},
		want: true,
	}, {
		name: "mark created, then sink",
		body: func(s *CronJobSourceStatus) {
			s.InitializeConditions()
			// This exchange should not be possible. CronJob should not start running until there is a sink.
			s.MarkCronJobCreated()
			s.MarkSink("example.com")
		},
		want: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &CronJobSourceStatus{}
			test.body(s)
			if got := s.Ready(); got != test.want {
				t.Errorf("CronJobSourceStatus %s: from Succeeded() got %t, wanted %t", test.name, got, test.want)
			}
		})
	}
}
