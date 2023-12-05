package router

import (
	"service-auth-client/config"
	client_controllers "service-auth-client/controllers"
	"service-auth-client/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouterConfig struct {
	Client           *mongo.Client
	ClientController *client_controllers.ClientController
}

func InitRouter(client *mongo.Client) *gin.Engine {
	routerConfig := RouterConfig{
		Client:           client,
		ClientController: client_controllers.InitClientController(client),
	}

	return routerConfig.SetRouter()
}

func (routerConfig *RouterConfig) SetRouter() *gin.Engine {
	// Set up the Gin router
	router := gin.New()

	router.GET("/live", LivenessCheck())
	router.GET("/ready", ReadinessCheck(routerConfig.Client))
	router.Use(middleware.CORS(), middleware.Timekeep(time.Duration(config.TimestampSkew)*time.Millisecond))

	// Define routes
	ap := middleware.AcceptableParams{
		Queries: []string{},
	}
	v1 := router.Group("/api/v1")
	v1.Use(gin.Logger(), gin.Recovery())
	v1.Use(middleware.Sanitize(ap))

	client := v1.Group("/client")
	client.POST("/login", routerConfig.ClientController.LoginClient())
	return router
}
