package jobsource

import (
	"context"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/jobsource/resources"

	corev1 "k8s.io/api/core/v1"
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

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &Reconciler{
			Base:   reconciler.NewBase(ctx, "JobSource", cmw),
			Lister: listers.GetJobSourceLister(),
		}
	}))
}
