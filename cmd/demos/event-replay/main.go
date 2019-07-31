package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	ceclient "github.com/n3wscott/sources/pkg/cloudeventclient"
)

type envConf struct {
	Sink         string                    `envconfig:"K_SINK"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT"`
}

func main() {
	var env envConf
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env: ", err)
	}

	if env.Sink == "" {
		log.Fatal("Runtime contract violated, no sink")
	} else if env.OutputFormat == "" {
		log.Fatal("Runtime contract violated, no output format")
	}

	client, err := ceclient.New(env.OutputFormat, env.Sink)
	if err != nil {
		log.Fatal("Could not create CloudEvents client: ", err)
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			Type:   "dev.knative.eventing.n3wscott.sources.demos.event-replay",
			Source: *cloudevents.ParseURLRef("http://dev.tryransom.com/"),
		}.AsV02(),
		Data: "Hello, world!",
	}

	if _, err := client.Send(context.Background(), event); err != nil {
		log.Fatal("Could not send event: ", err)
	}

	log.Println("Success; exiting now")
}
