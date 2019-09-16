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

package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	ceclient "github.com/n3wscott/sources/pkg/cloudeventclient"
)

type envConfig struct {
	// Source options
	Sink         string                    `envconfig:"K_SINK" required:"true"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT" required:"true"`
	Source       string                    `envconfig:"EVENT_SOURCE" required:"true"`
	Type         string                    `envconfig:"EVENT_TYPE" required:"true"`

	// Receiving options
	Port        string `envconfig:"PORT" required:"true"`
	ServePublic bool   `envconfig:"SERVE_PUBLICLY" default:"false"`
}

func makeReceive(ceclient cloudevents.Client, eventtype string, eventsrc cloudevents.URLRef) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Path[len("/"):]

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Could not read POST body:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   eventtype,
				Source: eventsrc,
				ID:     id, // will default if it's empty string
			}.AsV02(),
			Data: data,
		}

		log.Printf("Sending event with %d bytes of data\n", len(data))
		resp, err := ceclient.Send(r.Context(), event)
		if err != nil {
			log.Println("Failed to send cloud event:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resp == nil {
			// No response from sink
			w.WriteHeader(http.StatusOK)
			return
		}

		respBytes, err := resp.DataBytes()
		if err != nil {
			log.Println("Failed to read response:", err)
			// Assume success so that event is not duplicated
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env: ", err)
	}

	ceclient, err := ceclient.New(env.OutputFormat, env.Sink)
	if err != nil {
		log.Fatal("Could not create CloudEvents client: ", err)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	http.Handle("/", makeReceive(ceclient, env.Type, *cloudevents.ParseURLRef(env.Source)))

	http.HandleFunc("/quitquitquit", func(w http.ResponseWriter, r *http.Request) {
		shutdownChan <- syscall.SIGTERM
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	addr := "127.0.0.1"
	if env.ServePublic == true {
		addr = "0.0.0.0"
	}
	s := http.Server{
		Addr: addr + ":" + env.Port,
	}
	go func() {
		log.Println("Sink is configured as", env.Sink)
		log.Println("Starting adapter server")
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-shutdownChan
	log.Println("Received shutdown signal")
	s.Shutdown(context.Background())
	os.Exit(0)
}
