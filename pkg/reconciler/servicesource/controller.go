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

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	servicesourceinformer "github.com/n3wscott/sources/pkg/client/injection/informers/sources/v1alpha1/servicesource"
	"github.com/n3wscott/sources/pkg/reconciler"

	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	serviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1alpha1/service"
)

const (
	controllerAgentName = "servicesource-controller"
)

// NewController returns a new HPA reconcile controller.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	serviceSourceInformer := servicesourceinformer.Get(ctx)
	svcInformer := serviceinformer.Get(ctx)

	r := &Reconciler{
		Base:             reconciler.NewBase(ctx, "ServiceSource", cmw),
		Lister:           serviceSourceInformer.Lister(),
		ServingClientSet: servingclient.Get(ctx),
	}
	impl := controller.NewImpl(r, r.Logger, "ServiceSources")

	r.Logger.Info("Setting up event handlers for ServiceSources")

	serviceSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	svcInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("ServiceSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
