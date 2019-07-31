package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/n3wscott/sources/pkg/apis/sources/v1alpha1"
	ceclient "github.com/n3wscott/sources/pkg/cloudeventclient"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	// Replay speed will be rate limited so the generation is somewhat observable
	replaySpeed = time.Second / 5
)

type envConf struct {
	// Source options
	Sink         string                    `envconfig:"K_SINK" required:"true"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT" required:"true"`

	// Database credentials
	ServiceAccountCreds []byte `envconfig:"GOOGLE_APPLICATION_CREDS_JSON" required:"true"`
}

func replayEvents(dbclient *firestore.Client, ceclient cloudevents.Client) error {

	ticker := time.NewTicker(replaySpeed)
	defer ticker.Stop()

	// Iterate through each document and publish it
	iter := dbclient.Collection("purchases").Documents(context.Background())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Could not read from database, failed to iterate: %v", err)
		}

		// Generate cloud event w/o parsing the doc at all
		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   "dev.knative.eventing.n3wscott.sources.demos.event-replay",
				Source: *cloudevents.ParseURLRef("http://dev.tryransom.com/"),
			}.AsV02(),
			Data: doc.Data(),
		}

		if _, err := ceclient.Send(context.Background(), event); err != nil {
			return err
		}

		<-ticker.C
	}

	return nil
}

func main() {
	var env envConf
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env: ", err)
	}

	ceclient, err := ceclient.New(env.OutputFormat, env.Sink)
	if err != nil {
		log.Fatal("Could not create CloudEvents client: ", err)
	}

	dbclient, err := firestore.NewClient(context.Background(), firestore.DetectProjectID, option.WithCredentialsJSON(env.ServiceAccountCreds))
	if err != nil {
		log.Fatal("Could not create firestore client: ", err)
	}

	if err := replayEvents(dbclient, ceclient); err != nil {
		log.Fatalf("Could not replay events: %v\n", err)
	}

	log.Println("Success; exiting now")
}
