package client

import (
	"net"

	"github.com/sirupsen/logrus"
)

func ConnectToServer(userName string, serverAddr string) (ChatClient, error) {
	var err error

	chatClient := ChatClient{}

	// TODO(author): use user-defined address
	serverConn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		logrus.Error("ChatClient: failed to connect to server")
		return ChatClient{}, err
	}

	chatClient.serverConn = &serverConn

	// TOOD(author): use user-defined name
	newUserMsg, err := createMsgNewUser(userName)
	if err != nil {
		logrus.Error("ChatClient: failed to create new user message")
		serverConn.Close()
		return ChatClient{}, err
	}

	err = chatClient.SendServerMsg(newUserMsg)
	if err != nil {
		logrus.Error("ChatClient: failed to send new user message to server")
		serverConn.Close()
		return ChatClient{}, err
	}

	broadcastChan := make(chan string, 1)
	chatClient.broadcastChan = broadcastChan

	return chatClient, nil
}

func ReadFromServer(chatClient *ChatClient) {
	var err error
	var msg string

	for {
		msg, err = chatClient.GetServerMsg()
		if err != nil {
			logrus.Error("Failed to monitor server, closing broadcast channel")
			close(chatClient.broadcastChan)
			return
		}

		chatClient.broadcastChan <- msg
	}
}
