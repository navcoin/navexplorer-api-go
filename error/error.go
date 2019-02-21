package error

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandleError(c *gin.Context, err error, status int) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status": status,
		"message": err.Error(),
	})
}