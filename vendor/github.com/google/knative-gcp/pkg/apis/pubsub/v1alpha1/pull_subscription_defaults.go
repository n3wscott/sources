/*
Copyright 2019 Google LLC

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

package v1alpha1

import (
	"context"
	"time"

	"knative.dev/pkg/ptr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

const (
	defaultSecretName        = "google-cloud-key"
	defaultSecretKey         = "key.json"
	defaultRetentionDuration = 7 * 24 * time.Hour
	defaultAckDeadline       = 30 * time.Second
)

func (s *PullSubscription) SetDefaults(ctx context.Context) {
	s.Spec.SetDefaults(ctx)
}

// defaultSecretSelector is the default secret selector used to load the creds
// for the receive adapter to auth with Google Cloud.
func defaultSecretSelector() *corev1.SecretKeySelector {
	return &corev1.SecretKeySelector{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: defaultSecretName,
		},
		Key: defaultSecretKey,
	}
}

func (ss *PullSubscriptionSpec) SetDefaults(ctx context.Context) {
	if ss.AckDeadline == nil {
		ackDeadline := defaultAckDeadline
		ss.AckDeadline = ptr.String(ackDeadline.String())
	}

	if ss.RetentionDuration == nil {
		retentionDuration := defaultRetentionDuration
		ss.RetentionDuration = ptr.String(retentionDuration.String())
	}

	if ss.Secret == nil || equality.Semantic.DeepEqual(ss.Secret, &corev1.SecretKeySelector{}) {
		ss.Secret = defaultSecretSelector()
	}

	switch ss.Mode {
	case ModeCloudEventsBinary, ModeCloudEventsStructured, ModePushCompatible:
		// Valid Mode.
	default:
		// Default is CloudEvents Binary Mode.
		ss.Mode = ModeCloudEventsBinary
	}
}
