package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client
var mongoDatabase *mongo.Database
var mongoCollection *mongo.Collection

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func connectMongo() error {
	uri := envOrDefault("MONGO_URI", "mongodb://localhost:27017/")
	databaseName := envOrDefault("NEX_ACCOUNTS_DATABASE", "wsc")
	collectionName := envOrDefault("NEX_ACCOUNTS_COLLECTION", "nexaccounts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("create MongoDB client: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return fmt.Errorf("connect to MongoDB: %w", err)
	}

	mongoClient = client
	mongoDatabase = client.Database(databaseName)
	mongoCollection = mongoDatabase.Collection(collectionName)
	// A PID must map to exactly one persistent NEX password. Without this
	// unique index, concurrent first logins could create conflicting records.
	_, err = mongoCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "pid", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("nexaccounts_pid_unique"),
	})
	if err != nil {
		_ = client.Disconnect(context.Background())
		mongoClient = nil
		mongoDatabase = nil
		mongoCollection = nil
		return fmt.Errorf("ensure unique NEX PID index (resolve any duplicate PID records before startup): %w", err)
	}
	return nil
}

func getNEXAccountByPID(pid uint32) bson.M {
	if mongoCollection == nil {
		logger.Warning("NEX account collection is not initialized")
		return nil
	}

	var result bson.M
	err := mongoCollection.FindOne(
		context.Background(),
		bson.D{{Key: "pid", Value: pid}},
		options.FindOne(),
	).Decode(&result)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			logger.Warning(fmt.Sprintf("Could not look up NEX account PID %d: %v", pid, err))
		}
		return nil
	}

	return result
}
