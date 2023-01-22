package main

import (
	"context"
	"os"

	"cloud.google.com/go/pubsub"
)

var topic = os.Getenv("PUBSUB_TOPIC")

func Pub(ctx context.Context, projectID string, data []byte) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	t := client.Topic(topic)

	exists, err := t.Exists(ctx)
	if err != nil {
		return err
	}

	if !exists {
		t, err = client.CreateTopic(ctx, topic)
		if err != nil {
			return err
		}
	}

	msg := pubsub.Message{
		Data: data,
	}

	res := t.Publish(ctx, &msg)

	_, err = res.Get(ctx)
	if err != nil {
		return err
	}

	return nil
}
