package main

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
)

type DSClient interface {
	SaveMessage(ctx context.Context, message Message) error
	GetMessageByID(ctx context.Context, id string) (*Message, error)
}

type dsclient struct {
	client *datastore.Client
	key    string
}

func NewDatastore(client *datastore.Client, key string) DSClient {
	return &dsclient{
		client: client,
		key:    key,
	}
}

func (d *dsclient) SaveMessage(ctx context.Context, message Message) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	k := datastore.IncompleteKey(d.key, nil)
	_, err := d.client.Put(ctx, k, &message)
	return err
}

func (d *dsclient) GetMessageByID(ctx context.Context, id string) (*Message, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	q := datastore.NewQuery(d.key).Filter("ID =", id).Limit(1)

	var messages []Message
	_, err := d.client.GetAll(ctx, q, &messages)
	if err != nil {
		return nil, err
	}

	return &messages[0], nil
}
