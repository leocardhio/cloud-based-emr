package router

import (
	"service-pharmacy/config"
	fasyankes_controllers "service-pharmacy/controllers"
	"service-pharmacy/datastruct"
	"service-pharmacy/db/csfle"
	"service-pharmacy/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouterConfig struct {
	Client             *mongo.Client
	PharmacyController *fasyankes_controllers.PharmacyController
}

func InitRouter(client *mongo.Client, csfle *csfle.CSFLE) *gin.Engine {
	routerConfig := RouterConfig{
		Client:             client,
		PharmacyController: fasyankes_controllers.InitPharmacyController(client, csfle),
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
	consentGetter := routerConfig.PharmacyController.GetPatientConsent
	resource.Use(middleware.Authorization(datastruct.APOTEK))

	authUpdateConfig := map[string]string{
		"filterKey": "_id",
		"paramKey":  "Id",
	}

	ap := middleware.AcceptableParams{
		Queries: []string{},
	}

	ap2 := middleware.AcceptableParams{
		Queries: []string{"no_rekam_medis", "id_pelanggan", "id_obat", "nik"},
	}

	resource.GET("/pharmacy/:noIHS",
		middleware.GetConsent(consentGetter),
		middleware.Sanitize(ap2),
		routerConfig.PharmacyController.GetAllPharmacyHandler())

	resource.POST("/pharmacy",
		middleware.Sanitize(ap),
		routerConfig.PharmacyController.CreatePharmacyHandler())

	resource.PUT("/pharmacy/:noIHS/:Id",
		middleware.GetConsent(consentGetter),
		middleware.AuthorizationUpdate(authUpdateConfig, routerConfig.PharmacyController.FaskesCollection),
		middleware.Sanitize(ap),
		routerConfig.PharmacyController.UpdatePharmacyHandler())

	resource.DELETE("/pharmacy/:Id",
		middleware.AuthorizationDelete(authUpdateConfig, routerConfig.PharmacyController.FaskesCollection),
		middleware.Sanitize(ap),
		routerConfig.PharmacyController.DeletePharmacyHandler())

	resource.POST("/pharmacy/consent",
		middleware.Sanitize(ap),
		fasyankes_controllers.ConsentHandler(routerConfig.PharmacyController.ConsentCollection))

	request := v1.Group("/request")
	request.Use(middleware.Authorization(datastruct.DOKTER))

	request.GET("/pharmacy/:noIHS/:Id", middleware.GetConsent(consentGetter), routerConfig.PharmacyController.GetPharmacyDataById())
	request.POST("/pharmacy", routerConfig.PharmacyController.CreatePharmacyRequest())

	return router
}
