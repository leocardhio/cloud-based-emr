package fasyankes_controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"service-lab/datastruct/laboratory"
	specialityexamination "service-lab/datastruct/outpatient"
	"service-lab/datastruct/user"
	"service-lab/db/csfle"
	"service-lab/logger"
	"service-lab/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LabController struct {
	FaskesCollection  *mongo.Collection
	ConsentCollection *mongo.Collection

	ClientEncryption *mongo.ClientEncryption
	EncryptionOpts   *options.EncryptOptions
}

func InitLabController(client *mongo.Client, csfle *csfle.CSFLE) *LabController {
	return &LabController{
		FaskesCollection:  client.Database("fasyankes").Collection("laboratorium"),
		ConsentCollection: client.Database("emr").Collection("consent"),

		ClientEncryption: csfle.ClientEncryption,
		EncryptionOpts:   options.Encrypt().SetKeyID(*csfle.DEK),
	}
}

func (lc *LabController) GetPatientConsent(noihs string) (*user.PatientConsent, error) {
	filter := bson.M{}

	filter["no_ihs"] = noihs

	var result user.PatientConsent
	err := lc.ConsentCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (labController *LabController) GetLabDataById() gin.HandlerFunc {
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

		var labrequest specialityexamination.LaboratoryRequest
		err = labController.FaskesCollection.FindOne(context.Background(), filter).Decode(&labrequest)
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

		id := labrequest.ID
		signature := labrequest.Signature
		labrequest.Signature = nil
		labrequest.ID = primitive.NilObjectID

		dataByte, err := json.Marshal(labrequest)
		if err != nil {
			logger.LogPanic.Panicf("Failed to marshal json data")
		}

		_, err = utils.VerifySignature(string(dataByte), *signature)
		if err != nil {
			logger.LogWarning.Printf("Data with ID [%s] was tampered\n", labrequest.ID.Hex())
		}

		utils.Decrypt(
			labrequest.ConfidentialEncrypted,
			labController.ClientEncryption,
		).Unmarshal(&labrequest.ConfidentialData)

		utils.Decrypt(
			labrequest.NIKEncrypted,
			labController.ClientEncryption,
		).Unmarshal(&labrequest.NIK)

		labrequest.ConfidentialEncrypted = nil
		labrequest.NIKEncrypted = nil
		labrequest.Signature = signature
		labrequest.ID = id

		utils.JSON(c, http.StatusOK, labrequest)
	}
}

func (labController *LabController) CreateLabRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var labrequest specialityexamination.LaboratoryRequest
		if err := c.ShouldBindJSON(&labrequest); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		labrequest.CreatedAt = &now
		labrequest.UpdatedAt = &now

		confidentialEncryptedField := utils.EncryptRandom(
			labrequest.ConfidentialData,
			labController.ClientEncryption,
			labController.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			labrequest.NIK,
			labController.ClientEncryption,
			labController.EncryptionOpts,
		)

		labrequest.ConfidentialEncrypted = confidentialEncryptedField
		labrequest.ConfidentialData = nil

		labrequest.NIKEncrypted = nikEncryptedField
		labrequest.NIK = nil

		json, err := json.Marshal(labrequest)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		labrequest.Signature = &signature

		resultLabRequest, err := labController.FaskesCollection.InsertOne(context.Background(), labrequest)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, resultLabRequest.InsertedID.(primitive.ObjectID).Hex())
	}
}

func (labController *LabController) GetAllLabDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		namaPemeriksaan := c.Query("nama_pemeriksaan")
		noIHS := c.Param("noIHS")
		noRegLab := c.Query("no_registrasi_lab")
		nik := c.Query("nik")

		filter := bson.M{}
		filter["no_ihs"] = noIHS

		if namaPemeriksaan != "" {
			regex := primitive.Regex{
				Pattern: fmt.Sprintf(".*%s.*", namaPemeriksaan),
				Options: "",
			}

			filter["nama_pemeriksaan"] = regex
		}

		if noRegLab != "" {
			filter["no_registrasi_lab"] = noRegLab
		}

		if nik != "" {
			filter["encrypted_nik"] = utils.EncryptDeterministic(
				nik,
				labController.ClientEncryption,
				labController.EncryptionOpts,
			)
		}

		if !c.GetBool("patientConsent") {
			filter["$or"] = bson.A{
				bson.M{"client_id": c.GetString("userClient")},
				bson.M{"client_id": ""},
			}
		}

		// Query all laboratory data
		cursor, err := labController.FaskesCollection.Find(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var labData []laboratory.LaboratoryData
		for cursor.Next(context.Background()) {
			var data laboratory.LaboratoryData
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
				logger.LogPanic.Panicf("Failed to marshall json data")
			}

			_, err = utils.VerifySignature(string(dataByte), *signature)
			if err != nil {
				logger.LogWarning.Printf("Data with ID [%s] was tampered\n", id.Hex())
				continue
			}

			utils.Decrypt(
				data.ConfidentialEncrypted,
				labController.ClientEncryption,
			).Unmarshal(&data.ConfidentialData)

			utils.Decrypt(
				data.NIKEncrypted,
				labController.ClientEncryption,
			).Unmarshal(&data.NIK)

			data.ConfidentialEncrypted = nil
			data.NIKEncrypted = nil
			data.Signature = signature
			data.ID = id

			labData = append(labData, data)
		}

		if err := cursor.Err(); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, labData)
	}
}

func (labController *LabController) CreateLabDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var labdata laboratory.LaboratoryData

		// Bind the request body to the laboratory.LaboratoryData struct
		if err := c.ShouldBindJSON(&labdata); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		labdata.CreatedAt = &now
		labdata.UpdatedAt = &now

		labdata.ClientID = c.GetString("userClient")

		confidentialEncryptedField := utils.EncryptRandom(
			labdata.ConfidentialData,
			labController.ClientEncryption,
			labController.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			labdata.NIK,
			labController.ClientEncryption,
			labController.EncryptionOpts,
		)

		labdata.ConfidentialEncrypted = confidentialEncryptedField
		labdata.ConfidentialData = nil

		labdata.NIKEncrypted = nikEncryptedField
		labdata.NIK = nil

		json, err := json.Marshal(labdata)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		labdata.Signature = &signature

		// Insert the new laboratory data
		_, err = labController.FaskesCollection.InsertOne(context.Background(), labdata)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusCreated, gin.H{"message": "Laboratory data created successfully"})
	}
}

func (labController *LabController) UpdateLabDataHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("Id"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		noIHS := c.Param("noIHS")
		var newData laboratory.LaboratoryData

		if !c.GetBool("patientConsent") {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"forbidden": user.NotAuthorizedError.Error()})
			return
		}

		// Bind the request body to the laboratory.LaboratoryData struct
		if err = c.ShouldBindJSON(&newData); err != nil {
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
			labController.ClientEncryption,
			labController.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			newData.NIK,
			labController.ClientEncryption,
			labController.EncryptionOpts,
		)

		newData.ConfidentialEncrypted = confidentialEncryptedField
		newData.ConfidentialData = nil

		newData.NIKEncrypted = nikEncryptedField
		newData.NIK = nil

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

		// Update the document in the FaskesCollection
		result, err := labController.FaskesCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			utils.JSON(c, http.StatusNotFound, gin.H{"error": "No data matched the parameter"})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d laboratory data updated successfully", result.ModifiedCount)})
	}
}

func (labController *LabController) DeleteLabDataHandler() gin.HandlerFunc {
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

		// Update the document in the FaskesCollection
		_, err = labController.FaskesCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the document from the FaskesCollection
		result, err := labController.FaskesCollection.DeleteOne(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d laboratory data deleted successfully", result.DeletedCount)})
	}
}
