package queue

import "github.com/rs/zerolog/log"

type SingleQueue struct {
	queue        chan *queueItem
	synchronized Synchronized
}

func NewSingleQueue(
	size int,
	synchronized Synchronized,
) QueueLayer {
	q := &SingleQueue{
		queue:        make(chan *queueItem, size),
		synchronized: synchronized,
	}
	go func() {
		log.Info().Msg("started single queue")
		for {
			qi := <-q.queue
			//log.Debug("Popped queue item")
			q.synchronized.Synchronized(qi.lockTag, qi.callback)
		}
	}()
	return q
}

func (singleQueue *SingleQueue) Enqueue(lockTag string, callback SynchronizedAction) {
	//log.Debug("Queueing up for lock tag:", lockTag)
	singleQueue.queue <- &queueItem{lockTag: lockTag, callback: callback}
}
