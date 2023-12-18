package protocol

import (
	"errors"
	"fmt"
	"testing"
)

func TestProtocol_decodeMessageType(t *testing.T) {
	messages := [][]byte{
		{0}, {1},
	}

	for _, messageType := range messages {
		messageType, err := decodeServerMessageType(messageType)
		if err != nil {
			t.Error("Failed to decode ", messageType)
		}
	}
}

func TestProtocol_decodeLockTag(t *testing.T) {
	messages := [][]byte{
		// Acquire
		{0, 9, 70, 70, 70, 70, 70, 70, 70, 70, 70},
		{0, 8, 70, 70, 70, 70, 70, 70, 70, 70},
		// Release
		{1, 9, 49, 49, 49, 49, 49, 49, 49, 49, 49},
		{1, 8, 49, 49, 49, 49, 49, 49, 49, 49},
	}
	expected := []string{
		"FFFFFFFFF",
		"FFFFFFFF",
		"111111111",
		"11111111",
	}

	for i, bytes := range messages {
		lockTag, err := decodeLockTag(bytes, messages[i][1])
		if err != nil {
			t.Error("Failed to decode ", bytes, ": ", err)
		}
		if lockTag != expected[i] {
			t.Error(lockTag, "!=", expected[i])
		}
	}
}

func TestProtocol_messedUpMessages(t *testing.T) {
	acquireMessages := [][]byte{
		// Acquire
		{0, 9, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70},
		{0, 8, 70, 70, 70, 70, 70, 70, 70},
	}

	releaseMessages := [][]byte{
		// Release
		{1, 9},
		{1},
	}

	weirdMessages := [][]byte{
		// ????
		{100, 9, 70, 70, 70, 70, 70, 70, 70, 70, 70},
		{50, 8, 70, 70, 70, 70, 70, 70, 70, 70},
	}

	for _, message := range acquireMessages {
		_, err := DecodeServerMessage(message)
		if !errors.Is(err, LockTagSizeError) {
			t.Error("Did not get error...")
		} else {
			t.Log("Got expected LockTagSizeError: ", err)
		}
	}

	for _, message := range releaseMessages {
		_, err := DecodeServerMessage(message)
		if !errors.Is(err, ServerMessageDecodeError) {
			t.Error("Did not get error...")
		} else {
			t.Log("Got expected ServerMessageDecodeError: ", err)
		}
	}

	for _, message := range weirdMessages {
		_, err := DecodeServerMessage(message)
		if !errors.Is(err, ServerMessageTypeError) {
			t.Error("Did not get error...")
		} else {
			t.Log("Got expected ServerMessageTypeError: ", err)
		}
	}
}

func TestProtocol_BadLockTagEncoding(t *testing.T) {
	messages := [][]byte{
		{0, 2, 0xc3, 0x28},
	}

	for _, message := range messages {
		_, err := DecodeServerMessage(message)
		if !errors.Is(err, LockTagEncodingError) {
			t.Error("Did not get error...")
		} else {
			t.Log("Got expected error: ", err)
		}
	}
}

// Verifies decoding of protocol messages works as intended.
//
// msg id  lock tag size  lock tag:
//
// 1 byte  1 byte         1 byte - 255 bytes
func TestProtocol_IncomingMessage(t *testing.T) {
	// One byte array = 1 message
	messages := [][]byte{
		// Acquire
		{0, 9, 49, 49, 49, 49, 49, 49, 49, 49, 49},
		{0, 8, 49, 49, 49, 49, 49, 49, 49, 49},
		{0, 1, 70},
		{0, 255, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70},
		// Release
		{1, 9, 49, 49, 49, 49, 49, 49, 49, 49, 49},
		{1, 8, 49, 49, 49, 49, 49, 49, 49, 49},
		{1, 1, 70},
		{1, 255, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
			70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70},
	}
	expected := []*IncomingMessage{
		{MessageType: Acquire, LockTag: string(messages[0][2:])},
		{MessageType: Acquire, LockTag: string(messages[1][2:])},
		{MessageType: Acquire, LockTag: string(messages[2][2:])},
		{MessageType: Acquire, LockTag: string(messages[3][2:])},
		{MessageType: Release, LockTag: string(messages[4][2:])},
		{MessageType: Release, LockTag: string(messages[5][2:])},
		{MessageType: Release, LockTag: string(messages[6][2:])},
		{MessageType: Release, LockTag: string(messages[7][2:])},
	}

	for i, bytes := range messages {
		t.Run(fmt.Sprintf("Run test iteration %d", i), func(t *testing.T) {
			im, err := DecodeServerMessage(bytes)
			if err != nil {
				t.Error(err)
			}
			if im.MessageType != expected[i].MessageType {
				t.Error("MessageType did not match for iteration ", i)
			}
			if im.LockTag != expected[i].LockTag {
				t.Error("Lock tag did not match for iteration ", i)
				t.Errorf("Got: %s Expected: %s", im.LockTag, expected[i].LockTag)
			}
		})

	}
}

func TestProtocol_EncodeClientMessage(t *testing.T) {
	res := EncodeClientMessage(&OutgoingMessage{MessageType: Released, LockTag: "abc"})

	t.Log(res)
}
