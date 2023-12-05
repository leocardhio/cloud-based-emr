package middleware

import (
	"errors"
	"net/http"
	"service-auth/datastruct"
	user "service-auth/datastruct/user"
	"service-auth/logger"
	"service-auth/utils"

	"github.com/gin-gonic/gin"
)

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
