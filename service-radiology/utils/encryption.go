package utils

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EncryptRandom(v any, ce *mongo.ClientEncryption, eopts *options.EncryptOptions) *primitive.Binary {
	eopts.SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Random")
	encryptRawValueType, encryptRawValueData, err := bson.MarshalValue(v)
	if err != nil {
		panic(fmt.Errorf("failed to marshal data %v", err))
	}

	encryptRawValue := bson.RawValue{Type: encryptRawValueType, Value: encryptRawValueData}
	encryptedField, err := ce.Encrypt(
		context.Background(),
		encryptRawValue,
		eopts,
	)
	if err != nil {
		panic(fmt.Errorf("failed to encrypt %v", err))
	}

	return &encryptedField
}

func EncryptDeterministic(v any, ce *mongo.ClientEncryption, eopts *options.EncryptOptions) *primitive.Binary {
	eopts.SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic")
	encryptRawValueType, encryptRawValueData, err := bson.MarshalValue(v)
	if err != nil {
		panic(fmt.Errorf("failed to marshal data %v", err))
	}

	encryptRawValue := bson.RawValue{Type: encryptRawValueType, Value: encryptRawValueData}
	encryptedField, err := ce.Encrypt(
		context.Background(),
		encryptRawValue,
		eopts,
	)
	if err != nil {
		panic(fmt.Errorf("failed to encrypt %v", err))
	}

	return &encryptedField
}

func Decrypt(encryptedVal *primitive.Binary, ce *mongo.ClientEncryption) *bson.RawValue {
	valDecrypted, err := ce.Decrypt(
		context.Background(),
		*encryptedVal,
	)
	if err != nil {
		panic(fmt.Errorf("failed to decrypt %v", err))
	}

	return &valDecrypted
}
