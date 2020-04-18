package main

import (
	"flag"
	"net/http"

	client "internal.com/chat-client/internal/chat-client"

	"github.com/sirupsen/logrus"
)

var userName = flag.String("name", "bob", "user name")
var userAddr = flag.String("user-addr", ":8080", "http service address")
var serverAddr = flag.String("server-addr", ":1234", "server address")

func main() {
	var err error

	flag.Parse()

	logLevel, err := logrus.ParseLevel("debug")
	if err != nil {
		logrus.Fatal("Main: failed to parse log level")
	}
	logrus.SetLevel(logLevel)
	logrus.Info("Hello, World!")

	chatClient, err := client.ConnectToServer(*userName, *serverAddr)
	if err != nil {
		logrus.Fatal("Main: failed to connect to server")
	}

	go client.ReadFromServer(&chatClient)

	defer chatClient.Cleanup()

	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client.ServeWs(&chatClient, w, r)
	})

	err = http.ListenAndServe(*userAddr, nil)
	if err != nil {
		logrus.Fatal("http.ListenAndServer: ", err)
	}
}
