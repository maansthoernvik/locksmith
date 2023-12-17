package server

import (
	"context"
	"testing"
	"time"
)

func TestServer_Stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	locksmith := New(&LocksmithOptions{Port: 30001})

	go func() {
		for {
			if locksmith.status == STARTED {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		cancel()
	}()

	err := locksmith.Start(ctx)
	if err != nil {
		t.Error("Error from Locksmith.Start: ", err)
	}

	t.Log("Locksmith stopped")
}
