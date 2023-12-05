package fasyankes_controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	specialityexamination "service-pharmacy/datastruct/outpatient"
	"service-pharmacy/datastruct/pharmacy"
	"service-pharmacy/datastruct/user"
	"service-pharmacy/db/csfle"
	"service-pharmacy/logger"
	"service-pharmacy/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PharmacyController struct {
	FaskesCollection  *mongo.Collection
	ConsentCollection *mongo.Collection

	ClientEncryption *mongo.ClientEncryption
	EncryptionOpts   *options.EncryptOptions
}

func InitPharmacyController(client *mongo.Client, csfle *csfle.CSFLE) *PharmacyController {
	return &PharmacyController{
		FaskesCollection:  client.Database("fasyankes").Collection("apotek"),
		ConsentCollection: client.Database("emr").Collection("consent"),

		ClientEncryption: csfle.ClientEncryption,
		EncryptionOpts:   options.Encrypt().SetKeyID(*csfle.DEK),
	}
}

func (pc *PharmacyController) GetPatientConsent(noihs string) (*user.PatientConsent, error) {
	filter := bson.M{}

	filter["no_ihs"] = noihs

	var result user.PatientConsent
	err := pc.ConsentCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (pharmacyController *PharmacyController) GetPharmacyDataById() gin.HandlerFunc {
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

		var pharmacyrequest specialityexamination.PharmacyRequestDocument
		err = pharmacyController.FaskesCollection.FindOne(context.Background(), filter).Decode(&pharmacyrequest)
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

		signature := pharmacyrequest.Signature
		id := pharmacyrequest.ID
		pharmacyrequest.Signature = nil
		pharmacyrequest.ID = primitive.NilObjectID

		dataByte, err := json.Marshal(pharmacyrequest)
		if err != nil {
			logger.LogPanic.Panicf("Failed to marshal json data")
		}

		_, err = utils.VerifySignature(string(dataByte), *signature)
		if err != nil {
			logger.LogWarning.Printf("Data with ID [%s] was tampered\n", pharmacyrequest.ID)
		}

		utils.Decrypt(
			pharmacyrequest.Peresepan.ConfidentialEncrypted,
			pharmacyController.ClientEncryption,
		).Unmarshal(&pharmacyrequest.Peresepan.ConfidentialData)

		utils.Decrypt(
			pharmacyrequest.Peresepan.NIKEncrypted,
			pharmacyController.ClientEncryption,
		).Unmarshal(&pharmacyrequest.Peresepan.NIK)

		pharmacyrequest.Peresepan.ConfidentialEncrypted = nil
		pharmacyrequest.Peresepan.NIKEncrypted = nil
		pharmacyrequest.Signature = signature
		pharmacyrequest.ID = id

		utils.JSON(c, http.StatusOK, pharmacyrequest)
	}
}

func (pharmacyController *PharmacyController) CreatePharmacyRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pharmacyrequest specialityexamination.PharmacyRequestDocument
		if err := c.ShouldBindJSON(&pharmacyrequest); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		pharmacyrequest.CreatedAt = &now
		pharmacyrequest.UpdatedAt = &now

		confidentialEncryptedField := utils.EncryptRandom(
			pharmacyrequest.Peresepan.ConfidentialData,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			pharmacyrequest.Peresepan.NIK,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		pharmacyrequest.Peresepan.ConfidentialEncrypted = confidentialEncryptedField
		pharmacyrequest.Peresepan.ConfidentialData = nil

		pharmacyrequest.Peresepan.NIKEncrypted = nikEncryptedField
		pharmacyrequest.Peresepan.NIK = nil

		json, err := json.Marshal(pharmacyrequest)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		pharmacyrequest.Signature = &signature

		resultPharmacyRequest, err := pharmacyController.FaskesCollection.InsertOne(context.Background(), pharmacyrequest)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, resultPharmacyRequest.InsertedID.(primitive.ObjectID).Hex())
	}
}

func (pharmacyController *PharmacyController) GetAllPharmacyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		noRekamMedis := c.Query("no_rekam_medis")
		idPelanggan := c.Query("id_pelanggan")
		idObat := c.Query("id_obat")
		nik := c.Query("nik")

		noIHS := c.Param("noIHS")
		filter := bson.M{}

		filter["peresepan.no_ihs"] = noIHS

		if noRekamMedis != "" {
			filter["peresepan.no_rekam_medis"] = noRekamMedis
		}

		if idPelanggan != "" {
			filter["peresepan.id_pelanggan"] = idPelanggan
		}

		if idObat != "" {
			filter["peresepan.id_obat"] = idObat
		}

		if nik != "" {
			filter["encrypted_nik"] = utils.EncryptDeterministic(
				nik,
				pharmacyController.ClientEncryption,
				pharmacyController.EncryptionOpts,
			)
		}

		if !c.GetBool("patientConsent") {
			filter["$or"] = bson.A{
				bson.M{"client_id": c.GetString("userClient")},
				bson.M{"client_id": ""},
			}
		}

		// Query all pharmacy data
		cursor, err := pharmacyController.FaskesCollection.Find(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var pharmacyData []pharmacy.Pharmacy
		for cursor.Next(context.Background()) {
			var data pharmacy.Pharmacy
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
				logger.LogWarning.Printf("Data with ID [%s] was tampered\n", id)
				continue
			}

			if data.DispensingEncrypted != nil {
				utils.Decrypt(
					data.DispensingEncrypted,
					pharmacyController.ClientEncryption,
				).Unmarshal(&data.Dispensing)
			}

			utils.Decrypt(
				data.Peresepan.ConfidentialEncrypted,
				pharmacyController.ClientEncryption,
			).Unmarshal(&data.Peresepan.ConfidentialData)

			utils.Decrypt(
				data.Peresepan.NIKEncrypted,
				pharmacyController.ClientEncryption,
			).Unmarshal(&data.Peresepan.NIK)

			data.DispensingEncrypted = nil
			data.Peresepan.ConfidentialEncrypted = nil
			data.Peresepan.NIKEncrypted = nil
			data.Signature = signature
			data.ID = id

			pharmacyData = append(pharmacyData, data)
		}

		if err := cursor.Err(); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, pharmacyData)
	}
}

func (pharmacyController *PharmacyController) CreatePharmacyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data pharmacy.Pharmacy

		// Bind the request body to the pharmacy.Pharmacy struct
		if err := c.ShouldBindJSON(&data); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		data.CreatedAt = &now
		data.UpdatedAt = &now

		data.ClientID = c.GetString("userClient")

		dispensingEncryptedField := utils.EncryptRandom(
			data.Dispensing,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		confidentialEncryptedField := utils.EncryptRandom(
			data.Peresepan.ConfidentialData,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			data.Peresepan.NIK,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		data.DispensingEncrypted = dispensingEncryptedField
		data.Dispensing = nil

		data.Peresepan.ConfidentialEncrypted = confidentialEncryptedField
		data.Peresepan.ConfidentialData = nil

		data.Peresepan.NIKEncrypted = nikEncryptedField
		data.Peresepan.NIK = nil

		json, err := json.Marshal(data)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		data.Signature = &signature

		// Insert the new pharmacy data
		_, err = pharmacyController.FaskesCollection.InsertOne(context.Background(), data)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusCreated, gin.H{"message": "Pharmacy data created successfully"})
	}
}

func (pharmacyController *PharmacyController) UpdatePharmacyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("Id"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		noIHS := c.Param("noIHS")
		var newData pharmacy.Pharmacy

		if !c.GetBool("patientConsent") {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"forbidden": user.NotAuthorizedError.Error()})
			return
		}

		// Bind the request body to the pharmacy.Pharmacy struct
		if err := c.ShouldBindJSON(&newData); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if newData.CreatedAt == nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": "require created_at data"})
			return
		}

		// Define a filter to find the document by idResep
		filter := bson.M{
			"_id":              id,
			"peresepan.no_ihs": noIHS,
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))
		newData.UpdatedAt = &now

		dispensingEncryptedField := utils.EncryptRandom(
			newData.Dispensing,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		confidentialEncryptedField := utils.EncryptRandom(
			newData.Peresepan.ConfidentialData,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		nikEncryptedField := utils.EncryptDeterministic(
			newData.Peresepan.NIK,
			pharmacyController.ClientEncryption,
			pharmacyController.EncryptionOpts,
		)

		newData.DispensingEncrypted = dispensingEncryptedField
		newData.Dispensing = nil

		newData.Peresepan.ConfidentialEncrypted = confidentialEncryptedField
		newData.Peresepan.ConfidentialData = nil

		newData.Peresepan.NIKEncrypted = nikEncryptedField
		newData.Peresepan.NIK = nil

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

		// Update the document in the collection
		result, err := pharmacyController.FaskesCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			utils.JSON(c, http.StatusNotFound, gin.H{"error": "No data matched the parameter"})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d pharmacy data updated successfully", result.ModifiedCount)})
	}
}

func (pharmacyController *PharmacyController) DeletePharmacyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("Id"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Define a filter to find the document by idResep
		filter := bson.M{
			"_id": id}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		// Create an update document
		update := bson.M{"$set": bson.M{
			"deleted_at": now,
		}}

		// Update the document in the collection
		_, err = pharmacyController.FaskesCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the document from the collection
		result, err := pharmacyController.FaskesCollection.DeleteOne(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d pharmacy data deleted successfully", result.DeletedCount)})
	}
}
