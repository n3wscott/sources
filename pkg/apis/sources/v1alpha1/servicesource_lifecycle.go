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
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	// ServiceSourceConditionServiceReady is set to true when the underlying Knative service is ready
	// to receive traffic.
	ServiceSourceConditionServiceReady apis.ConditionType = "ServiceReady"

	// ServiceSourceConditionReady is the happy condition for a service source, true if there is
	// a ready service with a properly configured sink.
	ServiceSourceConditionReady = apis.ConditionReady

	serviceDeployingReason  = "Deploying"
	serviceDeployingMessage = "Service created; awaiting readiness"
)

var serviceSourceCondSet = apis.NewLivingConditionSet(
	SourceConditionSinkProvided,
	ServiceSourceConditionServiceReady,
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (js *ServiceSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ServiceSource")
}

func (s *ServiceSourceStatus) InitializeConditions() {
	serviceSourceCondSet.Manage(s).InitializeConditions()
}

func (s *ServiceSourceStatus) Ready() bool {
	return serviceSourceCondSet.Manage(s).IsHappy()
}

// MarkSink sets the conditions that the source has received a sink URI.
func (s *ServiceSourceStatus) MarkSink(uri string) {
	s.BaseSourceStatus.MarkSink(serviceSourceCondSet.Manage(s), uri)
}

func (s *ServiceSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	s.BaseSourceStatus.MarkNoSink(serviceSourceCondSet.Manage(s), reason, messageFormat, messageA...)
}

func (s *ServiceSourceStatus) MarkServiceReady() {
	serviceSourceCondSet.Manage(s).MarkTrue(ServiceSourceConditionServiceReady)
}

func (s *ServiceSourceStatus) MarkServiceReadyUnknown(reason string, messageFormat string, messageA ...interface{}) {
	serviceSourceCondSet.Manage(s).MarkUnknown(ServiceSourceConditionServiceReady, reason, messageFormat, messageA...)
}

func (s *ServiceSourceStatus) MarkServiceDeploying() {
	serviceSourceCondSet.Manage(s).MarkUnknown(ServiceSourceConditionServiceReady, serviceDeployingReason, serviceDeployingMessage)
}

func (s *ServiceSourceStatus) MarkServiceNotReady(reason, messageFormat string, messageA ...interface{}) {
	serviceSourceCondSet.Manage(s).MarkFalse(ServiceSourceConditionServiceReady, reason, messageFormat, messageA...)
}

func (s *ServiceSourceStatus) MarkAddress(addr *duckv1beta1.Addressable) {
	s.AddressStatus.Address = addr.DeepCopy()
}

func (s *ServiceSourceStatus) MarkNoAddress() {
	s.AddressStatus.Address = nil
}

func (s *ServiceSourceStatus) MarkURL(url *apis.URL) {
	s.URL = url.DeepCopy()
}

func (s *ServiceSourceStatus) MarkNoURL() {
	s.URL = nil
}
