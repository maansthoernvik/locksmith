package queuetest

import (
	"sync"
	"testing"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/vault/queue"
)

func Test_Single_Enqueue(t *testing.T) {
	log.SetLogLevel(log.WARNING)
	expectedCallCount := 100
	ts := &testSynchronized{}
	q := queue.NewSingleQueue(300, ts)
	wg := sync.WaitGroup{}
	wg.Add(expectedCallCount)

	for i := 0; i < expectedCallCount; i++ {
		q.Enqueue("lt", func(lockTag string) {
			wg.Done()
		})
	}

	wg.Wait()

	if expectedCallCount != ts.callCount {
		t.Error("Expected count", expectedCallCount, "got", ts.callCount)
	}
}

func Test_Single_Waitlist(t *testing.T) {
	log.SetLogLevel(log.WARNING)
	expectedCallCount := 10
	ts := &testSynchronized{}
	q := queue.NewSingleQueue(20, ts)
	wg := sync.WaitGroup{}
	wg.Add(expectedCallCount)

	for i := 0; i < expectedCallCount; i++ {
		q.Waitlist("lt", func(lockTag string) {
			wg.Done()
		})
	}

	if 0 != ts.callCount {
		t.Fatal("Somehow the callback has been called prematurely!")
	}

	for i := 0; i < expectedCallCount; i++ {
		q.PopWaitlist("lt")
	}

	wg.Wait()

	if expectedCallCount != ts.callCount {
		t.Error("Expected count", expectedCallCount, "got", ts.callCount)
	}
}
