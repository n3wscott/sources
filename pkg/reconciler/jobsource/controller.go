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

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	jsinformer "github.com/n3wscott/sources/pkg/client/injection/informers/sources/v1alpha1/jobsource"
	"github.com/n3wscott/sources/pkg/reconciler"
	jobinformer "knative.dev/pkg/injection/informers/kubeinformers/batchv1/job"

	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

const (
	controllerAgentName = "jobsource-controller"
)

// NewController returns a new HPA reconcile controller.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	jsInformer := jsinformer.Get(ctx)
	jobInformer := jobinformer.Get(ctx)

	r := &Reconciler{
		Base:   reconciler.NewBase(ctx, "JobSource", cmw),
		Lister: jsInformer.Lister(),
	}
	impl := controller.NewImpl(r, r.Logger, "JobSources")

	r.Logger.Info("Setting up event handlers for JobSources")

	jsInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	jobInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("JobSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
