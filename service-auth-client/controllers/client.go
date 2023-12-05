package client_controllers

import (
	"context"
	"errors"
	"net/http"
	"service-auth-client/config"
	"service-auth-client/datastruct"
	client_credential "service-auth-client/datastruct/client"
	"service-auth-client/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ClientController struct {
	Collection *mongo.Collection
}

func InitClientController(client *mongo.Client) *ClientController {
	return &ClientController{
		Collection: client.Database("client").Collection("credentials"),
	}
}

func (uc *ClientController) GetUserByClientID(client_id string) (*client_credential.GetClientData, error) {
	filter := bson.M{}

	filter["client_id"] = client_id

	var result client_credential.GetClientData
	err := uc.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (uc *ClientController) LoginClient() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data client_credential.Credential

		if err := c.ShouldBindJSON(&data); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userdata, err := uc.GetUserByClientID(data.ClientID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": client_credential.IncorrectCredentialError.Error()})
				return
			}

			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = userdata.CheckSecret(data.ClientSecret)
		if err != nil {
			if errors.Is(err, client_credential.IncorrectCredentialError) {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": client_credential.IncorrectCredentialError.Error()})
				return
			}
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		jwt := utils.JWTPayload{
			Issuer:   "13519220@oauth.std.stei.itb.ac.id",
			Role:     datastruct.ADMIN,
			Subject:  data.AdminName,
			Audience: []string{userdata.ClientID},
		}

		token, err := jwt.GenerateToken(config.JWTPrivateKey, time.Duration(config.JWTDuration)*time.Second)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, gin.H{"status": "success", "token": token})

	}
}
