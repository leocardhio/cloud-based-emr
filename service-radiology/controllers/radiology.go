package fasyankes_controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	specialityexamination "service-radiology/datastruct/outpatient"
	"service-radiology/datastruct/radiology"
	"service-radiology/datastruct/user"
	"service-radiology/db/csfle"
	"service-radiology/logger"
	"service-radiology/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RadiologyController struct {
	FaskesCollection  *mongo.Collection
	ConsentCollection *mongo.Collection

	ClientEncryption *mongo.ClientEncryption
	EncryptionOpts   *options.EncryptOptions
}

func InitRadiologyController(client *mongo.Client, csfle *csfle.CSFLE) *RadiologyController {
	return &RadiologyController{
		FaskesCollection:  client.Database("fasyankes").Collection("radiologi"),
		ConsentCollection: client.Database("emr").Collection("consent"),

		ClientEncryption: csfle.ClientEncryption,
		EncryptionOpts:   options.Encrypt().SetKeyID(*csfle.DEK), //
	}
}

func (rc *RadiologyController) GetPatientConsent(noihs string) (*user.PatientConsent, error) {
	filter := bson.M{}

	filter["no_ihs"] = noihs

	var result user.PatientConsent
	err := rc.ConsentCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (radiologyController *RadiologyController) GetRadiologyDataById() gin.HandlerFunc {
	return func(c *gin.Context) {
		objid, err := primitive.ObjectIDFromHex(c.Param("Id"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"_id": objid}
		if !c.GetBool("patientConsent") {
			filter["client_id"] = c.GetString("userClient")
		}

		var radiologyrequest specialityexamination.RadiologyRequest
		err = radiologyController.FaskesCollection.FindOne(context.Background(), filter).Decode(&radiologyrequest)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				if c.GetBool("patientConsent") {
					utils.JSON(c, http.StatusNotFound, gin.H{"error": "Data not found"})
					return
				} else {
					utils.JSON(c, http.StatusUnauthorized, gin.H{"error": user.NotAuthorizedError.Error()})
					return
				}
			} else {
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		id := radiologyrequest.ID
		signature := radiologyrequest.Signature
		radiologyrequest.Signature = nil
		radiologyrequest.ID = primitive.NilObjectID

		dataByte, err := json.Marshal(radiologyrequest)
		if err != nil {
			logger.LogPanic.Panicf("Failed to marshal json data")
		}

		_, err = utils.VerifySignature(string(dataByte), *signature)
		if err != nil {
			logger.LogWarning.Printf("Data with ID [%s] was tampered\n", radiologyrequest.ID.Hex())
		}

		utils.Decrypt(
			radiologyrequest.ConfidentialEncrypted,
			radiologyController.ClientEncryption,
		).Unmarshal(&radiologyrequest.ConfidentialData)

		radiologyrequest.ConfidentialEncrypted = nil
		radiologyrequest.Signature = signature
		radiologyrequest.ID = id

		utils.JSON(c, http.StatusOK, radiologyrequest)
	}
}

func (radiologyController *RadiologyController) CreateRadiologyRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var radiologyrequest specialityexamination.RadiologyRequest
		if err := c.ShouldBindJSON(&radiologyrequest); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		radiologyrequest.CreatedAt = &now
		radiologyrequest.UpdatedAt = &now

		confidentialEncryptedField := utils.EncryptRandom(
			radiologyrequest.ConfidentialData,
			radiologyController.ClientEncryption,
			radiologyController.EncryptionOpts,
		)

		radiologyrequest.ConfidentialEncrypted = confidentialEncryptedField
		radiologyrequest.ConfidentialData = nil

		json, err := json.Marshal(radiologyrequest)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		radiologyrequest.Signature = &signature

		resultRadiologyRequest, err := radiologyController.FaskesCollection.InsertOne(context.Background(), radiologyrequest)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, resultRadiologyRequest.InsertedID.(primitive.ObjectID).Hex())
	}
}

func (radiologyController *RadiologyController) GetAllRadiologyDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		jenisPemeriksaan := c.Query("jenis_pemeriksaan")
		namaPemeriksaan := c.Query("nama_pemeriksaan")

		noIHS := c.Param("noIHS")
		filter := bson.M{}

		filter["no_ihs"] = noIHS

		if jenisPemeriksaan != "" {
			regex := primitive.Regex{
				Pattern: fmt.Sprintf(".*%s.*", jenisPemeriksaan),
				Options: "",
			}
			filter["jenis_pemeriksaan"] = regex
		}

		if namaPemeriksaan != "" {
			regex := primitive.Regex{
				Pattern: fmt.Sprintf(".*%s.*", namaPemeriksaan),
				Options: "",
			}
			filter["nama_pemeriksaan"] = regex
		}

		if !c.GetBool("patientConsent") {
			filter["$or"] = bson.A{
				bson.M{"client_id": c.GetString("userClient")},
				bson.M{"client_id": ""},
			}
		}

		// Query all radiology data
		cursor, err := radiologyController.FaskesCollection.Find(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var radiologyData []radiology.RadiologyData
		for cursor.Next(context.Background()) {
			var data radiology.RadiologyData
			if err := cursor.Decode(&data); err != nil {
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			signature := data.Signature
			id := data.ID
			data.Signature = nil
			data.ID = primitive.NilObjectID

			dataByte, err := json.Marshal(data)
			if err != nil {
				logger.LogPanic.Panicf("Failed to marshal json data")
			}

			_, err = utils.VerifySignature(string(dataByte), *signature)
			if err != nil {
				logger.LogWarning.Printf("Data with ID [%s] was tampered\n", id.Hex())
				continue
			}

			utils.Decrypt(
				data.ConfidentialEncrypted,
				radiologyController.ClientEncryption,
			).Unmarshal(&data.ConfidentialData)

			data.ConfidentialEncrypted = nil
			data.Signature = signature
			data.ID = id

			radiologyData = append(radiologyData, data)
		}

		if err := cursor.Err(); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, radiologyData)
	}
}

func (radiologyController *RadiologyController) CreateRadiologyDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var radiologydata radiology.RadiologyData

		// Bind the request body to the radiology.RadiologyData struct
		if err := c.ShouldBindJSON(&radiologydata); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		radiologydata.CreatedAt = &now
		radiologydata.UpdatedAt = &now

		radiologydata.ClientID = c.GetString("userClient")

		confidentialEncryptedField := utils.EncryptRandom(
			radiologydata.ConfidentialData,
			radiologyController.ClientEncryption,
			radiologyController.EncryptionOpts,
		)

		radiologydata.ConfidentialEncrypted = confidentialEncryptedField
		radiologydata.ConfidentialData = nil

		json, err := json.Marshal(radiologydata)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		radiologydata.Signature = &signature

		// Insert the new radiology data
		_, err = radiologyController.FaskesCollection.InsertOne(context.Background(), radiologydata)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusCreated, gin.H{"message": "Radiology data created successfully"})
	}
}

func (radiologyController *RadiologyController) UpdateRadiologyDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("Id"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		noIHS := c.Param("noIHS")
		var newData radiology.RadiologyData

		if !c.GetBool("patientConsent") {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"forbidden": user.NotAuthorizedError.Error()})
			return
		}

		// Bind the request body to the radiology.RadiologyData struct
		if err := c.ShouldBindJSON(&newData); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if newData.CreatedAt == nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": "require created_at data"})
			return
		}

		// Define a filter to find the document by noPermintaan
		filter := bson.M{
			"_id":    id,
			"no_ihs": noIHS,
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))
		newData.UpdatedAt = &now

		confidentialEncryptedField := utils.EncryptRandom(
			newData.ConfidentialData,
			radiologyController.ClientEncryption,
			radiologyController.EncryptionOpts,
		)

		newData.ConfidentialEncrypted = confidentialEncryptedField
		newData.ConfidentialData = nil

		newData.ClientID = c.GetString("userClient")
		json, err := json.Marshal(newData)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		newData.Signature = &signature

		// Create an update document
		update := bson.M{"$set": newData}

		// Update the document in the Collection
		result, err := radiologyController.FaskesCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			utils.JSON(c, http.StatusNotFound, gin.H{"error": "No data matched the parameter"})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d radiology data updated successfully", result.ModifiedCount)})
	}
}

func (radiologyController *RadiologyController) DeleteRadiologyDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("Id"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Define a filter to find the document by noPermintaan
		filter := bson.M{"_id": id}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		// Create an update document
		update := bson.M{"$set": bson.M{
			"deleted_at": now,
		}}

		// Update the document in the Collection
		_, err = radiologyController.FaskesCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the document from the Collection
		result, err := radiologyController.FaskesCollection.DeleteOne(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d radiology data deleted successfully", result.DeletedCount)})
	}
}
