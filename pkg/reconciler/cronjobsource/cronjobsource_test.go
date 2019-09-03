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

package cronjobsource

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/cronjobsource/resources"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/ptr"

	. "github.com/n3wscott/sources/pkg/reconciler/testing"
	. "knative.dev/pkg/reconciler/testing"
)

const (
	sName         = "my-cronjobsource"
	sUID          = "1234"
	sJobFixedName = sName + "-cronjobsource-" + sUID
	sinkName      = "my-sink"
	ns            = "default"
	key           = ns + "/" + sName
	sinkURI       = "http://" + sinkName + "." + ns + ".svc.cluster.local/"

	failreason  = "fail reason"
	failmessage = "fail message"
)

var (
	svcSink = destMust(apisv1alpha1.NewDestination(&corev1.ObjectReference{
		Name:       sinkName,
		Namespace:  ns,
		APIVersion: "v1",
		Kind:       "Service",
	}))

	lastScheduleTime = metav1.Time{time.Time{}.Add(time.Duration(1337))}
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

func TestCronJobSource(t *testing.T) {
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
			NewCronJobSource(sName, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
			}),
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewCronJobSource(sName, func(s *v1alpha1.CronJobSource) {
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
			NewCronJobSource(sName, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = namedTestSink("dne") // does not exist
			}),
			// Sink not added to objects
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewCronJobSource(sName, func(s *v1alpha1.CronJobSource) {
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
			NewCronJobSource(sName, func(s *v1alpha1.CronJobSource) {
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
			Object: NewCronJobSource(sName, func(s *v1alpha1.CronJobSource) {
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
		Name: "having sink creates a cronjob",
		Objects: []runtime.Object{
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = svcSink
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Spec.Sink = svcSink

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkCronJobCreated()
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeCronJob(
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Spec.Sink = svcSink
				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
			}),
		)},
	}, {
		Name: "having uri sink creates a cronjob",
		Objects: []runtime.Object{
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}}
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}}

				s.Status.InitializeConditions()
				s.Status.MarkSink("http://example.com")
				s.Status.MarkCronJobCreated()
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeCronJob(
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Status.MarkSink("http://example.com")
			}),
		)},
	}, {
		Name: "having uri sink with path starts a job",
		Objects: []runtime.Object{
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}, Path: ptr.String("/foo/bar")}
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}, Path: ptr.String("/foo/bar")}

				s.Status.InitializeConditions()
				s.Status.MarkSink("http://example.com/foo/bar")
				s.Status.MarkCronJobCreated()
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeCronJob(
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Status.MarkSink("http://example.com/foo/bar")
			}),
		)},
	}, {
		Name: "cronjobsource mirrors cronjob status",
		Objects: []runtime.Object{
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = svcSink
			}),
			NewCronJob(
				NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
					s.UID = sUID
					s.Spec.Sink = svcSink
					s.Status.InitializeConditions()
					s.Status.MarkSink(sinkURI)
				}),
				func(cronjob *batchv1beta1.CronJob) {
					// Set an arbitrary last schedule time
					cronjob.Status.LastScheduleTime = &lastScheduleTime
				},
			),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Spec.Sink = svcSink

				s.Status.InitializeConditions()
				s.Status.MarkSink(sinkURI)
				s.Status.MarkCronJobCreated()

				// Time in the CronJobSource should match what we set in the CronJob
				s.Status.LastScheduleTime = &lastScheduleTime
			}),
		}},
	}, {
		Name: "sink updates change the cronjob",
		Objects: []runtime.Object{
			NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Status.InitializeConditions()
				s.Spec.Sink = namedTestSink(sinkName)
				s.Status.MarkSink(sinkURI)
				s.Status.MarkCronJobCreated()
			}),
			NewCronJob(NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
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
			Object: NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
				s.UID = sUID
				s.Spec.Sink = namedTestSink(sinkName)
				s.Status.InitializeConditions()
				s.Status.MarkSink("http://garbage")
				s.Status.MarkCronJobCreated()
			}),
		}},
		WantUpdates: []clientgotesting.UpdateActionImpl{
			clientgotesting.UpdateActionImpl{
				Object: NewCronJob(NewCronJobSource(sName, WithFakeCronJobSpec, func(s *v1alpha1.CronJobSource) {
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
			Base:   reconciler.NewBase(ctx, "CronJobSource", cmw),
			Lister: listers.GetCronJobSourceLister(),
		}
	}))
}
