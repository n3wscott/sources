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

package reconciler

import (
	"context"

	"github.com/knative/eventing/pkg/duck"
	eventingreconciler "github.com/knative/eventing/pkg/reconciler"
	clientset "github.com/n3wscott/sources/pkg/client/clientset/versioned"
	sourcesclient "github.com/n3wscott/sources/pkg/client/injection/client"
	"knative.dev/pkg/configmap"
)

// Base is a set of tools that source reconcilers need.
type Base struct {
	// Include the knative/eventing reconciler base.
	// +required
	*eventingreconciler.Base

	// ClientSet for Sources; note that EventingClientSet (from eventing Base) is unused
	// +required
	SourcesClientSet clientset.Interface

	// Used by all Sources to resolve their sink.
	// +required
	SinkReconciler *duck.SinkReconciler
}

func NewBase(ctx context.Context, controllerAgentName string, cmw configmap.Watcher) *Base {
	base := &Base{Base: eventingreconciler.NewBase(ctx, controllerAgentName, cmw)}

	base.SourcesClientSet = sourcesclient.Get(ctx)

	// TODO(spencer-p) This callback is a joke.
	// impl := controller.NewImpl(r, logger, controllerAgentName)
	// r.sinkReconciler = duck.NewSinkReconciler(ctx, impl.EnqueueKey)
	base.SinkReconciler = duck.NewSinkReconciler(ctx, func(_ string) {})

	return base
}
