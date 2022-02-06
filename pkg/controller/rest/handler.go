package rest

import (
	"dns-proxy/pkg/domain/denylist"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handler(denySvc denylist.Service) *gin.Engine {
	router := gin.New()
	router.GET("ping", ping)
	router.PUT("/deny/:domain", addDeniedDomain(denySvc))
	return router
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func addDeniedDomain(svc denylist.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Param("domain")
		if domain == "" {
			c.JSON(http.StatusBadRequest, newJSONError(errors.New("missing domain")))
			return
		}
		err := svc.AddDeniedDomain(domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, newJSONError(err))
			return
		}
		c.JSON(http.StatusOK, newJSONMessage(fmt.Sprintf("%s added to denylist successfully", domain)))
		return
	}
}

func newJSONError(err error) map[string]string {
	jsonError := make(map[string]string)
	jsonError["error"] = fmt.Sprintf("%v", err)
	return jsonError
}

func newJSONMessage(message string) map[string]string {
	jsonMessage := make(map[string]string)
	jsonMessage["message"] = message
	return jsonMessage
}
