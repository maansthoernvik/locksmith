package queue

import "github.com/maansthoernvik/locksmith/log"

type SingleQueue struct {
	queue        chan *queueItem
	waitlist     map[string][]*queueItem
	synchronized Synchronized
}

func NewSingleQueue(
	size int,
	synchronized Synchronized,
) QueueLayer {
	q := &SingleQueue{
		queue:        make(chan *queueItem, size),
		waitlist:     make(map[string][]*queueItem),
		synchronized: synchronized,
	}
	go func() {
		log.Info("Started single queue")
		for {
			qi := <-q.queue
			//log.Debug("Popped queue item")
			q.handlePop(qi)
		}
	}()
	return q
}

func (singleQueue *SingleQueue) Enqueue(lockTag string, callback func(string)) {
	//log.Debug("Queueing up for lock tag:", lockTag)
	singleQueue.queue <- &queueItem{lockTag: lockTag, callback: callback}
}

func (singleQueue *SingleQueue) Waitlist(lockTag string, callback func(string)) {
	log.Debug("Waitlisting client for lock tag: ", lockTag)
	_, ok := singleQueue.waitlist[lockTag]
	if !ok {
		singleQueue.waitlist[lockTag] = []*queueItem{{lockTag, callback}}
	} else {
		singleQueue.waitlist[lockTag] = append(singleQueue.waitlist[lockTag], &queueItem{lockTag, callback})
	}
	log.Debug("Resulting waitlist state:\n", singleQueue.waitlist)
}

func (singleQueue *SingleQueue) PopWaitlist(lockTag string) {
	log.Debug("Popping fom waitlist: ", lockTag)
	if wl, ok := singleQueue.waitlist[lockTag]; ok && len(wl) > 0 {
		log.Debug("Found waitlist for ", lockTag)
		first := wl[0]

		if len(wl) == 1 {
			delete(singleQueue.waitlist, lockTag)
		} else {
			singleQueue.waitlist[lockTag] = wl[1:]
		}
		log.Debug("Resulting waitlist state:\n", singleQueue.waitlist)

		singleQueue.handlePop(first)
	} else {
		log.Debug("No waitlisted clients for lock tag: ", lockTag)
	}
}

func (singleQueue *SingleQueue) handlePop(qi *queueItem) {
	singleQueue.synchronized.Synchronized(qi.lockTag, qi.callback)
}
