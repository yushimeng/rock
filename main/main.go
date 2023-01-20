package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yushimeng/rock/api"
)

func main() {
	api := api.NewApiServer()

	sip := sip.NewSipServer()
	api.Serve()
}
