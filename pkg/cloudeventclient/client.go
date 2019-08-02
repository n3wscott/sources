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
package cloudeventclient

import (
	"fmt"
	gohttp "net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"knative.dev/pkg/tracing"
)

// New creates a default client using one of the two source OutputFormatTypes.
func New(format v1alpha1.OutputFormatType, target ...string) (cloudevents.Client, error) {
	var tOpts []http.Option
	switch format {
	case v1alpha1.OutputFormatBinary:
		tOpts = append(tOpts, cloudevents.WithBinaryEncoding())
	case v1alpha1.OutputFormatStructured:
		tOpts = append(tOpts, cloudevents.WithStructuredEncoding())
	default:
		return nil, fmt.Errorf("Unknown OutputFormatType: %v", format)
	}
	tOpts = append(tOpts, http.WithMiddleware(tracing.HTTPSpanMiddleware))
	if len(target) > 0 && target[0] != "" {
		tOpts = append(tOpts, cloudevents.WithTarget(target[0]))
	}

	// Make an http transport for the CloudEvents client.
	t, err := cloudevents.NewHTTPTransport(tOpts...)
	if err != nil {
		return nil, err
	}
	// Add output tracing.
	t.Client = &gohttp.Client{
		Transport: &ochttp.Transport{
			Propagation: &b3.HTTPFormat{},
		},
	}

	// Use the transport to make a new CloudEvents client.
	c, err := cloudevents.NewClient(t,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
	)

	if err != nil {
		return nil, err
	}
	return c, nil
}
