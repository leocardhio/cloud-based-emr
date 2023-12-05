package router

import (
	"service-outpatient/config"
	emr_controllers "service-outpatient/controllers"
	"service-outpatient/datastruct"
	"service-outpatient/db/csfle"
	"service-outpatient/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type RouterConfig struct {
	Client *mongo.Client

	UserIdentityController *emr_controllers.UserIdentityController
	OutpatientExamination  *emr_controllers.OutpatientExaminationController
}

func InitRouter(client *mongo.Client, csfle *csfle.CSFLE) *gin.Engine {
	routerConfig := RouterConfig{
		Client:                 client,
		UserIdentityController: emr_controllers.InitUserIdentityController(client, csfle),
		OutpatientExamination:  emr_controllers.InitOutpatientExaminationController(client, csfle),
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
	consentGetter := routerConfig.OutpatientExamination.GetPatientConsent
	resource.Use(middleware.Authorization(datastruct.DOKTER))

	authUpdateConfig := map[string]string{
		"filterKey": "_id",
		"paramKey":  "objID",
	}

	// resource.GET("/identity", middleware.Authorization(datastruct.DOKTER), routerConfig.UserIdentityController.GetAllUserIdentityHandler())
	// resource.POST("/identity", middleware.Authorization(datastruct.DOKTER), routerConfig.UserIdentityController.CreateUserIdentityHandler())
	// resource.PUT("/identity/:noIHS", middleware.Authorization(datastruct.DOKTER), routerConfig.UserIdentityController.UpdateUserIdentityHandler())
	// resource.DELETE("/identity/:noIHS", middleware.Authorization(datastruct.DOKTER), routerConfig.UserIdentityController.DeleteUserIdentityHandler())

	ap := middleware.AcceptableParams{
		Queries: []string{},
	}

	resource.GET("/outpatient/patient/:noIHS",
		middleware.GetConsent(consentGetter),
		middleware.Sanitize(ap),
		routerConfig.OutpatientExamination.GetAllOutpatientExaminationHandler())

	resource.GET("/outpatient/:noIHS/:objID",
		middleware.GetConsent(consentGetter),
		middleware.Sanitize(ap),
		routerConfig.OutpatientExamination.GetOutpatientExaminationHandler())

	resource.POST("/outpatient",
		middleware.Sanitize(ap),
		routerConfig.OutpatientExamination.CreateOutpatientExaminationHandler())

	resource.PUT("/outpatient/:noIHS/:objID",
		middleware.GetConsent(consentGetter),
		middleware.AuthorizationUpdate(authUpdateConfig, routerConfig.OutpatientExamination.ExaminationCollection),
		middleware.Sanitize(ap),
		routerConfig.OutpatientExamination.UpdateOutpatientExaminationHandler())

	resource.DELETE("/outpatient/:objID",
		middleware.AuthorizationDelete(authUpdateConfig, routerConfig.OutpatientExamination.ExaminationCollection),
		middleware.Sanitize(ap),
		routerConfig.OutpatientExamination.DeleteOutpatientExaminationHandler())

	resource.POST("/outpatient/consent",
		middleware.Sanitize(ap),
		emr_controllers.ConsentHandler(routerConfig.OutpatientExamination.ConsentCollection))

	return router
}
