package queuetest

import (
	"sync"
	"testing"
	"time"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/vault/queue"
)

// This test is meant to run manually to compare queue implementations.
// TODO: replace with client E2E test.
func Test_SingleQueueTimeTaken(t *testing.T) {
	t.Skip()
	start := time.Now()
	ts := &testSynchronized{}
	sq := queue.NewSingleQueue(10000, ts)
	numEnqueues := 10000
	wg := sync.WaitGroup{}
	wg.Add(numEnqueues)

	t.Log("Starting to Enqueue", numEnqueues, "items at", time.Now())
	for i := 0; i < numEnqueues; i++ {
		sq.Enqueue(randSeq(50), func(lockTag string) {
			time.Sleep(1 * time.Millisecond)
			wg.Done()
		})
	}
	t.Log("Enqueueing done at", time.Now())

	wg.Wait()

	t.Log("Wait done at", time.Now())

	t.Log("Took", time.Since(start))
}

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
