package emr_controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"service-outpatient/datastruct"
	"service-outpatient/datastruct/outpatient"
	specialityexamination "service-outpatient/datastruct/outpatient/speciality-examination"
	"service-outpatient/datastruct/user"
	"service-outpatient/db/csfle"
	"service-outpatient/logger"
	"service-outpatient/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OutpatientExaminationController struct {
	ExaminationCollection *mongo.Collection
	ObatCollection        *mongo.Collection
	LabCollection         *mongo.Collection
	RadiologiCollection   *mongo.Collection
	ConsentCollection     *mongo.Collection

	ClientEncryption *mongo.ClientEncryption
	EncryptionOpts   *options.EncryptOptions
}

type ErrorMapper struct {
	Error string `json:"error" bson:"error"`
}

func InitOutpatientExaminationController(client *mongo.Client, csfle *csfle.CSFLE) *OutpatientExaminationController {
	return &OutpatientExaminationController{
		ExaminationCollection: client.Database("emr").Collection("pemeriksaan"),
		ObatCollection:        client.Database("fasyankes").Collection("apotek"),
		LabCollection:         client.Database("fasyankes").Collection("laboratorium"),
		RadiologiCollection:   client.Database("fasyankes").Collection("radiologi"),
		ConsentCollection:     client.Database("emr").Collection("consent"),

		ClientEncryption: csfle.ClientEncryption,
		EncryptionOpts:   options.Encrypt().SetKeyID(*csfle.DEK),
	}
}

func (oic *OutpatientExaminationController) GetPatientConsent(noihs string) (*user.PatientConsent, error) {
	filter := bson.M{}

	filter["no_ihs"] = noihs

	var result user.PatientConsent
	err := oic.ConsentCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (oic *OutpatientExaminationController) GetAllOutpatientExaminationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		noIHS := c.Param("noIHS")

		filterExamination := bson.M{"no_ihs": noIHS}
		if !c.GetBool("patientConsent") {
			filterExamination["client_id"] = c.GetString("userClient")
		}

		// Query all outpatient data
		cursor, err := oic.ExaminationCollection.Find(context.Background(), filterExamination)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var examinationDataList []outpatient.ExaminationDocument
		for cursor.Next(context.Background()) {
			var examinationdata outpatient.ExaminationDocument
			var obatdokumendata specialityexamination.PharmacyRequestDocument
			var labdata specialityexamination.LaboratoryRequest
			var radiologidata specialityexamination.RadiologyRequest

			if err := cursor.Decode(&examinationdata); err != nil {
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			id := examinationdata.ID
			signature := examinationdata.Signature
			examinationdata.Signature = nil
			examinationdata.ID = primitive.NilObjectID

			dataByte, err := json.Marshal(examinationdata)
			if err != nil {
				logger.LogPanic.Panicf("Failed to marshall json data")
			}

			_, err = utils.VerifySignature(string(dataByte), *signature)
			if err != nil {
				logger.LogWarning.Printf("Data with ID [%s] was tampered\n", id.Hex())
				continue
			}

			utils.Decrypt(
				examinationdata.ConfidentialEncrypted,
				oic.ClientEncryption,
			).Unmarshal(&examinationdata.ConfidentialData)

			examinationdata.ConfidentialEncrypted = nil

			if examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId != nil {
				var errorMapper ErrorMapper
				obatobjid := examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId

				g := utils.Getter{
					RefID:       *obatobjid,
					NoIHS:       noIHS,
					ServiceName: datastruct.PHARMACY,
				}
				respBody, err := g.GetRequest(c)
				if err != nil {
					utils.JSON(c, http.StatusInternalServerError, gin.H{"pharmacy service error": err.Error()})
					return
				}

				if err = json.Unmarshal(respBody, &errorMapper); err != nil {
					logger.LogError.Println("Failed unmarshal error pharmacy response")
					utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				if errorMapper.Error != "" {
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.HTTPResponseStatus = &errorMapper.Error
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat = nil
				} else {
					if err = json.Unmarshal(respBody, &obatdokumendata); err != nil {
						logger.LogError.Println("Failed unmarshal json pharmacy response")
						utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat = &obatdokumendata.Peresepan
				}

			}

			if examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId != nil {
				var errorMapper ErrorMapper
				labrefid := examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId

				g := utils.Getter{
					RefID:       *labrefid,
					NoIHS:       noIHS,
					ServiceName: datastruct.LABORATORY,
				}
				respBody, err := g.GetRequest(c)
				if err != nil {
					utils.JSON(c, http.StatusInternalServerError, gin.H{"lab service error": err.Error()})
					return
				}

				if err = json.Unmarshal(respBody, &errorMapper); err != nil {
					logger.LogError.Println("Failed unmarshal error lab response")
					utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				if errorMapper.Error != "" {
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabHTTPResponseStatus = &errorMapper.Error
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium = nil
				} else {
					if err = json.Unmarshal(respBody, &labdata); err != nil {
						logger.LogError.Println("Failed unmarshal json lab response")
						utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium = &labdata
				}

			}

			if examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId != nil {
				var errorMapper ErrorMapper
				radiologyrefid := examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId

				g := utils.Getter{
					RefID:       *radiologyrefid,
					NoIHS:       noIHS,
					ServiceName: datastruct.RADIOLOGY,
				}
				respBody, err := g.GetRequest(c)
				if err != nil {
					utils.JSON(c, http.StatusInternalServerError, gin.H{"radiology service error": err.Error()})
					return
				}

				if err = json.Unmarshal(respBody, &errorMapper); err != nil {
					logger.LogError.Println("Failed unmarshal error radiology response")
					utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				if errorMapper.Error != "" {
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiHTTPResponseStatus = &errorMapper.Error
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi = nil
				} else {
					if err = json.Unmarshal(respBody, &radiologidata); err != nil {
						logger.LogError.Println("Failed unmarshal json radiology response")
						utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi = &radiologidata
				}

			}
			examinationdata.Signature = signature
			examinationdata.ID = id

			examinationDataList = append(examinationDataList, examinationdata)

		}

		if err := cursor.Err(); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, examinationDataList)
	}
}

func (oic *OutpatientExaminationController) GetOutpatientExaminationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		noIHS := c.Param("noIHS")
		objID, err := primitive.ObjectIDFromHex(c.Param("objID"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var examinationdata outpatient.ExaminationDocument
		var obatdokumendata specialityexamination.PharmacyRequestDocument
		var labdata specialityexamination.LaboratoryRequest
		var radiologidata specialityexamination.RadiologyRequest

		filterExamination := bson.M{"_id": objID}

		// Query outpatient data
		cursor := oic.ExaminationCollection.FindOne(context.Background(), filterExamination)
		if err := cursor.Decode(&examinationdata); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.Decrypt(
			examinationdata.ConfidentialEncrypted,
			oic.ClientEncryption,
		).Unmarshal(&examinationdata.ConfidentialData)

		examinationdata.ConfidentialEncrypted = nil

		if examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId != nil {
			obatobjid := examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId

			g := utils.Getter{
				RefID:       *obatobjid,
				NoIHS:       noIHS,
				ServiceName: datastruct.PHARMACY,
			}
			respBody, err := g.GetRequest(c)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"pharmacy error": err.Error()})
				return
			}

			if err = json.Unmarshal(respBody, &obatdokumendata); err != nil {
				logger.LogError.Println("Failed unmarshal json pharmacy response")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat = &obatdokumendata.Peresepan
		}

		if examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId != nil {
			labrefid := examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId

			g := utils.Getter{
				RefID:       *labrefid,
				NoIHS:       noIHS,
				ServiceName: datastruct.LABORATORY,
			}
			respBody, err := g.GetRequest(c)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"laboratory error": err.Error()})
				return
			}

			if err = json.Unmarshal(respBody, &labdata); err != nil {
				logger.LogError.Println("Failed unmarshal json lab response")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium = &labdata
		}

		if examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId != nil {
			radiologyrefid := examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId

			g := utils.Getter{
				RefID:       *radiologyrefid,
				NoIHS:       noIHS,
				ServiceName: datastruct.RADIOLOGY,
			}
			respBody, err := g.GetRequest(c)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"radiology error": err.Error()})
				return
			}

			if err = json.Unmarshal(respBody, &radiologidata); err != nil {
				logger.LogError.Println("Failed unmarshal json radiology response")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi = &radiologidata
		}

		utils.JSON(c, http.StatusOK, examinationdata)
	}
}

func (oic *OutpatientExaminationController) CreateOutpatientExaminationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var examinationdata outpatient.ExaminationDocument
		var pharmacyrequestdata specialityexamination.PharmacyRequestDocument
		var labrequestdata specialityexamination.LaboratoryRequest
		var radiologirequestdata specialityexamination.RadiologyRequest

		// Bind the request body to the outpatient.OutpatientAdult struct
		if err := c.ShouldBindJSON(&examinationdata); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		drugreciperequestptr := examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat
		labrequestptr := examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium
		radiologirequestptr := examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		if drugreciperequestptr != nil {
			pharmacyrequestdata.Peresepan = *drugreciperequestptr
			examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat = nil

			pharmacyJson, err := json.Marshal(pharmacyrequestdata)
			if err != nil {
				logger.LogError.Println("Error marshalling pharmacy data")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sb, err := utils.PostRequest(c, datastruct.PHARMACY, pharmacyJson)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("pharmacy: %s", err.Error())})
				return
			}
			examinationdata.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId = &sb
		}

		if labrequestptr != nil {
			labrequestdata = *labrequestptr
			examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium = nil

			labJson, err := json.Marshal(labrequestdata)
			if err != nil {
				logger.LogError.Println("Error marshalling lab data")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sb, err := utils.PostRequest(c, datastruct.LABORATORY, labJson)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("lab: %s", err.Error())})
				return
			}

			examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId = &sb
		}

		if radiologirequestptr != nil {
			radiologirequestdata = *radiologirequestptr
			examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi = nil

			radiologiJson, err := json.Marshal(radiologirequestdata)
			if err != nil {
				logger.LogError.Println("Error marshalling radiology data")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sb, err := utils.PostRequest(c, datastruct.RADIOLOGY, radiologiJson)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("radiology: %s", err.Error())})
				return
			}
			examinationdata.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId = &sb
		}

		examinationdata.CreatedAt = &now
		examinationdata.UpdatedAt = &now

		confidentialEncryptedField := utils.EncryptRandom(
			examinationdata.ConfidentialData,
			oic.ClientEncryption,
			oic.EncryptionOpts,
		)

		examinationdata.ConfidentialEncrypted = confidentialEncryptedField
		examinationdata.ConfidentialData = nil
		examinationdata.ClientID = c.GetString("userClient")

		json, err := json.Marshal(examinationdata)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		examinationdata.Signature = &signature

		_, err = oic.ExaminationCollection.InsertOne(context.Background(), examinationdata)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusCreated, gin.H{"message": "Outpatient examination data created successfully"})
	}
}

func (oic *OutpatientExaminationController) UpdateOutpatientExaminationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		objID, err := primitive.ObjectIDFromHex(c.Param("objID"))
		noIHS := c.Param("noIHS")
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var newData outpatient.ExaminationDocument
		var pharmacyrequestdata specialityexamination.PharmacyRequestDocument
		var labrequestdata specialityexamination.LaboratoryRequest
		var radiologirequestdata specialityexamination.RadiologyRequest

		if !c.GetBool("patientConsent") {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"forbidden": user.NotAuthorizedError.Error()})
			return
		}

		// Bind the request body to the outpatient.OutpatientAdult struct
		if err := c.ShouldBindJSON(&newData); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if newData.CreatedAt == nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": "require created_at data"})
			return
		}

		drugreciperequestptr := newData.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat
		if newData.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId != nil {
			if drugreciperequestptr != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": "Cannot update with both RefID and its respective data on"})
				return
			}
		}

		labrequestptr := newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium
		if newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId != nil {
			if labrequestptr != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": "Cannot update with both RefID and its respective data on"})
				return
			}
		}

		radiologirequestptr := newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi
		if newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId != nil {
			if radiologirequestptr != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": "Cannot update with both RefID and its respective data on"})
				return
			}
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))
		newData.UpdatedAt = &now

		if drugreciperequestptr != nil {
			pharmacyrequestdata.Peresepan = *drugreciperequestptr
			newData.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObat = nil

			pharmacyJson, err := json.Marshal(pharmacyrequestdata)
			if err != nil {
				logger.LogError.Println("Error marshalling pharmacy data")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sb, err := utils.PostRequest(c, datastruct.PHARMACY, pharmacyJson)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("pharmacy: %s", err.Error())})
				return
			}
			newData.ConfidentialData.PemeriksaanSpesialistik.Terapi.ResepObatRefId = &sb
		}

		if labrequestptr != nil {
			labrequestdata = *labrequestptr
			newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Laboratorium = nil

			labJson, err := json.Marshal(labrequestdata)
			if err != nil {
				logger.LogError.Println("Error marshalling lab data")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sb, err := utils.PostRequest(c, datastruct.LABORATORY, labJson)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("lab: %s", err.Error())})
				return
			}

			newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.LabResultRefId = &sb
		}

		if radiologirequestptr != nil {
			radiologirequestdata = *radiologirequestptr
			newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.Radiologi = nil

			radiologiJson, err := json.Marshal(radiologirequestdata)
			if err != nil {
				logger.LogError.Println("Error marshalling radiology data")
				utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sb, err := utils.PostRequest(c, datastruct.RADIOLOGY, radiologiJson)
			if err != nil {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("radiology: %s", err.Error())})
				return
			}
			newData.ConfidentialData.PemeriksaanSpesialistik.PemeriksaanPenunjang.RadiologiResultRefId = &sb
		}

		// Define a filter to find the document by noRekamMedis
		filter := bson.M{
			"_id":    objID,
			"no_ihs": noIHS,
		}

		confidentialEncryptedField := utils.EncryptRandom(
			newData.ConfidentialData,
			oic.ClientEncryption,
			oic.EncryptionOpts,
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

		// Update the document in the collection
		result, err := oic.ExaminationCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			utils.JSON(c, http.StatusNotFound, gin.H{"error": "No data matched the parameter"})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d outpatient examination data updated successfully", result.ModifiedCount)})
	}
}

func (oic *OutpatientExaminationController) DeleteOutpatientExaminationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		objID, err := primitive.ObjectIDFromHex(c.Param("objID"))
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Define a filter to find the document by noPermintaan
		filter := bson.M{"_id": objID}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		// Create an update document
		update := bson.M{"$set": bson.M{
			"deleted_at": now,
		}}

		// Update the document in the collection
		_, err = oic.ExaminationCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete the document from the collection
		result, err := oic.ExaminationCollection.DeleteOne(context.Background(), filter)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success message
		utils.JSON(c, http.StatusOK, gin.H{"message": fmt.Sprintf("%d outpatient examination data deleted successfully", result.DeletedCount)})
	}
}
