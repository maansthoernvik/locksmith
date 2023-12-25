package queue

import "github.com/maansthoernvik/locksmith/log"

type multiQueue struct {
	queue        chan *queueItem
	waitlist     map[string][]*queueItem
	synchronized Synchronized
}

func NewMultiQueue(
	size int,
	synchronized Synchronized,
) QueueLayer {
	q := &multiQueue{
		queue:        make(chan *queueItem, size),
		waitlist:     make(map[string][]*queueItem),
		synchronized: synchronized,
	}
	go func() {
		for {
			qi := <-q.queue
			log.Debug("Popped queue item")
			q.handlePop(qi)
		}
	}()
	return q
}

func (multiQueue *multiQueue) Enqueue(lockTag string, callback func(string)) {
	log.Debug("Queueing up for lock tag:", lockTag)
	multiQueue.queue <- &queueItem{lockTag: lockTag, callback: callback}
}

func (multiQueue *multiQueue) Waitlist(lockTag string, callback func(string)) {
	log.Debug("Waitlisting client for lock tag:", lockTag)
	_, ok := multiQueue.waitlist[lockTag]
	if !ok {
		multiQueue.waitlist[lockTag] = []*queueItem{{lockTag, callback}}
	} else {
		multiQueue.waitlist[lockTag] = append(multiQueue.waitlist[lockTag], &queueItem{lockTag, callback})
	}
	log.Debug("Resulting waitlist state:\n", multiQueue.waitlist)
}

func (multiQueue *multiQueue) PopWaitlist(lockTag string) {
	log.Debug("Popping fom waitlist:", lockTag)
	if wl, ok := multiQueue.waitlist[lockTag]; ok && len(wl) > 0 {
		log.Debug("Found waitlist for", lockTag)
		first := wl[0]

		if len(wl) == 1 {
			delete(multiQueue.waitlist, lockTag)
		} else {
			multiQueue.waitlist[lockTag] = wl[1:]
		}
		log.Debug("Resulting waitlist state:\n", multiQueue.waitlist)

		multiQueue.handlePop(first)
	} else {
		log.Debug("No waitlisted clients for lock tag:", lockTag)
	}
}

func (multiQueue *multiQueue) handlePop(qi *queueItem) {
	multiQueue.synchronized.Synchronized(qi.lockTag, qi.callback)
}
