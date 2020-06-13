package framework

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	ErrorInternalError = errors.New("Whoops! Something went wrong :(")
	ErrorInvalidJson   = errors.New("The JSON request is invalid")
)

func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) != 0 {
		for _, e := range c.Errors {
			log.Println(e.Error(), e.Type)

			switch e.Type {
			case gin.ErrorTypePublic:
				if !c.Writer.Written() {
					status := http.StatusInternalServerError
					if c.Writer.Status() != http.StatusOK {
						status = c.Writer.Status()
					}
					c.AbortWithStatusJSON(status, gin.H{"error": e.Error()})
				}

			case gin.ErrorTypeBind:
				if e.Err == e.Err.(*json.SyntaxError) {
					c.AbortWithStatusJSON(c.Writer.Status(), gin.H{"error": ErrorInvalidJson.Error()})
				} else if e.Error() == "unexpected EOF" {
					c.AbortWithStatusJSON(c.Writer.Status(), gin.H{"error": ErrorInvalidJson.Error()})
				} else {
					errs := e.Err.(validator.ValidationErrors)
					list := make(map[string]string)
					for _, err := range errs {
						list[toSnakeCase(err.Name)] = validationErrorToText(err)
					}

					// Make sure we maintain the preset response status
					status := http.StatusBadRequest
					if c.Writer.Status() != http.StatusOK {
						status = c.Writer.Status()
					}
					c.AbortWithStatusJSON(status, gin.H{"errors": list})
				}
			}
		}

		if !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ErrorInternalError.Error()})
		}
	}
}

func validationErrorToText(e *validator.FieldError) string {
	switch e.Tag {
	case "required":
		return fmt.Sprintf("%s is required", e.Field)
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", e.Field, e.Param)
	case "min":
		return fmt.Sprintf("%s must be longer than %s", e.Field, e.Param)
	case "email":
		return fmt.Sprintf("Invalid email format")
	case "len":
		return fmt.Sprintf("%s must be %s characters long", e.Name, e.Param)
	}
	return fmt.Sprintf("%s is not valid", e.Field)
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
