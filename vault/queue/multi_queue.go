package queue

import (
	"math"

	"github.com/maansthoernvik/locksmith/log"
)

// An implementation of the QueueLayer interface utilizaing multiple channels
// to provide a greater number of synchronization threads to a using vault.
// Enqueues lead to one of the channels picking up a message and dispatching
// to the vault's Synchronized implementation. A channel is picked based on
// a hashing mechanism, each lock tag is hashed into a number between 0 and 65535,
// and then an index is derived from that hash, based on a simple calculation:
//
//	queueIndex = hash / (MAX_HASH / len(multiQueue.queues))
//
// A message is then dispatched to the picked channel by:
//
//	multiQueue.queues[queueIndex] <- &queueItem{lockTag: lockTag, callback: callback}
//
// This method ensures the same index channel always handles the same hash(es),
// and eliminates race conditions between clients of Locksmith.
type multiQueue struct {
	queues       []chan *queueItem
	hashFunc     func(string) uint16
	waitlist     map[string][]*queueItem
	synchronized Synchronized
}

// Creates a QueueLayer with multiple underlying go routines for quicker
// dispatch of lock acquisitions and releases. To dispatch, each lock tag
// is hashed into a number, each queue handles a range.
func NewMultiQueue(
	concurrency int,
	capacity int,
	synchronized Synchronized,
) QueueLayer {
	ql := &multiQueue{
		queues:       make([]chan *queueItem, concurrency),
		hashFunc:     fnv1aHash,
		waitlist:     make(map[string][]*queueItem),
		synchronized: synchronized,
	}

	// Initialize queues, queue[0] is responsible for the range 0 -> 65535 / numQueues and so on
	for i := 0; i < concurrency; i++ {
		ql.queues[i] = make(chan *queueItem, capacity)

		go func(i int, queue chan *queueItem) {
			log.Info("Starting multi queue #", i)
			for {
				qi := <-queue
				ql.synchronized.Synchronized(qi.lockTag, qi.callback)
			}
		}(i, ql.queues[i])
	}

	return ql
}

// Enqueue a lock tag, expect a call to the Synchronized implementor once the queue layer
// has gotten a hold of a synchronization Go-routine specific to the resulting hash of
// the lock tag.
func (multiQueue *multiQueue) Enqueue(lockTag string, callback func(string)) {
	log.Debug("Queueing up lock tag: ", lockTag)
	hash := multiQueue.hashFunc(lockTag)
	queueIndex := multiQueue.queueIndexFromHash(hash)
	log.Debug("Got hash ", hash, " enqueueing with queue #", queueIndex)
	multiQueue.queues[queueIndex] <- &queueItem{lockTag: lockTag, callback: callback}
}

// IMPORTANT: only call from synchronization Go-routines.
// Waitlist the input action, related to the given lock tag. Appends the action
// to the back of the waitlist of the lock tag.
func (multiQueue *multiQueue) Waitlist(lockTag string, callback func(string)) {
	log.Debug("Waitlisting client for lock tag: ", lockTag)
	_, ok := multiQueue.waitlist[lockTag]
	if !ok {
		multiQueue.waitlist[lockTag] = []*queueItem{{lockTag, callback}}
	} else {
		multiQueue.waitlist[lockTag] = append(multiQueue.waitlist[lockTag], &queueItem{lockTag, callback})
	}
	log.Debug("Resulting waitlist state:\n", multiQueue.waitlist)
}

// IMPORTANT: only call from synchronization Go-routines.
// Pop from the waitlist belonging to the input lock tag, results in a waitlisted
// action being called directly as this function assumes we're already in the scope
// of a synchronization thread (thus skipping a second call to Synchronized(...)).
func (multiQueue *multiQueue) PopWaitlist(lockTag string) {
	log.Debug("Popping from waitlist: ", lockTag)
	if wl, ok := multiQueue.waitlist[lockTag]; ok && len(wl) > 0 {
		log.Debug("Found waitlist for ", lockTag)
		first := wl[0]

		if len(wl) == 1 {
			delete(multiQueue.waitlist, lockTag)
		} else {
			multiQueue.waitlist[lockTag] = wl[1:]
		}
		log.Debug("Resulting waitlist state:\n", multiQueue.waitlist)

		first.callback(first.lockTag)
	} else {
		log.Info("No waitlisted clients for lock tag: ", lockTag)
	}
}

// Get a queue index from an input hash to select which queue should handle an
// Enqueue(...) call.
func (multiQueue *multiQueue) queueIndexFromHash(hash uint16) uint16 {
	if hash == math.MaxUint16 {
		return uint16(len(multiQueue.queues)) - 1
	}

	return uint16(float32(hash) / (float32(MAX_HASH) / float32((len(multiQueue.queues)))))
}
