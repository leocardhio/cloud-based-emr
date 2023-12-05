package router

import (
	"service-auth/config"
	user_controllers "service-auth/controllers"
	"service-auth/datastruct"
	"service-auth/db/csfle"
	"service-auth/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouterConfig struct {
	Client         *mongo.Client
	UserController *user_controllers.UserController
}

func InitRouter(client *mongo.Client, csfle *csfle.CSFLE) *gin.Engine {
	routerConfig := RouterConfig{
		Client:         client,
		UserController: user_controllers.InitUserController(client, csfle),
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

	user := v1.Group("/users")
	user.POST("/login", routerConfig.UserController.Login())

	admin := v1.Group("/admin")
	admin.Use(middleware.Authentication(config.JWTAdminPublicKey))
	admin.POST("/registeruser", middleware.Authorization(datastruct.ADMIN), routerConfig.UserController.Register())

	return router
}
