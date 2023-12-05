package middleware

import (
	"errors"
	"net/http"
	emr_controllers "service-outpatient/controllers"
	"service-outpatient/datastruct"
	"service-outpatient/datastruct/user"
	"service-outpatient/logger"
	"service-outpatient/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConsentGetter func(noIHS string) (*user.PatientConsent, error)

func Authentication(jwtPublicKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sentToken, err := utils.ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		claim, err := utils.VerifyToken(sentToken, jwtPublicKey)
		if errors.Is(err, user.UnauthorizedIssuerError) {
			logger.LogWarning.Printf("Subject: %s | ClientID: %s | Issuer: %s | Trying to access system using unverified token",
				claim.Subject,
				claim.Audience[0],
				claim.Issuer,
			)
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		} else if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		logger.LogInfo.Printf("Subject: %s | ClientID: %s | Issuer: %s | Accessing System",
			claim.Subject,
			claim.Audience[0],
			claim.Issuer,
		)

		// note: set context key to global constant
		c.Set("userIdentification", claim.Subject) // user's name
		c.Set("userRole", string(claim.Role))      // user's role
		c.Set("userClient", claim.Audience[0])     // which client do the user from

		c.Next()
	}
}

func Authorization(authorizedRoles ...datastruct.RoleType) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimedRole := c.GetString("userRole")

		isRoleFound := false

		for i := 0; i < len(authorizedRoles); i++ {
			if claimedRole == string(authorizedRoles[i]) {
				isRoleFound = true
				break
			}
		}

		if !isRoleFound {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"error": user.NotAuthorizedError.Error()})
			return
		}

		c.Next()
	}
}

func GetConsent(consentGetFunc ConsentGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		noIHS := c.Param("noIHS")
		clientId := c.GetString("userClient")

		patientConsent, err := consentGetFunc(noIHS)
		isConsentFound := bool(datastruct.OPTOUT)
		if err == nil {
			for i := 0; i < len(patientConsent.ConsentTo); i++ {
				if clientId == patientConsent.ConsentTo[i].ClientID {
					isConsentFound = bool(datastruct.OPTIN)
					break
				}
			}
		}

		c.Set("patientConsent", isConsentFound)

		c.Next()
	}
}

func AuthorizationUpdate(authUpdateConfig map[string]string, filteredCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		pathParamValue := c.Param(authUpdateConfig["paramKey"])

		uf := emr_controllers.UniqueFilter{
			Collection: filteredCollection,
			Key:        authUpdateConfig["filterKey"],
			Value:      pathParamValue,
		}

		haveUpdatePermission, err := emr_controllers.HaveUpdatePermission(uf, c.GetString("userClient"))
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !*haveUpdatePermission {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"forbidden": user.NotAuthorizedError.Error()})
			return
		}

		c.Next()
	}
}

func AuthorizationDelete(authUpdateConfig map[string]string, filteredCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		pathParamValue := c.Param(authUpdateConfig["paramKey"])

		uf := emr_controllers.UniqueFilter{
			Collection: filteredCollection,
			Key:        authUpdateConfig["filterKey"],
			Value:      pathParamValue,
		}

		haveDeletePermission, err := emr_controllers.HaveDeletePermission(uf, c.GetString("userClient"))
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if !*haveDeletePermission {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"forbidden": user.NotAuthorizedError.Error()})
			return
		}

		c.Next()
	}
}
