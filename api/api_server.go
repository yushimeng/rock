package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	Route *gin.Engine
}

func NewApiServer() *APIServer {
	apis := &APIServer{
		Route: gin.Default(),
	}

	return apis
}

func (apis *APIServer) Serve() (err error) {
	apis.Route.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	apis.Route.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	logrus.Info("api done.....")
	return
}
