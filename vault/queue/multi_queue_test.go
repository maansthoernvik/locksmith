package queue

import (
	"sync"
	"testing"

	"github.com/maansthoernvik/locksmith/log"
)

type testSynchronized struct{}

func (ts *testSynchronized) Synchronized(lockTag string, callback SynchronizedAction) {
	callback(lockTag)
}

func Test_queueIndexDistribution(t *testing.T) {
	log.SetLogLevel(log.WARNING)
	numQueues := 1000
	mq := &multiQueue{
		queues:   make([]chan *queueItem, numQueues),
		hashFunc: fnv1aHash,
	}

	numberOfTags := 1000000
	result := map[uint16]int{}

	for i := 0; i < numberOfTags; i++ {
		tag := randSeq(50)
		hash := mq.hashFunc(tag)
		index := mq.queueIndexFromHash(hash)

		if index >= uint16(numQueues) {
			t.Error("Somehow hash", hash, "gave an index outside the range...")
		}

		v, ok := result[index]
		if !ok {
			result[index] = 1
		} else {
			result[index] = v + 1
		}
	}

	if len(result) != numQueues {
		t.Fatal("Length of result not equal to the number of queues, something is seriously wrong:", len(result))
	}

	previous := 0
	for _, v := range result {
		if previous == 0 {
			previous = v
			continue
		}

		// Tolerance is +- previous/10
		if v > (previous+(previous/5)) || v < (previous-(previous/5)) {
			t.Fatal("Distribution is outside tolerances, "+
				"something is wrong with the index calculator:\nprevious=", previous, "\nnew=", v)
		}
	}
}

func Test_queueIndexFromHash(t *testing.T) {
	log.SetLogLevel(log.WARNING)

	mq := &multiQueue{queues: make([]chan *queueItem, 10)}

	qi := mq.queueIndexFromHash(65535)
	if qi >= uint16(len(mq.queues)) {
		t.Fatal("Gotten queue index out of range for hash 65535:", qi)
	}

	qi = mq.queueIndexFromHash(0)
	if qi >= uint16(len(mq.queues)) {
		t.Fatal("Gotten queue index out of range for hash 0:", qi)
	}
}

func Test_WaitList(t *testing.T) {
	log.SetLogLevel(log.DEBUG)
	mq := NewMultiQueue(5, 10, &testSynchronized{}).(*multiQueue)

	firstCalled := false
	mq.Waitlist("lt", func(lockTag string) { firstCalled = true })
	if len(mq.waitlist) != 1 {
		t.Fatal("Expected wait list size to be 1, was:", len(mq.waitlist))
	}

	secondCalled := false
	mq.Waitlist("lt", func(lockTag string) { secondCalled = true })
	if len(mq.waitlist) != 1 {
		t.Fatal("Expected wait list size to be 1, was:", len(mq.waitlist))
	}

	if len(mq.waitlist["lt"]) != 2 {
		t.Fatal("Expected wait list for 'lt' to be 2, was:", len(mq.waitlist))
	}
	if firstCalled || secondCalled {
		t.Fatal("For some reason first or second waitlisted callback has been called...")
	}

	mq.PopWaitlist("lt")
	if len(mq.waitlist) != 1 {
		t.Fatal("Expected wait list size to be 1, was:", len(mq.waitlist))
	}
	if !firstCalled {
		t.Fatal("First waitlisted callback should have been called")
	}
	if secondCalled {
		t.Fatal("Second callback should NOT have been called")
	}

	if len(mq.waitlist["lt"]) != 1 {
		t.Fatal("Expected wait list for 'lt' to be 1, was:", len(mq.waitlist))
	}

	mq.PopWaitlist("lt")
	if !secondCalled {
		t.Fatal("Second callback should have been called")
	}
	if len(mq.waitlist) != 0 {
		t.Fatal("Expected wait list size to be 0, was:", len(mq.waitlist))
	}
}

func Test_Enqueue(t *testing.T) {
	mq := NewMultiQueue(5, 10, &testSynchronized{}).(*multiQueue)
	calls := 1000

	wg := sync.WaitGroup{}
	wg.Add(calls)
	for i := 0; i < calls; i++ {
		mq.Enqueue(randSeq(20), func(lockTag string) {
			wg.Done()
		})
	}

	wg.Wait()
}

const BENCHMARKING_SEQUENCE_SIZE = 100

func Benchmark_queueIndex(b *testing.B) {
	mq := &multiQueue{queues: make([]chan *queueItem, 10)}

	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mq.queueIndexFromHash(fnv1aHash(randSeq(BENCHMARKING_SEQUENCE_SIZE)))
		}
	})
}
