package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/yushimeng/rock/api"
	"github.com/yushimeng/rock/sip_server"
)

// Create a new instance of the logger. You can have any number of instances.
var log = logrus.New()

func init() {
	// The API for setting attributes is a little different than the package level
	// exported logger. See Godoc.
	log.Out = os.Stdout

	// Only log the warning severity or above.
	log.SetLevel(logrus.DebugLevel)

	// You could set this to any `io.Writer` such as a file
	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err == nil {
	// 	log.Out = file
	// } else {
	// 	log.Info("Failed to log to file, using default stderr")
	// }

	// log := log.WithFields(logrus.Fields{"request_id": 123, "user_ip": 122})

	// log.WithFields(logrus.Fields{
	// 	"animal": "walrus",
	// 	"size":   10,
	// }).Info("A group of walrus emerges from the ocean")

	// log.Info("info")
	// log.Infof("aaa")
	// requestLogger.Info("requestlogger")
	log.Info("Welcome to server.")
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), string("log"), log))
	defer cancel()

	api := api.NewApiServer()

	api.Serve()

	sip := sip_server.NewSipServer(ctx)
	go sip.ListenAndServe(ctx, "udp", "127.0.0.1:5060")

	select {
	case <-stop:
		log.Warn("recv stop signal, ready to exit")
	case <-ctx.Done():
		log.Warn("recv ctx Done, read to exit")
	}
	// no need to cancel

}
