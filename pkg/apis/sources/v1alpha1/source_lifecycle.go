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
	"knative.dev/pkg/apis"
)

// MarkSink sets the conditions that the source has received a sink URI.
func (s *BaseSourceStatus) MarkSink(mgr apis.ConditionManager, uri string) {
	s.SinkURI = uri
	if len(uri) > 0 {
		mgr.MarkTrue(SourceConditionSinkProvided)
	} else {
		mgr.MarkUnknown(SourceConditionSinkProvided, "SinkEmpty", "Sink resolved to empty URI")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
// TODO(spencer-p) This method adds almost nothing -- would be nice to have MarkSinkInvalid, MarkSinkNotResolved, etc
func (s *BaseSourceStatus) MarkNoSink(mgr apis.ConditionManager, reason, messageFormat string, messageA ...interface{}) {
	mgr.MarkFalse(SourceConditionSinkProvided, reason, messageFormat, messageA...)
}
