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
	"fmt"
	"html/template"
	"log"
	gohttp "net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"knative.dev/pkg/tracing"
)

const (
	VERSION = "v0.0.1"

	EVENT_SOURCE = "http://github.com/n3wscott/sources/cmd/demos/forwarder"
	EVENT_TYPE   = "com.tryransom.forwarder"
)

type envConfig struct {
	// Source options
	Sink         string                    `envconfig:"K_SINK" required:"true"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT" required:"true"`

	// Ko deployment options
	DataPath string `envconfig:"KO_DATA_PATH"`

	// Service options
	Port int `envconfig:"PORT"`
}

func makeIndexHandler(dir string) gohttp.HandlerFunc {
	templates := template.Must(template.ParseGlob(path.Join(dir, "*")))

	return func(w gohttp.ResponseWriter, r *gohttp.Request) {
		if r.Method != "GET" {
			w.WriteHeader(gohttp.StatusMethodNotAllowed)
			return
		}

		log.Println(r.Method, r.URL)

		err := templates.Execute(w, nil)
		if err != nil {
			log.Printf("Error serving index template: %v\n", err)
		}
	}
}

func makeImporterHandle(client cloudevents.Client) interface{} {
	return func(event cloudevents.Event, r *cloudevents.EventResponse) error {
		log.Printf("Importing an event: %+v\n", event)

		response, err := client.Send(context.Background(), event)
		if err != nil {
			log.Printf("Failed to send event: %v\n", err)
			return err
		}

		// If the response was not good, construct a friendly response
		if response == nil {
			responseval := cloudevents.NewEvent()
			response = &responseval
			response.SetSource(EVENT_SOURCE)
			response.SetType(EVENT_TYPE)
			response.SetID(uuid.New().String())

			if err := response.SetData("ok"); err != nil {
				return err
			}
		}

		r.RespondWith(gohttp.StatusOK, response)

		return nil
	}
}

func convert(ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error) {
	if msg, ok := m.(*http.Message); ok {

		vals, err := url.ParseQuery(string(msg.Body))
		if err != nil {
			// TODO(spencer-p) this is a bug for incoming requests
			return nil, nil
		}

		dataslice, ok := vals["data"]
		if !ok {
			// TODO(spencer-p) bug see above
			return nil, nil
		}

		data := dataslice[0]

		event := cloudevents.NewEvent()
		event.SetSource(EVENT_SOURCE)
		event.SetType(EVENT_TYPE)
		event.SetID(uuid.New().String())

		if err := event.SetData(data); err != nil {
			return nil, err
		}

		return &event, nil
	}
	return nil, err
}

func (env *envConfig) registerHandlers(mux *gohttp.ServeMux) {
	mux.HandleFunc("/", makeIndexHandler(env.DataPath))

	mux.HandleFunc("/healthz", func(w gohttp.ResponseWriter, r *gohttp.Request) {
		w.WriteHeader(200)
	})

	gohttp.HandleFunc("/versionz", func(w gohttp.ResponseWriter, r *gohttp.Request) {
		w.Write([]byte(VERSION))
	})
}

func (env *envConfig) makeClient(mux *gohttp.ServeMux) (cloudevents.Client, error) {
	// Build transport options
	tOpts := []http.Option{
		http.WithMiddleware(tracing.HTTPSpanMiddleware),
		cloudevents.WithTarget(env.Sink),
		cloudevents.WithPath("/import"),
		cloudevents.WithPort(env.Port),
	}
	switch env.OutputFormat {
	case v1alpha1.OutputFormatBinary:
		tOpts = append(tOpts, cloudevents.WithBinaryEncoding())
	case v1alpha1.OutputFormatStructured:
		tOpts = append(tOpts, cloudevents.WithStructuredEncoding())
	default:
		return nil, fmt.Errorf("Unknown OutputFormatType: %q", env.OutputFormat)
	}

	// Construct the transport
	transport, err := cloudevents.NewHTTPTransport(tOpts...)
	if err != nil {
		return nil, err
	}

	// Add output tracing.
	transport.Client = &gohttp.Client{
		Transport: &ochttp.Transport{
			Propagation: &b3.HTTPFormat{},
		},
	}

	// Set the mux
	transport.Handler = mux

	// Construct the client
	ceclient, err := cloudevents.NewClient(transport,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
		cloudevents.WithConverterFn(convert),
	)
	if err != nil {
		return nil, err
	}

	return ceclient, nil
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env: ", err)
	}

	if env.Port == 0 {
		env.Port = 80
	}

	mux := gohttp.NewServeMux()
	env.registerHandlers(mux)
	ceclient, err := env.makeClient(mux)
	if err != nil {
		log.Fatal("Could not create CloudEvents client: ", err)
	}

	ctx, shutdown := context.WithCancel(context.Background())
	go func() {
		log.Fatal(ceclient.StartReceiver(ctx, makeImporterHandle(ceclient)))
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutdown signal received, exiting...")

	shutdown()
}
