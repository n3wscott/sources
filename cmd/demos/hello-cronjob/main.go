package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"knative.dev/eventing/pkg/kncloudevents"

	"github.com/kelseyhightower/envconfig"
)

type Conf struct {
	Sink string `envconfig:"K_SINK" required:"true"`

	// 100 means always fail, 0 means never fail
	FailOdds int `envconfig:"FAIL_ODDS_PERCENT" default:"0"`
}

func main() {
	log.Println("Starting up source container")

	// Parse environment
	var env Conf
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env: ", err)
	}

	log.Println("Sink endpoint is", env.Sink)

	rand.Seed(time.Now().UnixNano())
	if env.FailOdds > rand.Intn(100) {
		log.Fatal("Luck was not on your side.")
	}

	client, err := kncloudevents.NewDefaultClient(env.Sink)
	if err != nil {
		log.Fatal("Could not create a client: ", err)
	}

	event := cloudevents.NewEvent()
	event.SetSource("https://github.com/n3wscott/sources/cmd/demos/hello-cronjob")
	event.SetType("com.github.n3wscott.sources.demos.hello-cronjob.hello")
	if err := event.SetData(map[string]interface{}{
		"Hello": "world!",
	}); err != nil {
		log.Fatal("Failed to set cloud event data: ", err)
	}

	if _, err = client.Send(context.Background(), event); err != nil {
		log.Fatal("Failed to send cloud event: ", err)
	}
}
