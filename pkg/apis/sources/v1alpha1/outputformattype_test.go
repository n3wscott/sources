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
	"testing"
)

func TestOutputFormatTypeValid(t *testing.T) {
	tests := []struct {
		o    OutputFormatType
		want bool
	}{
		{"structured", true},
		{"binary", true},
		{"trinary", false},
		{"quantum", false},
		{"http", false}, // all events are over HTTP
		{"messenger_pigeon", false},
	}

	for _, test := range tests {
		t.Run(string(test.o), func(t *testing.T) {
			if got := test.o.Validate(context.Background()) == nil; got != test.want {
				t.Errorf("OutputFormatType %s got %t for Valid(), wanted %t", test.o, got, test.want)
			}
		})
	}
}
