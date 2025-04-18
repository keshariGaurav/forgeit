package configs

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName      = "golangAPI"
	connectionTimeout = 20 * time.Second
)

// ConnectDB establishes connection with MongoDB
func ConnectDB() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	// Configure MongoDB client
	fmt.Println("🔗 Using Mongo URI:", EnvMongoURI())
	clientOpts := options.Client().
		ApplyURI(EnvMongoURI()).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ MongoDB ping failed: %v", err)
	}

	log.Println("✅ Successfully connected to MongoDB")
	return client
}

// DB holds the MongoDB client
var DB *mongo.Client = ConnectDB()

// GetCollection returns a MongoDB collection
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database(databaseName).Collection(collectionName)
}
