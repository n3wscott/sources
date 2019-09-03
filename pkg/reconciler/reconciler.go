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
	"errors"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	clientset "github.com/n3wscott/sources/pkg/client/clientset/versioned"
	sourcesclient "github.com/n3wscott/sources/pkg/client/injection/client"
	eventingreconciler "knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/resolver"
)

var (
	ErrSinkMissing = errors.New("Sink missing from spec")
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
	SinkResolver *resolver.URIResolver
}

func NewBase(ctx context.Context, controllerAgentName string, cmw configmap.Watcher) *Base {
	base := &Base{Base: eventingreconciler.NewBase(ctx, controllerAgentName, cmw)}

	base.SourcesClientSet = sourcesclient.Get(ctx)

	// TODO(spencer-p) This callback is a nop and should be changed.
	// impl := controller.NewImpl(r, logger, controllerAgentName)
	// r.sinkReconciler = resolver.NewURIResolver(ctx, impl.EnqueueKey)
	base.SinkResolver = resolver.NewURIResolver(ctx, func(_ string) {})

	return base
}

func (r *Base) ReconcileSink(ctx context.Context, source v1alpha1.Source) error {
	dest := source.GetSink()

	if dest.ObjectReference == nil && dest.URI == nil {
		source.GetStatus().MarkNoSink("Missing", "Sink missing from spec")
		return ErrSinkMissing
	}

	// If using the ObjectReference w/o a namespace, mirror the source's ns
	if dest.ObjectReference != nil && dest.ObjectReference.Namespace == "" {
		dest.ObjectReference.Namespace = source.GetNamespace()
	}

	uri, err := r.SinkResolver.URIFromDestination(dest, source)
	if err != nil {
		source.GetStatus().MarkNoSink("NotFound", "Could not resolve sink URI: %v", err)
		return err
	}

	source.GetStatus().MarkSink(uri)

	return nil
}
