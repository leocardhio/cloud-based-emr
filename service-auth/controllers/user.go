package user_controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"service-auth/config"
	"service-auth/datastruct/user"
	"service-auth/db/csfle"
	"service-auth/logger"
	"service-auth/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	Collection *mongo.Collection

	ClientEncryption *mongo.ClientEncryption
	EncryptionOpts   *options.EncryptOptions
}

func InitUserController(client *mongo.Client, csfle *csfle.CSFLE) *UserController {
	return &UserController{
		Collection: client.Database("user").Collection("credentials"),

		ClientEncryption: csfle.ClientEncryption,
		EncryptionOpts:   options.Encrypt().SetKeyID(*csfle.DEK),
	}
}

func (uc *UserController) GetUserByEmail(email string) (*user.CreateUserData, error) {
	filter := bson.M{}

	filter["encrypted_email"] = utils.EncryptDeterministic(
		email,
		uc.ClientEncryption,
		uc.EncryptionOpts,
	)

	var result user.CreateUserData
	err := uc.Collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (uc *UserController) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data user.CreateUserData

		if err := c.ShouldBindJSON(&data); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// check if user email already exists
		userData, err := uc.GetUserByEmail(*data.Email)
		if userData != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": user.DuplicateEmailError.Error()})
			return
		}

		err = data.HashPassword()
		if err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().Truncate(time.Duration(time.Millisecond))

		data.CreatedAt = &now
		data.UpdatedAt = &now
		data.UserCreator = c.GetString("userIdentification")
		data.CreatedBy = c.GetString("userClient")

		nameEncryptedField := utils.EncryptDeterministic(
			data.Name,
			uc.ClientEncryption,
			uc.EncryptionOpts,
		)

		passwordEncryptedField := utils.EncryptDeterministic(
			data.Password,
			uc.ClientEncryption,
			uc.EncryptionOpts,
		)

		emailEncryptedField := utils.EncryptDeterministic(
			data.Email,
			uc.ClientEncryption,
			uc.EncryptionOpts,
		)

		data.NameEncrypted = nameEncryptedField
		data.Name = nil

		data.PasswordEncrypted = passwordEncryptedField
		data.Password = nil

		data.EmailEncrypted = emailEncryptedField
		data.Email = nil

		json, err := json.Marshal(data)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		signature := utils.GenerateSignature(string(json))
		data.Signature = &signature

		_, err = uc.Collection.InsertOne(context.Background(), data)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusCreated, gin.H{"message": "User data created successfully"})
	}
}

func (uc *UserController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data user.Credential

		if err := c.ShouldBindJSON(&data); err != nil {
			utils.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userdata, err := uc.GetUserByEmail(data.Email)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": user.IncorrectCredentialError.Error()})
				return
			}

			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id := userdata.ID
		signature := userdata.Signature
		userdata.Signature = nil
		userdata.ID = primitive.NilObjectID

		dataByte, err := json.Marshal(userdata)
		if err != nil {
			logger.LogPanic.Panicf("Failed to marshall json data")
		}

		_, err = utils.VerifySignature(string(dataByte), *signature)
		if err != nil {
			logger.LogWarning.Printf("Data with ID [%s] was tampered\n", id.Hex())
			utils.JSON(c, http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		utils.Decrypt(
			userdata.PasswordEncrypted,
			uc.ClientEncryption,
		).Unmarshal(&userdata.Password)

		err = userdata.CheckPassword(data.Password)
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				utils.JSON(c, http.StatusBadRequest, gin.H{"error": user.IncorrectCredentialError.Error()})
				return
			}
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		jwt := utils.JWTPayload{
			Issuer:   "13519220@auth.std.stei.itb.ac.id",
			Role:     userdata.Role,
			Subject:  data.Email,
			Audience: []string{data.ClientID},
		}

		token, err := jwt.GenerateToken(config.JWTPrivateKey, time.Duration(config.JWTDuration)*time.Second)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.JSON(c, http.StatusOK, gin.H{"status": "success", "token": token})
	}
}
