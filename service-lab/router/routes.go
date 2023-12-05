package router

import (
	"service-lab/config"
	fasyankes_controllers "service-lab/controllers"
	"service-lab/datastruct"
	"service-lab/db/csfle"
	"service-lab/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouterConfig struct {
	Client        *mongo.Client
	LabController *fasyankes_controllers.LabController
}

func InitRouter(client *mongo.Client, csfle *csfle.CSFLE) *gin.Engine {
	routerConfig := RouterConfig{
		Client:        client,
		LabController: fasyankes_controllers.InitLabController(client, csfle),
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
	consentGetter := routerConfig.LabController.GetPatientConsent
	resource.Use(middleware.Authorization(datastruct.LABORATORIUM))

	authUpdateConfig := map[string]string{
		"filterKey": "_id",
		"paramKey":  "Id",
	}

	ap := middleware.AcceptableParams{
		Queries: []string{},
	}

	ap2 := middleware.AcceptableParams{
		Queries: []string{"nama_pemeriksaan", "noIHS", "no_registrasi_lab", "nik"},
	}
	resource.GET("/laboratory/:noIHS",
		middleware.GetConsent(consentGetter),
		middleware.Sanitize(ap2),
		routerConfig.LabController.GetAllLabDataHandler())

	resource.POST("/laboratory",
		middleware.Sanitize(ap),
		routerConfig.LabController.CreateLabDataHandler())

	resource.PUT("/laboratory/:noIHS/:Id",
		middleware.GetConsent(consentGetter),
		middleware.AuthorizationUpdate(authUpdateConfig, routerConfig.LabController.FaskesCollection),
		middleware.Sanitize(ap),
		routerConfig.LabController.UpdateLabDataHandler())

	resource.DELETE("/laboratory/:Id",
		middleware.AuthorizationDelete(authUpdateConfig, routerConfig.LabController.FaskesCollection),
		middleware.Sanitize(ap),
		routerConfig.LabController.DeleteLabDataHandler())

	resource.POST("/laboratory/consent",
		middleware.Sanitize(ap),
		fasyankes_controllers.ConsentHandler(routerConfig.LabController.ConsentCollection))

	request := v1.Group("/request")
	request.Use(middleware.Authorization(datastruct.DOKTER))

	request.GET("/laboratory/:noIHS/:Id",
		middleware.GetConsent(consentGetter),
		routerConfig.LabController.GetLabDataById())

	request.POST("/laboratory",
		routerConfig.LabController.CreateLabRequest())

	return router
}
