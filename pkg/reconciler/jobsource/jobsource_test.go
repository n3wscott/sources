package jobsource

import (
	"context"
	"testing"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	. "github.com/n3wscott/sources/pkg/reconciler/testing"
	. "knative.dev/pkg/reconciler/testing"
)

const (
	jsName = "my-jobsource"
)

func init() {
	// Add types to scheme
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

func TestJobSource(t *testing.T) {
	table := TableTest{{
		Name: "create jobsource",
		Objects: []runtime.Object{
			NewJobSource(jsName, func(js *v1alpha1.JobSource) {
				js.Status.InitializeConditions()
			}),
		},
	}}

	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		return &Reconciler{
			//KubeClientSet:  asdf,
			//Client:         asdf,
			//Lister:         asdf,
			//Tracker:        asdf,
			//Recorder:       asdf,
			//sinkReconciler: asdf,
		}
	}))
}
