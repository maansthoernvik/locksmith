package queuetest

import (
	"math/rand"
)

type testSynchronized struct {
	callCount int
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func (ts *testSynchronized) Synchronized(lockTag string, callback func(string)) {
	ts.callCount++
	//log.Println("Synchronized called, call count =", ts.callCount)
	callback(lockTag)
}

func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
