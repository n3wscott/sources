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
	"errors"
	"fmt"
	"reflect"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	listers "github.com/n3wscott/sources/pkg/client/listers/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/servicesource/resources"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/controller"
	_ "knative.dev/pkg/logging"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
	servingclientset "knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	serviceReadyUnknownReason  = "NoStatus"
	serviceReadyUnknownMessage = "Failed to look up status of Service"
)

var (
	ErrServiceReadyConditionMissing = errors.New("service does not have a ready condition")
)

// Reconciler implements controller.Reconciler for ServiceSource resources.
type Reconciler struct {
	// +required
	*reconciler.Base

	// Lister allows us to query for ServiceSources
	// +required
	Lister listers.ServiceSourceLister

	// ServingClientSet is for querying Knative Serving Services
	ServingClientSet servingclientset.Interface
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile implements controller.Reconciler
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		r.Logger.Errorf("invalid resource key: %s", key)
		return nil
	}
	r.Logger.Warnf("ns/name is %s/%s", namespace, name)

	// If our controller has configuration state, we'd "freeze" it and
	// attach the frozen configuration to the context.
	//    ctx = r.configStore.ToContext(ctx)

	// Get the resource with this namespace/name.
	original, err := r.Lister.ServiceSources(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		r.Logger.Errorf("resource %q no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}
	// Don't modify the informers copy.
	resource := original.DeepCopy()

	// Reconcile this copy of the resource and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.reconcile(ctx, resource)
	if equality.Semantic.DeepEqual(original.Status, resource.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err = r.updateStatus(resource); err != nil {
		r.Logger.Warnw("Failed to update resource status", zap.Error(err))
		r.Recorder.Eventf(resource, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}
	if reconcileErr != nil {
		r.Logger.Warnw("Internal error reconciling:", zap.Error(reconcileErr))
		r.Recorder.Event(resource, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}
	return reconcileErr
}

func (r *Reconciler) reconcile(ctx context.Context, source *v1alpha1.ServiceSource) error {

	if source.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	source.Status.InitializeConditions()

	if err := r.ReconcileSink(ctx, source); err != nil {
		return err
	}

	if err := r.reconcileService(ctx, source); err != nil {
		return err
	}

	source.Status.ObservedGeneration = source.Generation
	return nil
}

// reconcileService enforces the creation and lifecycle of the service. It assumes that the sink exists and is valid.
func (r *Reconciler) reconcileService(ctx context.Context, source *v1alpha1.ServiceSource) error {
	service, err := r.getService(ctx, source)

	// The state we would like to see show up in K8s eventually
	desired := resources.MakeService(source)

	if apierrs.IsNotFound(err) {
		// No service, must create it
		service, err := r.ServingClientSet.ServingV1beta1().Services(source.Namespace).Create(desired)
		if err != nil || service == nil {
			msg := "Failed to make Service."
			if err != nil {
				msg = msg + " " + err.Error()
			}
			source.Status.MarkServiceNotReady("FailedCreate", msg)
			return fmt.Errorf("failed to create Service: %s", err)
		}

		source.Status.MarkServiceDeploying()
		return nil
	} else if err != nil {
		r.Logger.Warnw("Failed get:", zap.Error(err))
		source.Status.MarkServiceNotReady("FailedGet", err.Error())
		return fmt.Errorf("failed to get Service: %s", err)
	}

	// The service exists; check if it looks like we expect
	if diff := cmp.Diff(desired.Spec, service.Spec); diff != "" {
		service.Spec = desired.Spec
		service, err := r.ServingClientSet.ServingV1beta1().Services(source.Namespace).Update(service)
		r.Logger.Desugar().Info("Service updated.",
			zap.Error(err), zap.Any("service", service), zap.String("diff", diff))
		source.Status.MarkServiceDeploying()
		return err
	}

	// Service exists and looks fine, propagate its status
	cond := serviceReadyCondition(service)
	if cond == nil {
		source.Status.MarkServiceReadyUnknown(serviceReadyUnknownReason, serviceReadyUnknownMessage)
		return ErrServiceReadyConditionMissing
	}

	switch cond.Status {
	case corev1.ConditionUnknown:
		source.Status.MarkServiceReadyUnknown(cond.Reason, cond.Message)
	case corev1.ConditionFalse:
		source.Status.MarkServiceNotReady(cond.Reason, cond.Message)
	case corev1.ConditionTrue:
		source.Status.MarkServiceReady()
	}

	source.Status.MarkAddress(service.Status.Address)
	source.Status.MarkURL(service.Status.URL)

	return nil
}

func (r *Reconciler) getService(ctx context.Context, owner metav1.Object) (*servingv1beta1.Service, error) {
	return r.ServingClientSet.ServingV1beta1().Services(owner.GetNamespace()).Get(resources.ServiceName(owner), metav1.GetOptions{})
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Reconciler) updateStatus(desired *v1alpha1.ServiceSource) (*v1alpha1.ServiceSource, error) {
	actual, err := r.Lister.ServiceSources(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return actual, nil
	}
	// Don't modify the informers copy
	existing := actual.DeepCopy()
	existing.Status = desired.Status
	return r.SourcesClientSet.SourcesV1alpha1().ServiceSources(desired.Namespace).UpdateStatus(existing)
}

// serviceReadyCondition retrieves the ready condition for a service.
func serviceReadyCondition(service *servingv1beta1.Service) *apis.Condition {
	for _, c := range service.Status.Conditions {
		if c.Type == servingv1beta1.ServiceConditionReady {
			return &c
		}
	}
	return nil
}
