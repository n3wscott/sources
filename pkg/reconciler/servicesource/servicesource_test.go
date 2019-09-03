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

package servicesource

import (
	"context"
	"fmt"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/servicesource/resources"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	_ "knative.dev/pkg/ptr"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
	fakeservingclient "knative.dev/serving/pkg/client/injection/client/fake"

	. "github.com/n3wscott/sources/pkg/reconciler/testing"
	. "knative.dev/pkg/reconciler/testing"
)

const (
	sName             = "my-servicesource"
	sUID              = "1234"
	sServiceFixedName = sName + "-source-" + sUID
	sinkName          = "my-sink"
	ns                = "default"
	key               = ns + "/" + sName
	sinkURI           = "http://" + sinkName + "." + ns + ".svc.cluster.local/"

	notReadyReason  = "not ready reason"
	notReadyMessage = "not ready message"
)

var (
	refDest = destMust(apisv1alpha1.NewDestination(&corev1.ObjectReference{
		Name:       sinkName,
		Namespace:  ns,
		APIVersion: "v1",
		Kind:       "Service",
	}))

	uriDest = destMust(apisv1alpha1.NewDestinationURI(&apis.URL{
		Scheme: "http",
		Host:   sinkName + "." + ns + ".svc.cluster.local",
		Path:   "/",
	}))
)

func init() {
	// Add types to scheme
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

func namedTestSink(name string) apisv1alpha1.Destination {
	return destMust(apisv1alpha1.NewDestination(&corev1.ObjectReference{
		Name:       name,
		Namespace:  ns,
		APIVersion: "testing.eventing.knative.dev/v1alpha1",
		Kind:       "Sink",
	}))
}

// destMust eats errors related to destination creation, which should not happen for our known test inputs.
func destMust(dest *apisv1alpha1.Destination, err error) apisv1alpha1.Destination {
	if err != nil {
		panic(fmt.Errorf("destination construction should not error: %v", err))
	}
	return *dest
}

func newUnstructuredSink(scheme, hostname string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "testing.eventing.knative.dev/v1alpha1",
			"kind":       "Sink",
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      sinkName,
			},
			"status": map[string]interface{}{
				"address": map[string]interface{}{
					"url": scheme + "://" + hostname,
				},
			},
		},
	}
}

func TestServiceSource(t *testing.T) {
	table := TableTest{{
		Name: "bad workqueue key",
		// Make sure Reconcile handles bad keys.
		Key: "too/many/parts",
	}, {
		Name: "key not found",
		// Make sure Reconcile handles good keys that don't exist.
		Key: "foo/not-found",
	}, {
		Name: "missing sink in spec causes errors",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
			}),
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID

				s.Status.InitializeConditions()
				s.Status.MarkNoSink("Missing", "Sink missing from spec")
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "UpdateFailed", "Failed to update status for %q: expected exactly one, got neither: spec.sink.uri, spec.sink[apiVersion, kind, name]", sName),
		},
	}, {
		Name: "sink not existing causes errors",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = namedTestSink("dne") // does not exist
			}),
			// Sink not added to objects
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = namedTestSink("dne")

				s.Status.InitializeConditions()
				s.Status.MarkNoSink("NotFound", `Could not resolve sink URI: failed to get ref %+v: sinks.testing.eventing.knative.dev "dne" not found`, s.Spec.Sink)
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError", `failed to get ref %+v: sinks.testing.eventing.knative.dev "dne" not found`, namedTestSink("dne")),
		},
	}, {
		Name: "sink with bad address causes errors",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = namedTestSink(sinkName)
			}),
			// The sink we are referencing here has an empty host, which should be erroneous.
			// This will result in a lookup fail.
			newUnstructuredSink("messengerpigeon", ""),
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = namedTestSink(sinkName)

				s.Status.InitializeConditions()
				s.Status.MarkNoSink("NotFound", fmt.Sprintf(`Could not resolve sink URI: hostname missing in address of %+v`, namedTestSink(sinkName)))
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError", `hostname missing in address of %+v`, namedTestSink(sinkName)),
		},
	}, {
		Name: "having sink creates a service",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = refDest
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceDeploying()
			}),
		}},
		WantCreates: []runtime.Object{
			resources.MakeService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
			})),
		},
	}, {
		Name: "having sink uri creates a service",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = uriDest
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = uriDest

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceDeploying()
			}),
		}},
		WantCreates: []runtime.Object{
			resources.MakeService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = uriDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
			})),
		},
	}, {
		Name: "service ready propagates to servicesource",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = refDest
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceDeploying()
			}),
			NewService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
			}), func(svc *servingv1beta1.Service) {
				svc.Status.Conditions = append(svc.Status.Conditions, apis.Condition{
					Type:   servingv1beta1.ServiceConditionReady,
					Status: corev1.ConditionTrue,
				})
				svc.Status.Address = &duckv1beta1.Addressable{&apis.URL{
					Host:   fmt.Sprintf("%s.%s.cluster.local", svc.Name, svc.Namespace),
					Scheme: "http",
				}}
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceReady()
				s.Status.MarkAddress(&duckv1beta1.Addressable{&apis.URL{
					Host:   fmt.Sprintf("%s.%s.cluster.local", resources.ServiceName(s), ns),
					Scheme: "http",
				}})
			}),
		}},
	}, {
		Name: "service ready means servicesource ready",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceDeploying()
			}),
			NewService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
			}), func(service *servingv1beta1.Service) {
				// Force a ready condition
				service.Status.Conditions = append(service.Status.Conditions, apis.Condition{
					Type:   servingv1beta1.ServiceConditionReady,
					Status: corev1.ConditionTrue,
				})
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceReady()
			}),
		}},
	}, {
		Name: "service not ready means servicesource not ready",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceDeploying()
			}),
			NewService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
			}), func(service *servingv1beta1.Service) {
				// Force a not ready condition
				service.Status.Conditions = append(service.Status.Conditions, apis.Condition{
					Type:    servingv1beta1.ServiceConditionReady,
					Status:  corev1.ConditionFalse,
					Reason:  notReadyReason,
					Message: notReadyMessage,
				})
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = refDest

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceNotReady(notReadyReason, notReadyMessage)
			}),
		}},
	}, {
		Name: "sink updates restart service",
		Objects: []runtime.Object{
			NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = namedTestSink(sinkName)
				s.Status.MarkSink(sinkURI)
				s.Status.MarkServiceReady()
			}),
			NewService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = namedTestSink(sinkName)
				s.Status.MarkSink(sinkURI)
			})),

			// This address is different than the one we marked
			newUnstructuredSink("http", "garbage"),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
				s.UID = sUID
				s.Spec.Sink = namedTestSink(sinkName)

				s.Status.InitializeConditions()
				s.Status.MarkSink("http://garbage")
				s.Status.MarkServiceDeploying()
			}),
		}},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			clientgotesting.UpdateActionImpl{
				Object: NewService(NewServiceSource(sName, WithMinServiceSpec, func(s *v1alpha1.ServiceSource) {
					s.UID = sUID
					s.Status.InitializeConditions()
					s.Spec.Sink = namedTestSink(sinkName)
					s.Status.MarkSink("http://garbage")
				})),
			},
		},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &Reconciler{
			Base:             reconciler.NewBase(ctx, "ServiceSource", cmw),
			Lister:           listers.GetServiceSourceLister(),
			ServingClientSet: fakeservingclient.Get(ctx),
		}
	}))
}
