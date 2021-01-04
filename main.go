package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/momper14/msw/init"
	"github.com/momper14/msw/web"
	"github.com/momper14/msw/wrapper"
	"github.com/sirupsen/logrus"
)

const (
	timeout = 30 * time.Second
)

func main() {

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, os.Interrupt)

	mcController := wrapper.NewController()
	webController := web.NewController(mcController.Wrapper())

	go webController.Run()
	go mcController.Run()

	<-quit

	logrus.Info("Shutting down server...")
	var wg sync.WaitGroup
	wg.Add(2)

	go webController.Down(&wg, timeout)

	go mcController.Down(&wg, timeout)

	wg.Wait()

}
