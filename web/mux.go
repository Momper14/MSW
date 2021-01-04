package web

import (
	"html/template"
	"net/http"
	"sync/atomic"

	"github.com/momper14/msw/wrapper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, wrapper *wrapper.Wrapper, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
}

func serveHome(wr *wrapper.Wrapper, w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("template/home.html")
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	data := IndexTemplate{
		State:  wr.CurrentState().String(),
		Prefix: viper.GetString("web.prefix"),
	}

	switch wrapper.ServerStateFor(data.State) {
	case wrapper.ServerStarting:
		data.Starting = true
	case wrapper.ServerOnline:
		data.Online = true
	case wrapper.ServerStopping, wrapper.ServerOffline:
		data.Offline = true
	}

	data.Log = latestLog()

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
}
