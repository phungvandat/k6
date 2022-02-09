package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func OpenMongoDBConn(ctx context.Context, uri, dbName string) (*mongo.Database, func()) {
	var (
		opts        = options.Client().ApplyURI(uri)
		client, err = mongo.Connect(ctx,
			opts,
			options.Client().SetMaxPoolSize(100),
			options.Client().SetMaxConnIdleTime(10*time.Second),
		)
	)

	if err != nil {
		panic(err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	log.Println("db connected")

	db := client.Database(dbName)

	return db, func() {
		disCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		err = client.Disconnect(disCtx)
		if err != nil {
			log.Printf("failed close db %v\n", err)
			return
		}

		log.Println("db closed")
	}
}
