package middleware

import (
	"errors"
	"net/http"
	"service-radiology/utils"

	"github.com/gin-gonic/gin"
)

var (
	ErrQueryNotAllowed = errors.New("some queries are forbidden")
)

type AcceptableParams struct {
	Queries []string
}

func Sanitize(ap AcceptableParams) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := limitQueryTo(c.Request.URL.Query(), ap.Queries)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.Next()
	}
}

func limitQueryTo(queries map[string][]string, acceptableQ []string) error {
	counter := 0
	for i := 0; i < len(acceptableQ); i++ {
		if queries[acceptableQ[i]] != nil {
			counter++
		}
	}

	if counter != len(queries) {
		return ErrQueryNotAllowed
	}

	return nil
}
