package middleware

import (
	"fmt"
	"net/http"
	"service-auth/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Timekeep(skew time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		sentTs, err := strconv.ParseInt(c.GetHeader("X-Timestamp"), 10, 64)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to parse timestamp from header: %s", err.Error())})
			return
		}

		receiveTs := time.Now().UnixMilli()
		fmt.Println(receiveTs, sentTs, receiveTs-sentTs, skew.Milliseconds())

		timeDiff := receiveTs - sentTs
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		if timeDiff > skew.Milliseconds() {
			utils.AbortWithStatusJSON(c, http.StatusUnauthorized, gin.H{"error": "timestamp exceed tolerance"})
			return
		}
		c.Next()
	}
}
