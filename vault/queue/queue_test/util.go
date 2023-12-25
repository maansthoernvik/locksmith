package queuetest

import "log"

type testSynchronized struct {
	callCount int
}

func (ts *testSynchronized) Synchronized(lockTag string, callback func(string)) {
	ts.callCount++
	log.Println("Synchronized called, call count =", ts.callCount)
	callback(lockTag)
}
