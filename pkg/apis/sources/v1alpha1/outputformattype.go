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

	"knative.dev/pkg/apis"
)

// OutputFormatType describes the data format that a Source is expected to output when it sends Cloud Events.
type OutputFormatType string

const (
	// Sources are required to support structured and binary output formats.
	// Any other format type is invalid.
	OutputFormatStructured OutputFormatType = "structured"
	OutputFormatBinary                      = "binary"
)

// Check that OutputFormatType is Validatable
var _ apis.Validatable = OutputFormatType("")

// Validate ensures that the OutputFormatType is one of the two allowed types.
// It assumes that its field is "outputFormat".
func (o OutputFormatType) Validate(ctx context.Context) *apis.FieldError {
	switch o {
	case OutputFormatStructured, OutputFormatBinary:
		return nil
	case "":
		return apis.ErrMissingField("outputFormat")
	default:
		// Not supported.
		return apis.ErrInvalidValue(o, "outputFormat")
	}
}
