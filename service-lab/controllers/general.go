package fasyankes_controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"service-lab/datastruct"
	"service-lab/datastruct/user"
	"service-lab/logger"
	"service-lab/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UniqueFilter struct {
	Collection *mongo.Collection
	Key        string
	Value      string
}

type DocumentClientID struct {
	ClientID string `json:"-" bson:"client_id"`
}

func IsHaveConsent(consentGetter user.ConsentGetter, noIHS, cid string) *bool {
	patientConsent, err := consentGetter(noIHS)
	isConsentFound := false
	if err == nil {
		for i := 0; i < len(patientConsent.ConsentTo); i++ {
			if cid == patientConsent.ConsentTo[i].ClientID {
				isConsentFound = true
				break
			}
		}
	}

	return &isConsentFound
}

func HaveUpdatePermission(uf UniqueFilter, cid string) (*bool, error) {
	filter := bson.M{}
	filter[uf.Key] = uf.Value

	if uf.Key == "_id" {
		objId, err := primitive.ObjectIDFromHex(uf.Value)
		if err != nil {
			return nil, err
		}

		filter[uf.Key] = objId
	}

	var existingDoc DocumentClientID
	err := uf.Collection.FindOne(context.Background(), filter).Decode(&existingDoc)
	if err != nil {
		return nil, err
	}

	result := false
	if existingDoc.ClientID == "" || existingDoc.ClientID == cid {
		result = true
	}

	return &result, nil
}

func HaveDeletePermission(uf UniqueFilter, cid string) (*bool, error) {
	filter := bson.M{}
	filter[uf.Key] = uf.Value

	if uf.Key == "_id" {
		objId, err := primitive.ObjectIDFromHex(uf.Value)
		if err != nil {
			return nil, err
		}

		filter[uf.Key] = objId
	}

	var existingDoc DocumentClientID
	err := uf.Collection.FindOne(context.Background(), filter).Decode(&existingDoc)
	if err != nil {
		return nil, err
	}

	result := false
	if existingDoc.ClientID == cid {
		result = true
	}

	return &result, nil
}

func ConsentHandler(collection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		var consentBody user.ConsentBody
		if err := c.ShouldBindJSON(&consentBody); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		noihs := consentBody.NoIHS

		var patientConsent user.PatientConsent
		now := time.Now().Truncate(time.Duration(time.Millisecond))
		filter := bson.M{"no_ihs": noihs}

		err := collection.FindOne(context.Background(), filter).Decode(&patientConsent)
		if errors.Is(err, mongo.ErrNoDocuments) {
			patientConsent.NoIHS = noihs
			patientConsent.CreatedAt = &now
		} else if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if consentBody.ConsentType == datastruct.OPTIN {
			consentData := user.ConsentData{
				ClientID:     c.GetString("userClient"),
				ConsentGiver: consentBody.ConsentGiver,
			}
			patientConsent.ConsentTo = append(patientConsent.ConsentTo, consentData)
		} else {
			for i := 0; i < len(patientConsent.ConsentTo); i++ {
				if patientConsent.ConsentTo[i].ClientID == c.GetString("userClient") {
					patientConsent.ConsentTo = append(patientConsent.ConsentTo[:i], patientConsent.ConsentTo[i+1:]...)
					break
				}
			}
		}

		patientConsent.UpdatedAt = &now
		json, err := json.Marshal(patientConsent)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newSignature := utils.GenerateSignature(string(json))
		patientConsent.Signature = &newSignature

		opts := options.Update().SetUpsert(true)
		update := bson.M{"$set": patientConsent}
		res, err := collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		message := fmt.Sprintf("%d consent modified", res.ModifiedCount)
		if res.ModifiedCount == 0 {
			message = fmt.Sprintf("%d consent upserted", res.UpsertedCount)
		}

		utils.JSON(c, http.StatusAccepted, gin.H{"message": message})
	}
}

func GetPatientConsent(noihs string, consentCollection *mongo.Collection) (*user.PatientConsent, error) {
	filter := bson.M{}

	filter["no_ihs"] = noihs

	var result user.PatientConsent
	err := consentCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	signature := result.Signature
	result.Signature = nil

	dataByte, err := json.Marshal(result)
	if err != nil {
		logger.LogPanic.Panicf("Failed to marshal json data")
	}

	_, err = utils.VerifySignature(string(dataByte), *signature)
	if err != nil {
		logger.LogWarning.Printf("Consent Data with NoIHS [%s] was tampered\n", result.NoIHS)
	}

	return &result, nil
}
