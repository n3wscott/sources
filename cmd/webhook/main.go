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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/logging/logkey"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/system"
	"knative.dev/pkg/version"
	"knative.dev/pkg/webhook"

	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
)

const (
	component = "webhook"
)

var (
	masterURL  = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
)

type Store interface {
	WatchConfigs(configmap.Watcher)
	ToContext(context.Context) context.Context
}

type StoreFactory func(*zap.SugaredLogger) Store

func SharedMain(handlers map[schema.GroupVersionKind]webhook.GenericCRD, factories ...StoreFactory) {
	flag.Parse()
	cm, err := configmap.Load("/etc/config-logging")
	if err != nil {
		log.Fatal("Error loading logging configuration:", err)
	}
	config, err := logging.NewConfigFromMap(cm)
	if err != nil {
		log.Fatal("Error parsing logging configuration:", err)
	}
	logger, atomicLevel := logging.NewLoggerFromConfig(config, component)
	defer logger.Sync()
	logger = logger.With(zap.String(logkey.ControllerType, component))

	logger.Info("Starting the Configuration Webhook")

	// Set up signals so we handle the first shutdown signal gracefully.
	ctx := signals.NewContext()

	clusterConfig, err := clientcmd.BuildConfigFromFlags(*masterURL, *kubeconfig)
	if err != nil {
		logger.Fatalw("Failed to get cluster config", zap.Error(err))
	}

	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		logger.Fatalw("Failed to get the client set", zap.Error(err))
	}

	if err := version.CheckMinimumVersion(kubeClient.Discovery()); err != nil {
		logger.Fatalw("Version check failed", err)
	}

	// Watch the logging config map and dynamically update logging levels.
	configMapWatcher := configmap.NewInformedWatcher(kubeClient, system.Namespace())
	configMapWatcher.Watch(logging.ConfigMapName(), logging.UpdateLevelFromConfigMap(logger, atomicLevel, component))

	// If you want to control Defaulting or Validation, you can attach config state
	// to the context by watching the configmap here, and then uncommenting the logic
	// below.
	stores := make([]Store, 0, len(factories))
	for _, sf := range factories {
		store := sf(logger)
		store.WatchConfigs(configMapWatcher)
		stores = append(stores, store)
	}

	if err = configMapWatcher.Start(ctx.Done()); err != nil {
		logger.Fatalw("Failed to start the ConfigMap watcher", zap.Error(err))
	}

	options := webhook.ControllerOptions{
		ServiceName:                 "webhook",
		DeploymentName:              "webhook",
		Namespace:                   system.Namespace(),
		Port:                        8443,
		SecretName:                  "webhook-certs",
		ResourceMutatingWebhookName: fmt.Sprintf("webhook.%s.knative.dev", system.Namespace()),
	}

	controller, err := webhook.New(
		kubeClient,
		options,
		map[string]webhook.AdmissionController{
			"/": webhook.NewResourceAdmissionController(handlers, options, false),
		},
		logger,
		func(ctx context.Context) context.Context {
			for _, store := range stores {
				ctx = store.ToContext(ctx)
			}
			return ctx
		},
	)
	if err != nil {
		logger.Fatalw("Failed to create admission controller", zap.Error(err))
	}

	if err = controller.Run(ctx.Done()); err != nil {
		logger.Fatalw("Failed to start the admission controller", zap.Error(err))
	}
}

func main() {
	handlers := map[schema.GroupVersionKind]webhook.GenericCRD{
		v1alpha1.SchemeGroupVersion.WithKind("JobSource"):     &v1alpha1.JobSource{},
		v1alpha1.SchemeGroupVersion.WithKind("CronJobSource"): &v1alpha1.CronJobSource{},
		v1alpha1.SchemeGroupVersion.WithKind("ServiceSource"): &v1alpha1.ServiceSource{},

		// Bind an alias of the Pod type to corev1.Pod for sidecar injection (via SetDefaults).
		// The Knative webhook will subscribe to Pods and all subresources, which includes Bindings.
		// We have to register a nop handler for Bindings so that all Pods don't get rejected.
		corev1.SchemeGroupVersion.WithKind("Pod"):     &v1alpha1.SourcePod{},
		corev1.SchemeGroupVersion.WithKind("Binding"): &v1alpha1.NOPBinding{},
	}
	SharedMain(handlers)

	// To setup a config "Store" to track a set of configurations and persist itself to
	// the context passed to webhook invocations, you would pass something like this to
	// SharedMain as well:
	// func(logger *zap.SugaredLogger) Store {
	// 	return apiconfig.NewStore(logger.Named("config-store"))
	// }
}
