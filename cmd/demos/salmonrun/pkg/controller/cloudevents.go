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

package controller

import (
	"context"
	"log"
	"math/rand"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
)

func salmonEventReceiver(ctx context.Context, event cloudevents.Event, _ *cloudevents.EventResponse) error {
	var msg Message
	if err := event.DataAs(&msg); err != nil {
		log.Println("could not understand cloudevent: ", err)
		return err
	}

	// The event we received is "to" a salmon player "from" a bear that ate them.

	send, ok := conns[msg.To.Key()]
	if !ok {
		log.Printf("event for missing player %s/%s\n", msg.To.Name, msg.To.UUID)
		return nil // not technically an error
	}

	go send(&msg)
	return nil
}

func bearEventReceiver(ctx context.Context, event cloudevents.Event, _ *cloudevents.EventResponse) error {
	var msg Message
	if err := event.DataAs(&msg); err != nil {
		log.Println("could not understand cloudevent: ", err)
		return err
	}

	// Get a random user.
	// TODO This is broke
	if len(conns) == 0 {
		// If there are no connections, there is nothing to do
		log.Println("no bears, lonely salmon")
		return nil
	}

	var key string
	i := 0
	n := rand.Intn(len(conns))
	for key = range conns {
		i += 1
		if i == n {
			break
		}
	}

	send, ok := conns[key]
	if !ok {
		log.Println("no bears to deliver to")
		return nil
	}

	// Mark the message so that the handler knows to ignore dups
	if msg.Nonce == "" {
		msg.Nonce = uuid.New().String()
	}

	// Give the bear the salmon
	go send(&msg)

	// They have to beat this timeout though
	time.AfterFunc(time.Second*3, func() {
		timeoutchan, ok := timeoutchans[key]
		if !ok {
			log.Println("race condition - bear has connection but no timeout channel")
		}
		timeoutchan <- Message{
			To:    msg.From, // we are now sending TO the player we got this FROM
			Nonce: msg.Nonce,
		}
	})
	return nil
}
