package router

import (
	"context"
	"net/http"
	"service-auth/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func HealthCheck(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		if err := client.Ping(ctx, nil); err != nil {
			utils.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "service is healthy"})
	}
}
