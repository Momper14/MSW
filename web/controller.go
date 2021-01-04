package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/momper14/msw/wrapper"
	"github.com/shaj13/go-guardian/v2/auth"
	"github.com/shaj13/go-guardian/v2/auth/strategies/basic"
	"github.com/shaj13/go-guardian/v2/auth/strategies/union"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

var (
	healthy  int32
	strategy union.Union
)

// inits viper
func init() {
	viper.SetDefault("web.addr", ":8080")
	viper.SetDefault("web.prefix", "")
	viper.SetDefault("web.user", "user")
	viper.SetDefault("web.password", "password")
}

// init Go Guardian
func init() {
	strategy = union.New(basic.New(validateUser))
}

func validateUser(ctx context.Context, r *http.Request, userName, password string) (auth.Info, error) {
	if userName == viper.GetString("web.user") && password == viper.GetString("web.password") {
		return auth.NewDefaultUser(userName, "0", nil, nil), nil
	}

	return nil, fmt.Errorf("Invalid credentials")
}

func middleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	_, _, err := strategy.AuthenticateRequest(r)
	if err != nil {
		logrus.Warn(err)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
		http.Error(w, "Not Authorized", http.StatusUnauthorized)
		return
	}
	next.ServeHTTP(w, r)
}

// Controller to controll the web server
type Controller struct {
	Server *http.Server
	Hub    *Hub
}

// latestLog reads the latest Minecraft Server log
func latestLog() []string {
	file := fmt.Sprintf("%s/%s", viper.GetString("mc.workingdir"), viper.GetString("log.filename"))
	content, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return make([]string, 0)
	}
	if err != nil {
		logrus.Error(err)
		return make([]string, 0)
	}

	return strings.Split(string(content), "\n")

}

// NewController initialises a new web controller
func NewController(wrapper *wrapper.Wrapper) *Controller {
	c := Controller{}
	prefix := viper.GetString("web.prefix")
	logrus.Infof("using prefix %s", prefix)

	c.Hub = NewHub()
	c.Hub.Subscribe(wrapper)

	router := mux.NewRouter()
	router.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) { serveHome(wrapper, w, r) }).Methods("GET")
	router.PathPrefix(prefix + "/static/").Handler(http.StripPrefix(prefix+"/static/", http.FileServer(http.Dir("./static")))).Methods("GET")
	router.HandleFunc(prefix+"/ws", func(w http.ResponseWriter, r *http.Request) { ServeWs(c.Hub, wrapper, w, r) }).Methods("GET")
	router.Handle(prefix+"/healthz", healthz()).Methods("GET")

	n := negroni.Classic()
	//n.Use(auth.Basic(viper.GetString("web.user"), viper.GetString("web.password")))
	n.Use(negroni.HandlerFunc(middleware))
	n.UseHandler(router)

	c.Server = &http.Server{
		Addr:           viper.GetString("web.addr"),
		Handler:        n,
		WriteTimeout:   15 * time.Second,
		ReadTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &c
}

// Run starts the web server
func (c *Controller) Run() {
	go c.Hub.Run()
	go func() {
		if err := c.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
		}
	}()
	atomic.StoreInt32(&healthy, 1)
}

// Down stops the web server
func (c *Controller) Down(wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	atomic.StoreInt32(&healthy, 0)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.Server.SetKeepAlivesEnabled(false)
	if err := c.Server.Shutdown(ctx); err != nil {
		logrus.Errorf("Could not gracefully shutdown the server: %v\n", err)
		return
	}

	logrus.Info("Web Server Exited Properly")

}
