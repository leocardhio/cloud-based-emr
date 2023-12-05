package db

import (
	"context"
	"fmt"
	"service-radiology/config"
	"service-radiology/logger"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func ConnectDB(cfg *config.Config) *mongo.Client {
	// Connect to MongoDB Atlas
	mongoUri := fmt.Sprintf("mongodb+srv://%s:%s@%s/", cfg.DBUser, cfg.DBPassword, cfg.DBClusterURL)
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoUri).
		SetReadPreference(readpref.SecondaryPreferred(nil)))
	if err != nil {
		logger.LogFatal.Fatalf("Connect error for regular client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ping the MongoDB connection
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.LogPanic.Panicf("failed to ping mongo connection: %v", err)
	}

	logger.LogInfo.Println("Connected to Regular MongoDB successfully")

	return client
}

func DisconnectDB(client *mongo.Client) {
	err := client.Disconnect(context.Background())
	if err != nil {
		logger.LogFatal.Fatalf("failed to disconnect db: %v", err)
	}

	logger.LogInfo.Println("DB successfully disconnected")
}
