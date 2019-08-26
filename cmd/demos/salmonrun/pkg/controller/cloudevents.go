package controller

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go"
)

func receive(ctx context.Context, event cloudevents.Event, r *cloudevents.EventResponse) error {
	return nil
}
