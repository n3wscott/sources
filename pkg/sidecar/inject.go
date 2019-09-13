/*
Copyright 2019 The Knative Authors.

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

package sidecar

import (
	"fmt"
	"log" // TODO(spencer-p) use zap
	"os"
	"strconv"
	"strings"

	"knative.dev/pkg/apis"

	corev1 "k8s.io/api/core/v1"
)

const (
	IMAGE_KEY    = "K_SOURCE_ADAPTER_IMAGE"
	ADAPTER_PORT = 38080
	LABEL_NAME   = "eventing.knative.dev/inject"
	ADAPTER_KEY  = "cloudevents-adapter"
	FILTER_KEY   = "cloudevents-filter"

	CE_LABEL_PREFIX  = "cloudevents.io/"
	EVENT_SOURCE_KEY = "source"
	EVENT_TYPE_KEY   = "type"

	SIDECAR_NAME = "knative-sidecar"

	PORT_MAX = 1<<16 - 1
)

type SidecarArgs struct {
	// Always required.
	SinkURIVar, OutputFormatVar *corev1.EnvVar
	Image                       string
	Port                        int32

	// Required for the adapter.
	EventType, EventSource string

	// Optional for adapter.
	AddExtensions map[string]string

	// Optional for filter.
	FilterExtensions map[string]string
}

// ShouldInjectAdapter returns true if the adapter sidecar should be injected and has not already been injected.
func ShouldInjectAdapter(pod *corev1.Pod) bool {
	// Check for the label
	injectText, ok := pod.GetLabels()[LABEL_NAME]
	if !ok {
		return false
	}

	// Check for a request for the adapter in the label
	if !strings.Contains(injectText, ADAPTER_KEY) {
		return false
	}

	// Check if we've already done it
	// TODO(spencer-p) Use a label or annotation to mark this.
	for _, c := range pod.Spec.Containers {
		if c.Name == SIDECAR_NAME {
			return false
		}
	}

	return true
}

func constructArgs(pod *corev1.Pod) (*SidecarArgs, *apis.FieldError) {
	var errs *apis.FieldError

	_, srcSinkURI, srcOutputFormat, ok := getSourceContainer(pod)
	if !ok {
		if srcSinkURI == nil || srcSinkURI.Value == "" {
			errs = errs.Also(apis.ErrMissingField("K_SINK").ViaField("spec.containers[i].Env"))
		}
		if srcOutputFormat == nil || srcOutputFormat.Value == "" {
			errs = errs.Also(apis.ErrMissingField("K_OUTPUT_FORMAT").ViaField("spec.containers[i].Env"))
		}
	}

	port, err := findPort(pod, ADAPTER_PORT)
	if err != nil {
		errs = errs.Also(apis.ErrGeneric("No free port: " + err.Error()).ViaField("spec.containers[i].ports"))
	}

	ceSrc, labelerr := readAnnotation(pod, CE_LABEL_PREFIX+EVENT_SOURCE_KEY)
	errs = errs.Also(labelerr)

	ceType, labelerr := readAnnotation(pod, CE_LABEL_PREFIX+EVENT_TYPE_KEY)
	errs = errs.Also(labelerr)

	// TODO(spencer-p) This is the only item not found in the pod - make sense in a field error?
	img := os.Getenv(IMAGE_KEY)
	if img == "" {
		errs = errs.Also(apis.ErrMissingField("$" + IMAGE_KEY))
	}

	return &SidecarArgs{
		SinkURIVar:      srcSinkURI,
		OutputFormatVar: srcOutputFormat,
		Image:           img,
		Port:            port,
		EventSource:     ceSrc,
		EventType:       ceType,
	}, errs
}

// injectSidecar rewrites the containers of the pod such that the sidecar is injected as
// configured by SideCarArgs.
func injectSidecar(pod *corev1.Pod, args *SidecarArgs) {
	// Construct new container
	portStr := strconv.Itoa(int(args.Port))
	sidecarContainer := corev1.Container{
		Name:  SIDECAR_NAME,
		Image: args.Image,
		Ports: []corev1.ContainerPort{{
			ContainerPort: args.Port,
		}},
		Env: []corev1.EnvVar{{
			Name:  "PORT",
			Value: portStr,
		}, {
			Name:  "K_SINK",
			Value: args.SinkURIVar.Value,
		}, {
			Name:  "K_OUTPUT_FORMAT",
			Value: args.OutputFormatVar.Value,
		}, {
			Name:  "EVENT_SOURCE",
			Value: args.EventSource,
		}, {
			Name:  "EVENT_TYPE",
			Value: args.EventType,
		}},
	}

	// Rewire the source container
	args.SinkURIVar.Value = "http://127.0.0.1:" + portStr

	// Add the sidecar container
	pod.Spec.Containers = append(pod.Spec.Containers, sidecarContainer)
}

// Inject inspects and rewrites the Pod to have a Knative Adapter/Filter sidecar.
func Inject(pod *corev1.Pod) {
	args, errs := constructArgs(pod)
	if errs != nil {
		log.Printf("Cannot inject sidecar: %v\n", errs)
		return
	}

	injectSidecar(pod, args)
}

// getSourceContainer finds a container that looks like it was configured as a source. It returns
// the container itself, pointers to its useful environment variables, and an OK bool signifying
// that the container was found.
func getSourceContainer(pod *corev1.Pod) (
	container *corev1.Container,
	sinkURI *corev1.EnvVar,
	outputFormat *corev1.EnvVar,
	ok bool) {

	for i := range pod.Spec.Containers {
		for j, evar := range pod.Spec.Containers[i].Env {
			switch evar.Name {
			case "K_SINK":
				sinkURI = &pod.Spec.Containers[i].Env[j]
				container = &pod.Spec.Containers[i]
			case "K_OUTPUT_FORMAT":
				outputFormat = &pod.Spec.Containers[i].Env[j]
				container = &pod.Spec.Containers[i]
			}
		}

		// If we found the container, don't inspect others.
		// This guarantees we will not take K_SINK from one container and K_OUTPUT_FORMAT from another.
		if container != nil {
			break
		}
	}

	if sinkURI != nil && outputFormat != nil {
		ok = true
	}

	return
}

// findPort returns the first unused port in the container that is greater than or equal to startWith.
// If no ports are available it returns an error.
func findPort(pod *corev1.Pod, startWith int32) (int32, error) {
	usedports := make(map[int32]struct{})

	for _, c := range pod.Spec.Containers {
		for _, cport := range c.Ports {
			usedports[cport.ContainerPort] = struct{}{}
			usedports[cport.HostPort] = struct{}{}
		}
	}

	for port := startWith; port <= PORT_MAX; port++ {
		if _, used := usedports[port]; !used {
			return port, nil
		}
	}

	// Somehow all the ports are taken.
	return 0, fmt.Errorf("No ports on container >= %d available", startWith)
}

// readAnnotation returns the value of a label in a pod. If the value is missing, it returns a missing field error.
func readAnnotation(pod *corev1.Pod, key string) (string, *apis.FieldError) {
	val, ok := pod.GetAnnotations()[key]
	if !ok {
		return "", apis.ErrMissingField(key).ViaField("annotations")
	}
	return val, nil
}
