package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

var subscription = os.Getenv("PUBSUB_SUBSCRIPTION")

func Sub(ctx context.Context, projectID string, fn func(data []byte) error) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	sub := client.Subscription(subscription)

	sub.Receive(context.Background(), func(ctx context.Context, m *pubsub.Message) {
		fmt.Println("Received msg: ", string(m.Data))
		err = fn(m.Data)
		if err != nil {
			fmt.Println("Received err: ", err.Error())
			return
		}
		m.Ack()
	})

	return nil
}
