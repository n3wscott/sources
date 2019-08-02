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
package resources

import (
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
	"knative.dev/pkg/kmeta"
)

const (
	labelKey = "sources.knative.dev/jobsource"
)

// TODO(spencer-p) This arguments object is heavy; almost everything is already in the JobSource
type Arguments struct {
	Owner        kmeta.OwnerRefable
	Namespace    string
	Spec         *batchv1.JobSpec
	Annotations  map[string]string
	Labels       map[string]string
	SinkURI      string
	OutputFormat v1alpha1.OutputFormatType
}

func Labels(owner kmeta.OwnerRefable) map[string]string {
	return map[string]string{
		labelKey: owner.GetObjectMeta().GetName(),
	}
}
