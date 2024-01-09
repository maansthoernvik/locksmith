package server

import (
	"context"
	"testing"
)

func TestServer_Stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	locksmith := New(&LocksmithOptions{Port: 30001})

	go func() {
		cancel()
	}()

	err := locksmith.Start(ctx)
	if err != nil {
		t.Error("Error from Locksmith.Start: ", err)
	}

	t.Log("Locksmith stopped")
}
