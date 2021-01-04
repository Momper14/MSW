package wrapper

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Controller to controll the minecraft server wrapper
type Controller struct {
	wrapper *Wrapper
}

// NewController creates a new MSW Controller
func NewController() *Controller {
	return &Controller{wrapper: NewWrapper()}
}

// Run runs the MSW
func (c *Controller) Run() {
	if err := c.wrapper.Run(); err != nil {
		logrus.Error(err)
	}
}

// Down stops the MSW
func (c *Controller) Down(wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	cs := c.wrapper.CurrentState()

	if cs == ServerOffline {
		logrus.Info("Minecraft Server already stopped")
		return
	}

	if cs == ServerStarting || cs == ServerOnline {
		if err := c.wrapper.Stop(); err != nil {
			logrus.Errorf("Server Shutdown Failed:%+v", err)
			return
		}
	}

	if err := c.wrapper.WaitUntilOffline(timeout); err != nil {
		logrus.Errorf("Server Shutdown Failed:%+v", err)
		return
	}

	logrus.Info("Minecraft Server Exited Properly")
}

// Wrapper returns the Wrapper
func (c *Controller) Wrapper() *Wrapper {
	return c.wrapper
}
