package service

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	log "github.com/sirupsen/logrus"
)

type FirestoreClient interface {
	GetItemsFromCollection(context.Context, string) ([]*map[string]interface{}, error)
	StoreItemToCollection(context.Context, string, *map[string]interface{}) (string, error)
	UpdateItem(context.Context, string, string, map[string]interface{}) (error)
	UpdateValueOnItem(context.Context, string, string, string, interface{}) (error)
}

type Firestore struct {
	client	*firestore.Client
}

func InitializeFirestore(ctx context.Context) *Firestore {
	projectID := "natalya"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Errorf("Failed to create client: %v", err)
		return nil
	}

	c := Firestore{client}
	return &c
}

func (c *Firestore) GetItemsFromCollection(ctx context.Context, path string) (items []*map[string]interface{}, err error) {
	iter := c.client.Collection(path).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		item := doc.Data()
		item["_refID"] = doc.Ref.ID
		items = append(items, &item)
	}
	return items, nil
}

func (c *Firestore) StoreItemToCollection(ctx context.Context, path string, item *map[string]interface{}) (id string, err error) {
	doc, _, err := c.client.Collection(path).Add(ctx, item)
	return doc.ID, err
}

func (c *Firestore) UpdateItem(ctx context.Context, path string, refID string, updates map[string]interface{}) (err error) {
	_, err = c.client.Collection(path).Doc(refID).Set(ctx, updates, firestore.MergeAll)
	return err
}

func (c *Firestore) UpdateValueOnItem(ctx context.Context, path string, refID string, key string, value interface{}) (err error) {
	_, err = c.client.Collection(path).Doc(refID).Update(ctx, []firestore.Update{
		{
			Path: key,
			Value:value,
		},
	})
	return err
}