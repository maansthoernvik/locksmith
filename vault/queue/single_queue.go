package queue

import (
	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
)

type SingleQueue struct {
	queue        chan *queueItem
	waitlist     map[string][]waitListed
	synchronized Synchronized
}

type waitListed struct {
	client   string
	callback func(error)
}

type queueItem struct {
	action   protocol.ServerMessageType
	lockTag  string
	client   string
	callback func(error)
}

func NewSingleQueue(
	size int,
	synchronized Synchronized,
) QueueLayer {
	q := &SingleQueue{
		queue:        make(chan *queueItem, size),
		waitlist:     make(map[string][]waitListed),
		synchronized: synchronized,
	}
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
	_, ok := singleQueue.waitlist[lockTag]
	if !ok {
		singleQueue.waitlist[lockTag] = []waitListed{{client, callback}}
	} else {
		singleQueue.waitlist[lockTag] = append(singleQueue.waitlist[lockTag], waitListed{client, callback})
	}
	log.GlobalLogger.Debug("Resulting waitlist state:\n", singleQueue.waitlist)
}

func (singleQueue *SingleQueue) PopWaitlist(lockTag string) {
	log.GlobalLogger.Debug("Popping fom waitlist:", lockTag)
	if wl, ok := singleQueue.waitlist[lockTag]; ok && len(wl) > 0 {
		log.GlobalLogger.Debug("Found waitlist for", lockTag)
		first := wl[0]
		singleQueue.waitlist[lockTag] = wl[1:]
		singleQueue.synchronized.Synchronized(protocol.Acquire, lockTag, first.client, first.callback)
	}
	log.GlobalLogger.Debug("Resulting waitlist state:\n", singleQueue.waitlist)
}

func (singleQueue *SingleQueue) handlePop(qi *queueItem) {
	singleQueue.synchronized.Synchronized(qi.action, qi.lockTag, qi.client, qi.callback)
}
