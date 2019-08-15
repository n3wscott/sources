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

package v1alpha1

import (
	"testing"
)

func TestServiceSourceReady(t *testing.T) {
	tests := []struct {
		name string
		body func(s *ServiceSourceStatus)
		want bool
	}{{
		name: "initialize",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
		},
		want: false,
	}, {
		name: "has sink",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
		},
		want: false,
	}, {
		name: "sink not set but has service ready",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkServiceReady() // should not be possible
		},
		want: false,
	}, {
		name: "has sink but service not ready",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkServiceNotReady("MarkServiceNotReady", "")
		},
		want: false,
	}, {
		name: "has sink and service ready",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkServiceReady()
		},
		want: true,
	}, {
		name: "has sink but service ready unkown",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkServiceReadyUnknown("MarkServiceReadyUnknown", "")
		},
		want: false,
	}, {
		name: "no sink and service ready unknown",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkNoSink("MarkNoSink", "")
			s.MarkServiceReadyUnknown("Unknown", "")
		},
		want: false,
	}, {
		name: "no sink and service not ready",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkNoSink("MarkNoSink", "")
			s.MarkServiceNotReady("NotReady", "") // Does not make sense for this to happen
		},
		want: false,
	}, {
		name: "no sink and service ready",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkNoSink("MarkNoSink", "")
			s.MarkServiceReady()
		},
		want: false,
	}, {
		name: "take sink away",
		body: func(s *ServiceSourceStatus) {
			s.InitializeConditions()
			s.MarkSink("example.com")
			s.MarkServiceReady()
			s.MarkNoSink("MarkNoSink", "")
		},
		want: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &ServiceSourceStatus{}
			test.body(s)
			if got := s.Ready(); got != test.want {
				t.Errorf("Got %t, wanted %t", got, test.want)
			}
		})
	}
}
