package client

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func ServeHome(w http.ResponseWriter, r *http.Request) {
	logrus.Info(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "web/static/index.html")
}

// serveWs handles websocket requests from the peer
func ServeWs(chatClient *ChatClient, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("upgrader.Upgrade: ", err)
		return
	}

	chatClient.userConn = conn

	go chatClient.WriteToUser()
	go chatClient.ReadFromUser()
}
