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

package jobsource

import (
	"context"
	"fmt"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/jobsource/resources"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	jsName         = "my-jobsource"
	jsUID          = "1234"
	jsJobFixedName = jsName + "-jobsource-" + jsUID
	sinkName       = "my-sink"
	ns             = "default"
	key            = ns + "/" + jsName
	sinkURI        = "http://" + sinkName + "." + ns + ".svc.cluster.local/"

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
)

func namedTestSink(name string) apisv1alpha1.Destination {
	return destMust(apisv1alpha1.NewDestination(&corev1.ObjectReference{
		Name:       name,
		Namespace:  ns,
		APIVersion: "testing.eventing.knative.dev/v1alpha1",
		Kind:       "Sink",
	}))
}

func init() {
	// Add types to scheme
	_ = v1alpha1.AddToScheme(scheme.Scheme)
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

func TestJobSource(t *testing.T) {
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
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
			}),
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID

				js.Status.InitializeConditions()
				js.Status.MarkNoSink("Missing", "Sink missing from spec")
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "UpdateFailed", "Failed to update status for %q: expected exactly one, got neither: spec.sink.uri, spec.sink[apiVersion, kind, name]", jsName),
		},
	}, {
		Name: "sink not existing causes errors",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = namedTestSink("dne") // does not exist
			}),
			// Sink not added to objects
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = namedTestSink("dne")

				js.Status.InitializeConditions()
				js.Status.MarkNoSink("NotFound", `Could not resolve sink URI: failed to get ref %+v: sinks.testing.eventing.knative.dev "dne" not found`, js.Spec.Sink)
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError", `failed to get ref %+v: sinks.testing.eventing.knative.dev "dne" not found`, namedTestSink("dne")),
		},
	}, {
		Name: "sink with bad address causes errors",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = namedTestSink(sinkName)
			}),
			// The sink we are referencing here has an empty host, which should be erroneous.
			// This will result in a lookup fail.
			newUnstructuredSink("messengerpigeon", ""),
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = namedTestSink(sinkName)

				js.Status.InitializeConditions()
				js.Status.MarkNoSink("NotFound", fmt.Sprintf(`Could not resolve sink URI: hostname missing in address of %+v`, namedTestSink(sinkName)))
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError", `hostname missing in address of %+v`, namedTestSink(sinkName)),
		},
	}, {
		Name: "having sink starts a job",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = svcSink
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink

				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeJob(
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
			}),
		)},
	}, {
		Name: "having uri sink starts a job",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}}
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}}

				js.Status.InitializeConditions()
				js.Status.MarkSink("http://example.com")
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeJob(
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
			}),
		)},
	}, {
		Name: "having uri sink with path starts a job",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}, Path: ptr.String("/foo/bar")}
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = apisv1alpha1.Destination{URI: &apis.URL{Host: "example.com", Scheme: "http"}, Path: ptr.String("/foo/bar")}

				js.Status.InitializeConditions()
				js.Status.MarkSink("http://example.com/foo/bar")
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeJob(
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
			}),
		)},
	}, {
		Name: "job succeed implies jobsource succeed",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
			NewJob(NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
			}), func(job *batchv1.Job) {
				// Mark the job as completed.  If this test
				// fails in the future, it may be because the
				// MakeJob function is inserting a success
				// condition with status false.
				job.Status.Conditions = append(job.Status.Conditions, batchv1.JobCondition{
					Type:   batchv1.JobComplete,
					Status: corev1.ConditionTrue,
				})
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink

				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobSucceeded()
			}),
		}},
	}, {
		Name: "job failed implies jobsource failed",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
			NewJob(NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
			}), func(job *batchv1.Job) {
				// Mark the job as failed.
				// See above test for potential issues.
				job.Status.Conditions = append(job.Status.Conditions, batchv1.JobCondition{
					Type:    batchv1.JobFailed,
					Status:  corev1.ConditionTrue,
					Reason:  failreason,
					Message: failmessage,
				})
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink

				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobFailed(failreason, failmessage)
			}),
		}},
	}, {
		Name: "job running means jobsource status unknown",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
			}),
			NewJob(NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink
				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
			}), func(job *batchv1.Job) {}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = svcSink

				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobRunning("Job %q already exists.", jsJobFixedName)
			}),
		}},
	}, {
		Name: "sink updates do not change anything after job starts",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = namedTestSink(sinkName)
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
			NewJob(NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = namedTestSink(sinkName)
				js.Status.MarkSink(sinkURI)
			}), func(job *batchv1.Job) {}),

			// This address is different than the one we marked
			newUnstructuredSink("http", "garbage"),
		},
		Key: key,
		// Expect nothing to happen
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &Reconciler{
			Base:   reconciler.NewBase(ctx, "JobSource", cmw),
			Lister: listers.GetJobSourceLister(),
		}
	}))
}
