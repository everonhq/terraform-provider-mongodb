package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	URL string
}

func (c *Config) loadAndValidate() (*mongo.Client, error) {

	mURI, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}

	mURI.RawQuery = mURI.Query().Encode()

	// TODO: If URL contains any information other than hostname:port, then return invalid URL error

	// Connect to mongodb using environment variables
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	mongoConnectURI := fmt.Sprintf(mURI.String())
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnectURI))

	// TODO: Allow SSL config

	return client, nil
}
