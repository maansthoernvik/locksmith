package queuetest

import (
	"sync"
	"testing"

	"github.com/maansthoernvik/locksmith/vault/queue"
)

func Test_Multi_Enqueue(t *testing.T) {
	expectedCallCount := 100
	ts := &testSynchronized{}
	q := queue.NewMultiQueue(10, 300, ts)
	wg := sync.WaitGroup{}
	wg.Add(expectedCallCount)

	for i := 0; i < expectedCallCount; i++ {
		t.Log("enqueued", i)
		q.Enqueue("lt", func(lockTag string) {
			t.Log("callback called")
			wg.Done()
		})
	}

	wg.Wait()

	if expectedCallCount != ts.callCount {
		t.Error("Expected count", expectedCallCount, "got", ts.callCount)
	}
}

func Test_Multi_Waitlist(t *testing.T) {
	expectedCallCount := 10
	ts := &testSynchronized{}
	q := queue.NewMultiQueue(10, 300, ts)
	wg := sync.WaitGroup{}
	wg.Add(expectedCallCount)

	for i := 0; i < expectedCallCount; i++ {
		t.Log("waitlisting", i)
		q.Waitlist("lt", func(lockTag string) {
			t.Log("callback called")
			wg.Done()
		})
	}

	if 0 != ts.callCount {
		t.Fatal("Somehow the callback has been called prematurely!")
	}

	for i := 0; i < expectedCallCount; i++ {
		t.Log("popping", i)
		q.PopWaitlist("lt")
	}

	wg.Wait()

	if expectedCallCount != ts.callCount {
		t.Error("Expected count", expectedCallCount, "got", ts.callCount)
	}
}
