package protocol

import (
	"errors"
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

type IncomingMessage struct {
	MessageType ServerMessageType
	LockTag     string
}

func DecodeServerMessage(bytes []byte) (*IncomingMessage, error) {
	if len(bytes) < 3 || len(bytes) > 257 {
		return nil, ServerMessageDecodeError
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

func decodeServerMessageType(bytes []byte) (ServerMessageType, error) {
	switch bytes[0] {
	case byte(Acquire):
		return Acquire, nil
	case byte(Release):
		return Release, nil
	}
	return 0, ServerMessageTypeError
}

func decodeLockTag(bytes []byte, lock_tag_size byte) (string, error) {
	if len(bytes) != int(lock_tag_size)+2 {
		return "", LockTagSizeError
	}
	return string(bytes[2:]), nil
}
