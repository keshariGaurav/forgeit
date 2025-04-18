package configs

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
  	clientOpts := options.Client().ApplyURI(EnvMongoURI()).SetServerAPIOptions(serverAPI)

    client, err := mongo.Connect(ctx, clientOpts)
    if err != nil {
        log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
    }

    // Ping the database
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatalf("❌ MongoDB ping failed: %v", err)
    }

    fmt.Println("✅ Connected to MongoDB")
    return client
}

// Global DB client instance
var DB *mongo.Client = ConnectDB()

// GetCollection returns a MongoDB collection by name
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
    return client.Database("golangAPI").Collection(collectionName)
}
