package emr_controllers

import (
	"context"
	"fmt"
	"net/http"
	"service-outpatient/datastruct/outpatient/identity"
	"service-outpatient/db/csfle"
	"service-outpatient/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserIdentityController struct {
	Collection *mongo.Collection

	ClientEncryption *mongo.ClientEncryption
	EncryptionOpts   *options.EncryptOptions
}

func InitUserIdentityController(client *mongo.Client, csfle *csfle.CSFLE) *UserIdentityController {
	return &UserIdentityController{
		Collection:       client.Database("emr").Collection("identitas"),
		ClientEncryption: csfle.ClientEncryption,
		EncryptionOpts:   options.Encrypt().SetKeyID(*csfle.DEK),
	}
}

func (uic UserIdentityController) GetAllUserIdentityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		noIHS := c.Query("no_ihs")
		nik := c.Query("nik")
		identitasLain := c.Query("identitas_lain")

		filter := bson.M{}

		if noIHS != "" {
			filter["no_ihs"] = noIHS
		}

		if nik != "" {
			filter["encrypted_nik"] = utils.EncryptDeterministic(
				nik,
				uic.ClientEncryption,
				uic.EncryptionOpts,
			)
		}

		if identitasLain != "" {
			filter["encrypted_identitas_lain"] = utils.EncryptDeterministic(
				identitasLain,
				uic.ClientEncryption,
				uic.EncryptionOpts,
			)
		}

		// Query all outpatient data
		cursor, err := uic.Collection.Find(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var outpatientIdentityData []identity.AdultPatient
		for cursor.Next(context.Background()) {
			var data identity.AdultPatient
			if err := cursor.Decode(&data); err != nil {
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			utils.Decrypt(
				data.ConfidentialEncrypted,
				uic.ClientEncryption,
			).Unmarshal(&data.ConfidentialData)

			utils.Decrypt(
				data.NIKEncrypted,
				uic.ClientEncryption,
			).Unmarshal(&data.NIK)

			utils.Decrypt(
				data.NamaEncrypted,
				uic.ClientEncryption,
			).Unmarshal(&data.NamaLengkap)

			utils.Decrypt(
				data.IdentitasLainEncrypted,
				uic.ClientEncryption,
			).Unmarshal(&data.IdentitasLain)

			data.ConfidentialEncrypted = nil
			data.NIKEncrypted = nil
			data.NamaEncrypted = nil
			data.IdentitasLainEncrypted = nil

			outpatientIdentityData = append(outpatientIdentityData, data)
		}

		if err := cursor.Err(); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, outpatientIdentityData)
	}
}

func (uic UserIdentityController) CreateUserIdentityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data identity.AdultPatient

		// Bind the request body to the outpatient.OutpatientAdult struct
		if err := c.ShouldBindJSON(&data); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		data.CreatedAt = &now
		data.UpdatedAt = &now

		confidentialEncryptedField := utils.EncryptRandom(
			data.ConfidentialData,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		nameEncryptedField := utils.EncryptDeterministic(
			data.NamaLengkap,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			data.NIK,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		identitasLainEncryptedField := utils.EncryptDeterministic(
			data.IdentitasLain,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		data.ConfidentialEncrypted = confidentialEncryptedField
		data.ConfidentialData = nil

		data.NamaEncrypted = nameEncryptedField
		data.NamaLengkap = nil

		data.NIKEncrypted = nikEncryptedField
		data.NIK = nil

		data.IdentitasLainEncrypted = identitasLainEncryptedField
		data.IdentitasLain = nil

		// Insert the new outpatient data
		_, err := uic.Collection.InsertOne(context.Background(), data)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusCreated, gin.H{"message": "Outpatient identity data created successfully"})
	}
}

func (uic UserIdentityController) UpdateUserIdentityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		noIHS := c.Param("noIHS")
		var newData identity.AdultPatient

		// Bind the request body to the outpatient.OutpatientAdult struct
		if err := c.ShouldBindJSON(&newData); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if newData.CreatedAt == nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": "require created_at data"})
			return
		}

		// Define a filter to find the document by NoIHS
		filter := bson.M{"no_ihs": noIHS}

		now := time.Now().Truncate(time.Duration(time.Millisecond))
		newData.UpdatedAt = &now

		encryptedField := utils.EncryptRandom(
			newData.ConfidentialData,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		nameEncryptedField := utils.EncryptDeterministic(
			newData.NamaLengkap,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			newData.NIK,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		identitasLainEncryptedField := utils.EncryptDeterministic(
			newData.IdentitasLain,
			uic.ClientEncryption,
			uic.EncryptionOpts,
		)

		newData.ConfidentialEncrypted = encryptedField
		newData.ConfidentialData = nil

		newData.NamaEncrypted = nameEncryptedField
		newData.NamaLengkap = nil

		newData.NIKEncrypted = nikEncryptedField
		newData.NIK = nil

		newData.IdentitasLainEncrypted = identitasLainEncryptedField
		newData.IdentitasLain = nil

		// Create an update document
		update := bson.M{"$set": newData}

		// Update the document in the collection
		result, err := uic.Collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			utils.JSON(c, http.StatusNotFound, gin.H{"error": "No data matched the parameter"})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d outpatient identity data updated successfully", result.ModifiedCount)})
	}
}

func (uic UserIdentityController) DeleteUserIdentityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		noIHS := c.Param("noIHS")

		// Define a filter to find the document by noPermintaan
		filter := bson.M{"no_ihs": noIHS}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		// Create an update document
		update := bson.M{"$set": bson.M{
			"deleted_at": now,
		}}

		// Update the document in the collection
		_, err := uic.Collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the document from the collection
		result, err := uic.Collection.DeleteOne(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d outpatient identity data deleted successfully", result.DeletedCount)})
	}
}
