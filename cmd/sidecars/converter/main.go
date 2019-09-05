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
	"io/ioutil"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	ceclient "github.com/n3wscott/sources/pkg/cloudeventclient"
)

type envConfig struct {
	// Source options
	Sink         string                    `envconfig:"K_SINK" required:"true"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT" required:"true"`

	// Receiving options
	Port string `envconfig:"PORT" required:"true"`
}

func makeReceive(ceclient cloudevents.Client) http.HandlerFunc {
	eventsrc := *cloudevents.ParseURLRef("http://todo.com")
	eventtype := "todo"

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Could not read POST body: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   eventtype,
				Source: eventsrc,
			}.AsV02(),
			Data: data,
		}

		if _, err := ceclient.Send(context.TODO(), event); err != nil {
			log.Println("Failed to send cloud event: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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

	http.Handle("/", makeReceive(ceclient))
	http.ListenAndServe(":"+env.Port, nil)
}
