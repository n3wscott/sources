// This program generates a set of mock datapoints in Google Cloud Firestore
// for testing purposes.  To use it, make sure you first have a service account
// with permission to write to Firestore and have the JSON credentials on hand.
// Then run:
//	go build
//	GOOGLE_APPLICATION_CREDENTIALS=path/to/your/creds.json ./gen-fake-events
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/api/option"
)

type Config struct {
	// Options for the generated events
	Collection string `envconfig:"COLLECTION" default:"purchases"`
	NumDocs    int    `envconfig:"NUMDOCS" default:"10"`

	// Database credentials
	ServiceAccountCreds []byte `envconfig:"GOOGLE_APPLICATION_CREDS_JSON" required:"true"`
}

var (
	// Change these demo values as needed
	names = []string{"Grant", "Scott", "Spencer", "Nacho", "Akash", "Adam", "Grace", "Xiyue"}
	items = []string{"banana", "apple", "chewy granola bar", "banana sticker", "crepe", "boat", "salad", "smoothie", "jacket", "computer", "custom resource definition"}
)

// Types that we will use to populate the DB
type Purchase struct {
	Name    string
	CCId    int32
	Address string
	Items   []*Item
	Time    int64
}

type Item struct {
	Name  string
	Price int32
}

// randAddress generates a random shipping address (or IP address :P).
func randAddress() string {
	return fmt.Sprintf("https://%d.%d.%d.%d",
		rand.Intn(255),
		rand.Intn(255),
		rand.Intn(255),
		rand.Intn(255),
	)
}

func main() {
	ctx := context.Background()

	var env Config
	envconfig.MustProcess("", &env)

	client, err := firestore.NewClient(context.Background(), firestore.DetectProjectID, option.WithCredentialsJSON(env.ServiceAccountCreds))
	if err != nil {
		log.Fatal("Failed to open firestore client: ", err)
	}
	defer client.Close()

	// Create memory we will mutate for each doc
	var item Item
	purchase := &Purchase{
		Items: []*Item{&item},
	}

	rand.Seed(time.Now().UnixNano())

	log.Printf("Writing documents to collection %q...\n", env.Collection)
	collRef := client.Collection(env.Collection)
	for i := 0; i < env.NumDocs; i++ {
		// Randomize the doc
		purchase.Name = names[rand.Intn(len(names))]
		purchase.Address = randAddress()
		purchase.Time = time.Now().Unix()
		item.Name = items[rand.Intn(len(items))]
		item.Price = int32(rand.Intn(10 + 1))

		// Push the data to the DB
		_, _, err := collRef.Add(ctx, purchase)
		if err != nil {
			log.Fatal("Failed to commit: ", err)
		}
	}

	log.Println("Success")
}
