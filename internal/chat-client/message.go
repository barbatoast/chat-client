package client

import (
	"bytes"
	"fmt"
	"net"
	"encoding/binary"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	msgTypeNewUser = 1
	msgTypeRegisteredUser = 2
	msgTypeBroadcastMsg = 3
	fieldTypeString = 1
	fieldTypeUint32 = 2
	fieldIdTimestamp = 1
	fieldIdName = 2
	fieldIdMsg = 3
)

type msgHdr struct {
	Type uint16
	Len uint16
	Flags uint8
	Ret uint8
}

type fieldHdr struct {
	Type uint16
	Id uint16
	Len uint16
}

func createMsgNewUser(name string) ([]byte, error) {
	var err error
	var hdr msgHdr
	var field fieldHdr

	field.Type = fieldTypeString
	field.Id = fieldIdName
	field.Len = uint16(len(name) + 1)
	hdr.Type = msgTypeNewUser
	hdr.Len = uint16(binary.Size(field)) + field.Len

	buf := bytes.NewBuffer([]byte{})

	err = binary.Write(buf, binary.BigEndian, &hdr)
	if err != nil {
		logrus.Error("binary.Write: ", err)
		return []byte{}, err
	}

	err = binary.Write(buf, binary.BigEndian, &field)
	if err != nil {
		logrus.Error("binary.Write: ", err)
		return []byte{}, err
	}

	nameBuf := []byte(name)
	nameBuf = append(nameBuf, 0x00)
	err = binary.Write(buf, binary.BigEndian, &nameBuf)
	if err != nil {
		logrus.Error("binary.Write: ", err)
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func CreateMsgRegisteredUser(msg []byte) ([]byte, error) {
	var err error
	var hdr msgHdr
	var field fieldHdr

	field.Type = fieldTypeString
	field.Id = fieldIdMsg
	field.Len = uint16(len(msg) + 1)
	hdr.Type = msgTypeRegisteredUser
	hdr.Len = uint16(binary.Size(field)) + field.Len

	buf := bytes.NewBuffer([]byte{})

	err = binary.Write(buf, binary.BigEndian, &hdr)
	if err != nil {
		logrus.Error("binary.Write: ", err)
		return []byte{}, err
	}

	err = binary.Write(buf, binary.BigEndian, &field)
	if err != nil {
		logrus.Error("binary.Write: ", err)
		return []byte{}, err
	}

	msg = append(msg, 0x00)
	logrus.Debug("sending message to server ", string(msg))
	err = binary.Write(buf, binary.BigEndian, &msg)
	if err != nil {
		logrus.Error("binary.Write: ", err)
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func readServerMsg(serverConn *net.Conn) (string, error) {
	var err error
	var i int
	var timestamp uint32
	var name string
	var msg string
	var buf []byte
	var timeString string
	var hdr msgHdr
	var field fieldHdr

	err = binary.Read(*serverConn, binary.BigEndian, &hdr)
	if err != nil {
		logrus.Error("binary.Read: ", err)
		return "", err
	}

	if hdr.Type != msgTypeBroadcastMsg {
		logrus.Warn(
			"wrong message type, expected %d, got %d",
			msgTypeBroadcastMsg,
			hdr.Type,
		)
	}

	i = 0
	for i < int(hdr.Len) {
		err = binary.Read(*serverConn, binary.BigEndian, &field)
		if err != nil {
			logrus.Error("binary.Read: ", err)
			return "", err
		}

		switch field.Id {
		case fieldIdTimestamp:
			err = binary.Read(*serverConn, binary.BigEndian, &timestamp)
			if err != nil {
				logrus.Error("binary.Read: ", err)
				return "", err
			}
			tm := time.Unix(int64(timestamp), 0)
			timeString = fmt.Sprintf(
				"%d-%02d-%02d %02d:%02d:%02d",
				tm.Year(), tm.Month(), tm.Day(),
				tm.Hour(), tm.Minute(), tm.Second(),
			)
		case fieldIdName:
			buf = make([]byte, field.Len)
			err = binary.Read(*serverConn, binary.BigEndian, &buf)
			if err != nil {
				logrus.Error("binary.Read: ", err)
				return "", err
			}
			name = string(buf)
		case fieldIdMsg:
			buf = make([]byte, field.Len)
			err = binary.Read(*serverConn, binary.BigEndian, &buf)
			if err != nil {
				logrus.Error("binary.Read: ", err)
				return "", err
			}
			msg = string(buf)
		default:
			logrus.Warn("unknown field Id found %d", field.Id)
		}

		i += binary.Size(field) + int(field.Len)
	}

	return timeString + ": " + name + ": " + msg, nil
}
