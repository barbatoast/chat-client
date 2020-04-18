package client

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// send pings to peer with this period, must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// maximum message size allowed from peer
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space = []byte{' '}
)

type ChatClient struct {
	serverConn *net.Conn
	userConn *websocket.Conn
	broadcastChan chan string
}

func (chatClient *ChatClient) Cleanup() {
	(*chatClient.serverConn).Close()
}

func (chatClient *ChatClient) GetServerMsg() (string, error) {
	var err error
	var msg string

	msg, err = readServerMsg(chatClient.serverConn)
	if err != nil {
		logrus.Error("ChatClient: failed to read message from server")
		return "", err
	}

	return msg, nil
}

func (chatClient *ChatClient) ReadFromUser() {
	defer func() {
		chatClient.userConn.Close()
	}()
	chatClient.userConn.SetReadLimit(maxMessageSize)
	chatClient.userConn.SetReadDeadline(time.Now().Add(pongWait))
	chatClient.userConn.SetPongHandler(func(string) error {
		chatClient.userConn.SetReadDeadline(time.Now().Add(pongWait));
		return nil;
	})
	for {
		_, message, err := chatClient.userConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
												websocket.CloseGoingAway,
												websocket.CloseAbnormalClosure) {
				logrus.Error("websocket.IsUnexpectedError: ", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		message, err = CreateMsgRegisteredUser(message)
		if err != nil {
			logrus.Error("ChatClient: failed to create registered user message")
			break;
		}
		err = chatClient.SendServerMsg(message)
		if err != nil {
			logrus.Error("failed to send message to server")
			break;
		}
	}
}

func (chatClient *ChatClient) SendServerMsg(msg []byte) (error) {
	var err error

	err = binary.Write(*chatClient.serverConn, binary.BigEndian, msg)
	if err != nil {
		logrus.Error("ChatClient: failed to send message to server")
		return err
	}

	return nil
}

func (chatClient *ChatClient) WriteToUser() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		chatClient.userConn.Close()
	}()
	for {
		select {
		case message, ok := <- chatClient.broadcastChan:
			chatClient.userConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the server has closed the channel
				chatClient.userConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := chatClient.userConn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write([]byte(message))
			w.Write(newline)	// ?
			logrus.Info(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			chatClient.userConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := chatClient.userConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
