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
	"errors"
	"fmt"
	"reflect"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/cronjobsource/resources"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/google/go-cmp/cmp"
	listers "github.com/n3wscott/sources/pkg/client/listers/sources/v1alpha1"
	"go.uber.org/zap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

var (
	errSinkMissing = errors.New("Sink missing from spec")
)

// Reconciler implements controller.Reconciler for CronJobSource resources.
type Reconciler struct {
	// +required
	*reconciler.Base

	// Lister allows us to query for CronJobSources
	// +required
	Lister listers.CronJobSourceLister
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile implements controller.Reconciler
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// If our controller has configuration state, we'd "freeze" it and
	// attach the frozen configuration to the context.
	//    ctx = r.configStore.ToContext(ctx)

	// Get the resource with this namespace/name.
	original, err := r.Lister.CronJobSources(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Errorf("resource %q no longer exists", key)
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
		logger.Warnw("Failed to update resource status", zap.Error(err))
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

func (r *Reconciler) reconcile(ctx context.Context, s *v1alpha1.CronJobSource) error {

	if s.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	s.Status.InitializeConditions()

	// Having a sink is a prereq for the cronjob
	if err := r.ReconcileSink(ctx, s); err != nil {
		return err
	}

	if err := r.reconcileCronJob(ctx, s); err != nil {
		return err
	}

	s.Status.ObservedGeneration = s.Generation
	return nil
}

// Ensure the CronJob exists according to the CronJobSourceSpec.
// Assumes Status.SinkURI is set.
func (r *Reconciler) reconcileCronJob(ctx context.Context, s *v1alpha1.CronJobSource) error {
	cronjob, err := r.getCronJob(ctx, s)
	desired := resources.MakeCronJob(s)

	if apierrs.IsNotFound(err) {
		// No job, must create it
		cronjob, err := r.KubeClientSet.BatchV1beta1().CronJobs(s.Namespace).Create(desired)
		if err != nil || cronjob == nil {
			msg := "Failed to make CronJob."
			if err != nil {
				msg = msg + " " + err.Error()
			}
			s.Status.MarkNoCronJob("FailedCreate", msg)
			return fmt.Errorf("failed to create CronJob: %s", err)
		}

		s.Status.MarkCronJobCreated()
		s.Status.PropagateCronJobStatus(&cronjob.Status)
		return nil
	} else if err != nil {
		r.Logger.Warnw("Failed get:", zap.Error(err))
		s.Status.MarkNoCronJob("FailedGet", err.Error())
		return fmt.Errorf("failed to get CronJob: %s", err)
	}

	// CronJob exists. Make sure it matches what we want.
	// The service exists; check if it looks like we expect
	if diff := cmp.Diff(desired.Spec, cronjob.Spec); diff != "" {
		cronjob.Spec = desired.Spec
		cronjob, err := r.KubeClientSet.BatchV1beta1().CronJobs(s.Namespace).Update(cronjob)
		r.Logger.Desugar().Info("CronJob updated.",
			zap.Error(err), zap.Any("cronjob", cronjob), zap.String("diff", diff))
		// TODO(spencer-p) What should the status be at this point?
		s.Status.MarkCronJobCreated()
		s.Status.PropagateCronJobStatus(&cronjob.Status)
		return err
	}

	// Copy its status.
	s.Status.MarkCronJobCreated()
	s.Status.PropagateCronJobStatus(&cronjob.Status)

	return nil
}

func (r *Reconciler) getCronJob(ctx context.Context, owner metav1.Object) (*batchv1beta1.CronJob, error) {
	return r.KubeClientSet.BatchV1beta1().CronJobs(owner.GetNamespace()).Get(resources.CronJobName(owner), metav1.GetOptions{})
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Reconciler) updateStatus(desired *v1alpha1.CronJobSource) (*v1alpha1.CronJobSource, error) {
	actual, err := r.Lister.CronJobSources(desired.Namespace).Get(desired.Name)
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
	return r.SourcesClientSet.SourcesV1alpha1().CronJobSources(desired.Namespace).UpdateStatus(existing)
}
