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

type envConfig struct {
	// Source options
	Sink         string                    `envconfig:"K_SINK" required:"true"`
	OutputFormat v1alpha1.OutputFormatType `envconfig:"K_OUTPUT_FORMAT" required:"true"`

	// Database credentials
	ServiceAccountCreds []byte `envconfig:"GOOGLE_APPLICATION_CREDS_JSON" required:"true"`

	// Firestore collection to read from
	Collection string `envconfig:"FROM_COLLECTION" required:"true"`

	// Time frame to replay
	Since string        `envconfig:"SINCE" default:"0s"`
	since time.Duration // parsed version of above
}

func (env *envConfig) replayEvents(dbclient *firestore.Client, ceclient cloudevents.Client) error {
	ctx := context.Background()
	ticker := time.NewTicker(replaySpeed)
	defer ticker.Stop()

	var iter *firestore.DocumentIterator
	if env.since > 0 {
		// Query documents in the time frame specified
		cutoff := time.Now().Add(-1 * env.since).Unix()
		q := dbclient.Collection(env.Collection).Where("Time", ">", cutoff).OrderBy("Time", firestore.Asc)
		iter = q.Documents(ctx)
	} else {
		// Query all documents
		iter = dbclient.Collection(env.Collection).Documents(ctx)
	}

	// Iterate through each document and publish it
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
				Source: *cloudevents.ParseURLRef("http://example.com/"),
			}.AsV02(),
			Data: doc.Data(),
		}

		if _, err := ceclient.Send(ctx, event); err != nil {
			return err
		}

		<-ticker.C
	}

	return nil
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatal("Failed to process env: ", err)
	}
	since, err := time.ParseDuration(env.Since)
	if err != nil {
		log.Fatalf("Invalid duration for time since: %q: %v\n", env.Since, err)
	}
	env.since = since

	ceclient, err := ceclient.New(env.OutputFormat, env.Sink)
	if err != nil {
		log.Fatal("Could not create CloudEvents client: ", err)
	}

	dbclient, err := firestore.NewClient(context.Background(), firestore.DetectProjectID, option.WithCredentialsJSON(env.ServiceAccountCreds))
	if err != nil {
		log.Fatal("Could not create firestore client: ", err)
	}

	if err := env.replayEvents(dbclient, ceclient); err != nil {
		log.Fatalf("Could not replay events: %v\n", err)
	}

	log.Println("Success; exiting now")
}
