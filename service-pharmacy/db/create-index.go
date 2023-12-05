package db

import (
	"context"
	"fmt"
	"service-pharmacy/logger"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateKeyVaultIndex(keyVaultClient *mongo.Client, keyVaultNamespace string) error {
	keyVaultComp := strings.Split(keyVaultNamespace, ".")
	keyVaultDb := keyVaultComp[0]
	keyVaultColl := keyVaultComp[1]

	logger.LogInfo.Println("Checking index for key vault collection...")

	cursor, err := keyVaultClient.Database(keyVaultDb).Collection(keyVaultColl).Indexes().List(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get index list: %v", err)
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return fmt.Errorf("failed to decode index list: %v", err)
	}

	if results != nil {
		logger.LogInfo.Println("Index has already exist")
		return nil
	}

	logger.LogInfo.Println("Create index for key vault collection")

	keyVaultIndex := mongo.IndexModel{
		Keys: bson.D{{"keyAltNames", 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.D{
				{"keyAltNames", bson.D{
					{"$exists", true},
				}},
			}),
	}

	_, err = keyVaultClient.Database(keyVaultDb).Collection(keyVaultColl).Indexes().CreateOne(context.TODO(), keyVaultIndex)
	if err != nil {
		logger.LogPanic.Panicf("failed to create key vault index: %v", err)
	}
	return nil

}
