package utils

import (
	"github.com/gin-gonic/gin"
)

func JSON(c *gin.Context, code int, obj any) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.JSON(code, obj)
}

func AbortWithStatusJSON(c *gin.Context, code int, obj gin.H) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.AbortWithStatusJSON(code, obj)
}
