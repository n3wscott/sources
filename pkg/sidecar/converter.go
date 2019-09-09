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
	"log" // TODO(spencer-p) use zap
	"os"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

const (
	IMAGE_KEY       = "K_SOURCE_CONVERTER_IMAGE"
	CONVERT_PORT    = 57070
	ANNOTATION_NAME = "inject-knative-source-converter"
	SIDECAR_NAME    = "source-converter"
)

var (
	CONVERT_PORT_STR = strconv.Itoa(CONVERT_PORT)
)

func ShouldAddConverter(pod *corev1.Pod) bool {
	// Check for the label
	should, ok := pod.GetAnnotations()[ANNOTATION_NAME]
	if !ok || should != "true" {
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

func AddConverter(pod *corev1.Pod) {
	_, srcSinkURI, srcOutputFormat, ok := getSourceContainer(pod)
	if !ok {
		log.Println("Source pod is missing a container with proper env vars")
		return
	}

	img := os.Getenv(IMAGE_KEY)
	if img == "" {
		log.Println("Missing image key for converter")
		return
	}

	// Construct new container
	convertContainer := corev1.Container{
		Name:  "source-converter",
		Image: img,
		Ports: []corev1.ContainerPort{{
			HostPort:      CONVERT_PORT,
			ContainerPort: CONVERT_PORT,
		}},
		Env: []corev1.EnvVar{{
			Name:  "PORT",
			Value: CONVERT_PORT_STR,
		}, {
			Name:  "K_SINK",
			Value: srcSinkURI.Value,
		}, {
			Name:  "K_OUTPUT_FORMAT",
			Value: srcOutputFormat.Value,
		}, {
			Name:  "EVENT_SOURCE",
			Value: "http://todo",
		}, {
			Name:  "EVENT_TYPE",
			Value: "todo",
		}},
	}

	// Rewire the source container
	srcSinkURI.Value = "http://127.0.0.1:" + CONVERT_PORT_STR

	// Add the convert container
	pod.Spec.Containers = append(pod.Spec.Containers, convertContainer)
}

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
	}

	if sinkURI != nil && outputFormat != nil {
		ok = true
	}

	return
}
