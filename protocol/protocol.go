package protocol

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/maansthoernvik/locksmith/log"
)

type ServerMessageType byte

const (
	Acquire ServerMessageType = 0
	Release ServerMessageType = 1
)

type ClientMessageType byte

const (
	Acquired ClientMessageType = 0
)

var ServerMessageDecodeError = errors.New("Server message decoding error")
var ClientMessageDecodeError = errors.New("Client message decoding error")
var ServerMessageTypeError = errors.New("Server message type not found")
var ClientMessageTypeError = errors.New("Client message type not found")
var LockTagSizeError = errors.New("Lock tag size does not match actual lock tag size")
var LockTagEncodingError = errors.New("Lock tag was not valid UTF8")

type ServerMessage struct {
	Type    ServerMessageType
	LockTag string
}

type ClientMessage struct {
	Type    ClientMessageType
	LockTag string
}

func DecodeServerMessage(bytes []byte) (*ServerMessage, error) {
	log.Debug("Decoding:", bytes)
	if len(bytes) < 3 || len(bytes) > 257 {
		return nil, ServerMessageDecodeError
	}
	log.Debug("Lock tag: ", bytes[2:])
	log.Debug("Supposed lock tag size: ", int(bytes[1]))
	if len(bytes[2:]) != int(bytes[1]) {
		return nil, LockTagSizeError
	}
	messageType, err := decodeServerMessageType(bytes)
	if err != nil {
		return nil, err
	}
	lockTag, err := decodeLockTag(bytes)
	if err != nil {
		return nil, err
	}

	return &ServerMessage{Type: messageType, LockTag: lockTag}, nil
}

func EncodeServerMessage(serverMessage *ServerMessage) []byte {
	bytes := make([]byte, 2+len(serverMessage.LockTag))
	bytes[0] = byte(serverMessage.Type)
	bytes[1] = byte(len(serverMessage.LockTag))
	for i := 0; i < len(serverMessage.LockTag); i++ {
		bytes[i+2] = byte(serverMessage.LockTag[i])
	}
	return bytes
}

func DecodeClientMessage(bytes []byte) (*ClientMessage, error) {
	log.Debug("Decoding: ", bytes)
	if len(bytes) < 3 || len(bytes) > 257 {
		return nil, ClientMessageDecodeError
	}
	log.Debug("Lock tag: ", bytes[2:])
	log.Debug("Supposed lock tag size: ", int(bytes[1]))
	if len(bytes[2:]) != int(bytes[1]) {
		return nil, LockTagSizeError
	}
	messageType, err := decodeClientMessageType(bytes)
	if err != nil {
		return nil, err
	}
	lockTag, err := decodeLockTag(bytes)
	if err != nil {
		return nil, err
	}

	return &ClientMessage{Type: messageType, LockTag: lockTag}, nil
}

func EncodeClientMessage(clientMessage *ClientMessage) []byte {
	bytes := make([]byte, 2+len(clientMessage.LockTag))
	log.Debug("Initialized slice with size: ", len(bytes))
	bytes[0] = byte(Acquired)
	log.Debug("Added Acquired message type: ", bytes)
	bytes[1] = byte(len(clientMessage.LockTag))
	log.Debug("Added lock tag size: ", bytes)
	log.Debug("Encoding lock tag: ", clientMessage.LockTag)
	for i := 0; i < len(clientMessage.LockTag); i++ {
		bytes[i+2] = byte(clientMessage.LockTag[i])
	}
	log.Debug("Encoded client message:", bytes)

	return bytes
}

func decodeServerMessageType(bytes []byte) (ServerMessageType, error) {
	switch ServerMessageType(bytes[0]) {
	case Acquire:
		return Acquire, nil
	case Release:
		return Release, nil
	}
	return 0, ServerMessageTypeError
}

func decodeClientMessageType(bytes []byte) (ClientMessageType, error) {
	switch ClientMessageType(bytes[0]) {
	case Acquired:
		return Acquired, nil
	}
	return 0, ClientMessageTypeError
}

func decodeLockTag(bytes []byte) (string, error) {
	lockTag := bytes[2:]
	if !utf8.Valid(lockTag) {
		return "", LockTagEncodingError
	}
	builder := strings.Builder{}
	builder.Write(lockTag)
	return builder.String(), nil
}
