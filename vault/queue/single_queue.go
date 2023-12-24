package queue

import (
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
)

type SingleQueue struct {
	queue                chan *queueItem
	synchronizedCallback func(protocol.ServerMessageType, string, string, func(error))
}

type queueItem struct {
	action   protocol.ServerMessageType
	lockTag  string
	client   string
	callback func(error)
}

func NewSingleQueue(size int, synchronizedCallback func(protocol.ServerMessageType, string, string, func(error))) QueueLayer {
	q := &SingleQueue{queue: make(chan *queueItem, size), synchronizedCallback: synchronizedCallback}
	go func() {
		for {
			qi := <-q.queue
			log.GlobalLogger.Debug("Popped queue item, handling lock tag", qi.lockTag, "for client", qi.client)
			q.handlePop(qi)
		}
	}()
	return q
}

func (singleQueue *SingleQueue) Enqueue(action protocol.ServerMessageType, lockTag string, client string, callback func(error)) {
	log.GlobalLogger.Debug("Queueing up client", client, "for lock tag", lockTag)
	singleQueue.queue <- &queueItem{action: action, lockTag: lockTag, client: client, callback: callback}
}

func (singleQueue *SingleQueue) Waitlist(lockTag string, client string, callback func(error)) {
	log.GlobalLogger.Debug("Waitlisting client", client, "for lock", lockTag)
}

func (singleQueue *SingleQueue) handlePop(qi *queueItem) {
	singleQueue.synchronizedCallback(qi.action, qi.client, qi.lockTag, qi.callback)
}
