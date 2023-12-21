package protocol

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type ServerMessageType byte

const (
	Acquire ServerMessageType = 0x0
	Release ServerMessageType = 0x1
)

type ClientMessage byte

const (
	Released ClientMessage = 0x0
)

var ServerMessageDecodeError = errors.New("Server message decoding error")
var ServerMessageTypeError = errors.New("Server message type not found")
var LockTagSizeError = errors.New("Lock tag size does not match actual lock tag size")
var LockTagEncodingError = errors.New("Lock tag was not valid UTF8")

type IncomingMessage struct {
	MessageType ServerMessageType
	LockTag     string
}

type OutgoingMessage struct {
	MessageType ClientMessage
	LockTag     string
}

func DecodeServerMessage(bytes []byte) (*IncomingMessage, error) {
	if len(bytes) < 3 || len(bytes) > 257 {
		return nil, ServerMessageDecodeError
	}
	if len(bytes[2:]) != int(bytes[1]) {
		return nil, LockTagSizeError
	}
	messageType, err := decodeServerMessageType(bytes)
	if err != nil {
		return nil, err
	}
	lockTag, err := decodeLockTag(bytes, bytes[1])
	if err != nil {
		return nil, err
	}

	return &IncomingMessage{MessageType: messageType, LockTag: lockTag}, nil
}

func EncodeClientMessage(clientMessage *OutgoingMessage) []byte {
	bytes := make([]byte, 2+len(clientMessage.LockTag))
	bytes[0] = byte(Released)
	bytes[1] = byte(len(clientMessage.LockTag))
	bytes = append(bytes, []byte(clientMessage.LockTag)...)

	return bytes
}

func decodeServerMessageType(bytes []byte) (ServerMessageType, error) {
	switch bytes[0] {
	case byte(Acquire):
		return Acquire, nil
	case byte(Release):
		return Release, nil
	}
	return 0, ServerMessageTypeError
}

func decodeLockTag(bytes []byte, lockTagSize byte) (string, error) {
	lockTag := bytes[2:]
	if !utf8.Valid(lockTag) {
		return "", LockTagEncodingError
	}
	builder := strings.Builder{}
	builder.Write(lockTag)
	return builder.String(), nil
}
