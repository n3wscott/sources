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

	"cloud.google.com/go/firestore"
)

const (
	// The firestore collection we will write to
	collection = "purchases"
	numDocs    = 100
)

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

	client, err := firestore.NewClient(ctx, firestore.DetectProjectID)
	if err != nil {
		log.Fatal("Failed to open firestore client: ", err)
	}
	defer client.Close()

	// Create memory we will mutate for each doc
	var item Item
	purchase := &Purchase{
		Items: []*Item{&item},
	}

	log.Printf("Writing documents to collection %q...\n", collection)
	for i := 0; i < numDocs; i++ {
		// Randomize the doc
		purchase.Name = names[rand.Intn(len(names))]
		purchase.Address = randAddress()
		item.Name = items[rand.Intn(len(items))]
		item.Price = int32(rand.Intn(10 + 1))

		// Push the data to the DB
		_, _, err := client.Collection(collection).Add(ctx, purchase)
		if err != nil {
			log.Fatal("Failed to commit: ", err)
		}
	}

	log.Println("Success")
}
