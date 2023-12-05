package router

import (
	"service-radiology/config"
	fasyankes_controllers "service-radiology/controllers"
	"service-radiology/datastruct"
	"service-radiology/db/csfle"
	"service-radiology/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouterConfig struct {
	Client              *mongo.Client
	RadiologyController *fasyankes_controllers.RadiologyController
}

func InitRouter(client *mongo.Client, csfle *csfle.CSFLE) *gin.Engine {
	routerConfig := RouterConfig{
		Client:              client,
		RadiologyController: fasyankes_controllers.InitRadiologyController(client, csfle),
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
	v1 := router.Group("/api/v1")
	v1.Use(gin.Logger(), gin.Recovery())
	v1.Use(middleware.Authentication(config.JWTPublicKey))

	resource := v1.Group("/resource")
	consentGetter := routerConfig.RadiologyController.GetPatientConsent
	resource.Use(middleware.Authorization(datastruct.RADIOLOGI))

	authUpdateConfig := map[string]string{
		"filterKey": "_id",
		"paramKey":  "Id",
	}

	ap := middleware.AcceptableParams{
		Queries: []string{},
	}

	ap2 := middleware.AcceptableParams{
		Queries: []string{"jenis_pemeriksaan", "nama_pemeriksaan"},
	}

	resource.GET("/radiology/:noIHS",
		middleware.GetConsent(consentGetter),
		middleware.Sanitize(ap2),
		routerConfig.RadiologyController.GetAllRadiologyDataHandler())

	resource.POST("/radiology",
		middleware.Sanitize(ap),
		routerConfig.RadiologyController.CreateRadiologyDataHandler())

	resource.PUT("/radiology/:noIHS/:Id",
		middleware.GetConsent(consentGetter),
		middleware.AuthorizationUpdate(authUpdateConfig, routerConfig.RadiologyController.FaskesCollection),
		middleware.Sanitize(ap),
		routerConfig.RadiologyController.UpdateRadiologyDataHandler())

	resource.DELETE("/radiology/:Id",
		middleware.AuthorizationDelete(authUpdateConfig, routerConfig.RadiologyController.FaskesCollection),
		middleware.Sanitize(ap),
		routerConfig.RadiologyController.DeleteRadiologyDataHandler())

	resource.POST("/radiology/consent",
		middleware.Sanitize(ap),
		fasyankes_controllers.ConsentHandler(routerConfig.RadiologyController.ConsentCollection))

	request := v1.Group("/request")
	request.Use(middleware.Authorization(datastruct.DOKTER))

	request.GET("/radiology/:noIHS/:Id", middleware.GetConsent(consentGetter), routerConfig.RadiologyController.GetRadiologyDataById())
	request.POST("/radiology", routerConfig.RadiologyController.CreateRadiologyRequest())

	return router
}
