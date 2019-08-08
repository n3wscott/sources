package jobsource

import (
	"context"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/jobsource/resources"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

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
)

func init() {
	// Add types to scheme
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

func newSink() *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Name:       sinkName,
		Namespace:  ns,
		APIVersion: "v1",
		Kind:       "Service",
	}
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
					"hostname": hostname,
					"scheme":   scheme,
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
		Name: "missing sink causes errors",
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
			Eventf(corev1.EventTypeWarning, "UpdateFailed", "Failed to update status for %q: missing field(s): spec.sink", jsName),
		},
	}, {
		Name: "sink not existing causes errors",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = &corev1.ObjectReference{
					Name:       "sink-that-does-not-exist",
					Namespace:  ns,
					APIVersion: "testing.eventing.knative.dev/v1alpha1",
					Kind:       "Sink",
				}
			}),
			// Sink not added to objects
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = &corev1.ObjectReference{
					Name:       "sink-that-does-not-exist",
					Namespace:  ns,
					APIVersion: "testing.eventing.knative.dev/v1alpha1",
					Kind:       "Sink",
				}

				js.Status.InitializeConditions()
				js.Status.MarkNoSink("NotFound", `Could not get sink URI from default/sink-that-does-not-exist: Error fetching sink &ObjectReference{Kind:Sink,Namespace:default,Name:sink-that-does-not-exist,UID:,APIVersion:testing.eventing.knative.dev/v1alpha1,ResourceVersion:,FieldPath:,} for source "default/my-jobsource, /, Kind=": sinks.testing.eventing.knative.dev "sink-that-does-not-exist" not found`)
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError", `Error fetching sink &ObjectReference{Kind:Sink,Namespace:default,Name:sink-that-does-not-exist,UID:,APIVersion:testing.eventing.knative.dev/v1alpha1,ResourceVersion:,FieldPath:,} for source "default/my-jobsource, /, Kind=": sinks.testing.eventing.knative.dev "sink-that-does-not-exist" not found`),
		},
	}, {
		Name: "sink with bad address causes errors",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = &corev1.ObjectReference{
					Name:       sinkName,
					Namespace:  ns,
					APIVersion: "testing.eventing.knative.dev/v1alpha1",
					Kind:       "Sink",
				}
			}),
			// The sink we are referencing here has an empty host, which should be erroneous.
			// This will actually result in a lookup fail.
			newUnstructuredSink("messenger_pigeon", ""),
		},
		Key:     key,
		WantErr: true,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = &corev1.ObjectReference{
					Name:       sinkName,
					Namespace:  ns,
					APIVersion: "testing.eventing.knative.dev/v1alpha1",
					Kind:       "Sink",
				}

				js.Status.InitializeConditions()
				js.Status.MarkNoSink("NotFound", `Could not get sink URI from default/my-sink: sink &ObjectReference{Kind:Sink,Namespace:default,Name:my-sink,UID:,APIVersion:testing.eventing.knative.dev/v1alpha1,ResourceVersion:,FieldPath:,} contains an empty hostname`)
			}),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError", `sink &ObjectReference{Kind:Sink,Namespace:default,Name:my-sink,UID:,APIVersion:testing.eventing.knative.dev/v1alpha1,ResourceVersion:,FieldPath:,} contains an empty hostname`),
		},
	}, {
		Name: "having sink starts a job",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Status.InitializeConditions()
				js.Spec.Sink = newSink()
			}),
		},
		Key: key,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = newSink()

				js.Status.InitializeConditions()
				js.Status.MarkSink(sinkURI)
				js.Status.MarkJobRunning("Created Job %q.", jsJobFixedName)
			}),
		}},
		WantCreates: []runtime.Object{resources.MakeJob(
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.UID = jsUID
				js.Spec.Sink = newSink()
			}),
		)},
	}}

	// TODO(spencer-p)
	// make the job succeed and see if the jobsource succeeds
	// force failed job

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &Reconciler{
			Base:   reconciler.NewBase(ctx, "JobSource", cmw),
			Lister: listers.GetJobSourceLister(),
		}
	}))
}
