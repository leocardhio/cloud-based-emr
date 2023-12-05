package csfle

import (
	"context"
	"fmt"
	"service-auth/config"
	"service-auth/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CSFLE struct {
	KeyVaultClient   *mongo.Client
	ClientEncryption *mongo.ClientEncryption
	DEK              *primitive.Binary
	AltKeyName       string

	Provider    string
	KMSProvider map[string]map[string]interface{}
	MasterKey   map[string]interface{}
}

func InitCSFLE(cfg *config.Config, keyVaultClient *mongo.Client) *CSFLE {
	return &CSFLE{
		KeyVaultClient: keyVaultClient,
		AltKeyName:     fmt.Sprintf("%s.%s", cfg.KMSKeyRing, cfg.KMSKeyName),
		Provider:       "gcp",
		KMSProvider: map[string]map[string]interface{}{
			"gcp": {
				"email":      cfg.SAEmail,
				"privateKey": cfg.SAPrivateKey,
			},
		},
		MasterKey: map[string]interface{}{
			"projectId": cfg.KMSProjectId,
			"location":  cfg.KMSLocation,
			"keyRing":   cfg.KMSKeyRing,
			"keyName":   cfg.KMSKeyName,
		},
	}
}

func (csfle *CSFLE) CreateClientEncryption(keyVaultNamespace string) *CSFLE {
	clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).
		SetKmsProviders(csfle.KMSProvider)
	clientEnc, err := mongo.NewClientEncryption(csfle.KeyVaultClient, clientEncryptionOpts)
	if err != nil {
		logger.LogPanic.Panicln(fmt.Errorf("NewClientEncryption error: %v", err))
	}

	csfle.ClientEncryption = clientEnc
	return csfle
}

func (csfle *CSFLE) CloseClient() {
	if err := csfle.ClientEncryption.Close(context.Background()); err != nil {
		logger.LogPanic.Panicf("failed to close client encryption: %v", err)
	}
	logger.LogInfo.Println("Client closed")
}

func (csfle *CSFLE) MakeKey() error {
	// start-create-dek
	dataKeyOpts := options.DataKey().
		SetMasterKey(csfle.MasterKey).
		SetKeyAltNames([]string{csfle.AltKeyName})

	dataKeyID, err := csfle.ClientEncryption.
		CreateDataKey(context.TODO(), csfle.Provider, dataKeyOpts)
	if err != nil {
		return err
	}

	csfle.DEK = &dataKeyID
	// end-create-dek

	return nil
}

func (csfle *CSFLE) GetKey() error {
	var result bson.M

	if err := csfle.ClientEncryption.GetKeyByAltName(context.Background(), csfle.AltKeyName).
		Decode(&result); err != nil {
		return err
	}

	dataKeyID := result["_id"].(primitive.Binary)

	csfle.DEK = &dataKeyID

	return nil
}
