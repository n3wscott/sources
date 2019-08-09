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
	"errors"
	"fmt"
	"reflect"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"github.com/n3wscott/sources/pkg/reconciler"
	"github.com/n3wscott/sources/pkg/reconciler/jobsource/resources"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	listers "github.com/n3wscott/sources/pkg/client/listers/sources/v1alpha1"
	"go.uber.org/zap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
)

var (
	errSinkMissing = errors.New("Sink missing from spec")
)

// Reconciler implements controller.Reconciler for JobSource resources.
type Reconciler struct {
	// +required
	*reconciler.Base

	// Lister allows us to query for JobSources
	// +required
	Lister listers.JobSourceLister
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
	original, err := r.Lister.JobSources(namespace).Get(name)
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

func (r *Reconciler) reconcile(ctx context.Context, js *v1alpha1.JobSource) error {

	if js.GetDeletionTimestamp() != nil {
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	js.Status.InitializeConditions()

	// Having a sink is a prereq for starting the job, so we reconcile the sink first
	if err := r.reconcileSink(ctx, js); err != nil {
		return err
	}

	if err := r.reconcileJob(ctx, js); err != nil {
		return err
	}

	js.Status.ObservedGeneration = js.Generation
	return nil
}

// reconcileJob enforces the creation and lifecycle of the job. It assumes that the sink exists and is valid.
func (r *Reconciler) reconcileJob(ctx context.Context, js *v1alpha1.JobSource) error {
	job, err := r.getJob(ctx, js)

	if apierrs.IsNotFound(err) {
		// No job, must create it
		job = resources.MakeJob(js)

		job, err := r.KubeClientSet.BatchV1().Jobs(js.Namespace).Create(job)
		if err != nil || job == nil {
			msg := "Failed to make Job."
			if err != nil {
				msg = msg + " " + err.Error()
			}
			js.Status.MarkJobFailed("FailedCreate", msg)
			return fmt.Errorf("failed to create Job: %s", err)
		}

		js.Status.MarkJobRunning("Created Job %q.", job.Name)
		return nil
	} else if err != nil {
		r.Logger.Warnw("Failed get:", zap.Error(err))
		js.Status.MarkJobFailed("FailedGet", err.Error())
		return fmt.Errorf("failed to get Job: %s", err)
	}

	// Job exists, check if it is done
	if cond := getJobCompletedCondition(job); cond != nil && jobConditionSucceeded(cond) {
		js.Status.MarkJobSucceeded()
	} else if cond != nil && jobConditionFailed(cond) {
		js.Status.MarkJobFailed(cond.Reason, cond.Message)
	} else {
		// Job is not finished, make sure the status reflects that
		if !js.Status.IsJobRunning() {
			js.Status.MarkJobRunning("Job %q already exists.", job.Name)
		}
	}

	return nil
}

// reconcileSink attempts to reconcile the sink object reference to a URI and set the sink in the JobSourceStatus.
func (r *Reconciler) reconcileSink(ctx context.Context, js *v1alpha1.JobSource) error {
	if js.Spec.Sink == nil {
		js.Status.MarkNoSink("Missing", "Sink missing from spec")
		return errSinkMissing
	}

	ref := js.Spec.Sink
	if ref.Namespace == "" {
		ref.Namespace = js.Namespace
	}

	desc := fmt.Sprintf("%s/%s, %s", js.Namespace, js.Name, js.GroupVersionKind().String())
	uri, err := r.SinkReconciler.GetSinkURI(ref, js, desc)
	if err != nil {
		js.Status.MarkNoSink("NotFound", "Could not get sink URI from %s/%s: %v", ref.Namespace, ref.Name, err)
		return err
	}

	js.Status.MarkSink(uri)

	return nil
}

func (r *Reconciler) getJob(ctx context.Context, owner metav1.Object) (*batchv1.Job, error) {
	return r.KubeClientSet.BatchV1().Jobs(owner.GetNamespace()).Get(resources.JobName(owner), metav1.GetOptions{})
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Reconciler) updateStatus(desired *v1alpha1.JobSource) (*v1alpha1.JobSource, error) {
	actual, err := r.Lister.JobSources(desired.Namespace).Get(desired.Name)
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
	return r.SourcesClientSet.SourcesV1alpha1().JobSources(desired.Namespace).UpdateStatus(existing)
}

// getJobCompletedCondition finds a JobCondition of the Job that has information about its completedness.
func getJobCompletedCondition(job *batchv1.Job) *batchv1.JobCondition {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return &c
		}
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return &c
		}
	}
	return nil
}

// jobConditionSucceeded returns true if the given JobCondition marks its job as succeeded.
func jobConditionSucceeded(c *batchv1.JobCondition) bool {
	return c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue
}

// jobConditionFailed returns true if the given JobCondition marks its job as failed.
func jobConditionFailed(c *batchv1.JobCondition) bool {
	return c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue
}
