// account-admin adds or updates a NEX password record for an existing PID.
 // It does not create a Pretendo Network ID or change the PID assigned to a console.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func main() {
	pid := flag.Uint64("pid", 0, "existing Wii Sports Club/NEX PID to provision (required)")
	flag.Parse()
	if *pid == 0 || *pid > 0xFFFFFFFF {
		log.Fatal("-pid must be an existing non-zero 32-bit PID")
	}

	password := os.Getenv("WSC_NEX_PASSWORD")
	if strings.TrimSpace(password) == "" {
		log.Fatal("set WSC_NEX_PASSWORD in the environment; do not pass passwords as command-line arguments")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	uri := envOrDefault("MONGO_URI", "mongodb://localhost:27017/")
	database := envOrDefault("NEX_ACCOUNTS_DATABASE", "wsc")
	collection := envOrDefault("NEX_ACCOUNTS_COLLECTION", "nexaccounts")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("create MongoDB client: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("connect to MongoDB: %v", err)
	}

	coll := client.Database(database).Collection(collection)
	_, err = coll.UpdateOne(
		ctx,
		bson.D{{Key: "pid", Value: uint32(*pid)}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "pid", Value: uint32(*pid)},
			{Key: "password", Value: password},
		}}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Fatalf("save NEX account record: %v", err)
	}

	fmt.Printf("Provisioned NEX password record for PID %d in %s.%s. Password not displayed.\n", *pid, database, collection)
}
