package error

import "github.com/gin-gonic/gin"

func HandleError(c *gin.Context, err error, status int) {
	c.AbortWithStatusJSON(status, gin.H{
		"status": status,
		"message": err.Error(),
	})
}