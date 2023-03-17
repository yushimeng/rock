package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yushimeng/rock/sip_server"
	"github.com/yushimeng/rock/util"
)

type APIServer struct {
	Route *gin.Engine
	sip   *sip_server.SipServer
	log   *logrus.Entry
}

func NewApiServer(ctx context.Context, sipServer *sip_server.SipServer) *APIServer {
	apis := &APIServer{
		Route: gin.Default(),
		sip:   sipServer,
		log:   ctx.Value(util.IdentifyLog).(*logrus.Logger).WithFields(logrus.Fields{string(util.IdentifyCaller): "api"}),
	}

	return apis
}

func (apis *APIServer) Serve() (err error) {
	apis.Route.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// However, this one will match /user/john/ and also /user/john/send
	// If no other routers match /user/john, it will redirect to /user/john/
	apis.Route.POST("/api/v1/invite/:device/:channel", func(c *gin.Context) {
		device := c.Param("device")
		channel := c.Param("channel")
		// check params
		if device == "" || channel == "" {
			c.String(http.StatusBadRequest, "params invalied")
			return
		}

		err := apis.sip.OnRecvHttpInvite(device, channel)
		if err == nil {
			c.String(http.StatusOK, "OK")
		} else {
			message := fmt.Sprintf("device:%s channel:%s err:%v", device, channel, err)
			apis.log.Infof("response: %s", message)
			c.String(http.StatusBadRequest, message)
		}
	})

	go apis.Route.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	return err
}
