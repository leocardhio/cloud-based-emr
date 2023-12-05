package router

import (
	"context"
	"net/http"
	"service-auth-client/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func LivenessCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.JSON(c, http.StatusOK, gin.H{"message": "service is live"})
	}
}

func ReadinessCheck(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		if err := client.Ping(ctx, nil); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		utils.JSON(c, http.StatusOK, gin.H{"message": "service is ready"})
	}
}
