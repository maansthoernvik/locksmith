// Package queue implements a way for the vault to ensure single-threaded handling of mutexes.
package queue

// The vault queue layer interface specifies a way for the vault to obtain
// a Go-routine for a given lock tag. Once the queue layer notifies the vault,
// the vault knows that for a given lock tag, it is safe to operate as the
// thread assigned to that lock tag is the one making the notifying call. The
// notifying call is made to an entity implementing the Synchronized interface
// below. One implementor being the vault.
type QueueLayer interface {
	// Request a Go-routine for the given lock tag.
	Enqueue(lockTag string, action SynchronizedAction)
}

type SynchronizedAction func(lockTag string)

// Interface for calls made by a synchronization thread in the queue layer.
type Synchronized interface {
	Synchronized(lockTag string, action SynchronizedAction)
}

type queueItem struct {
	lockTag  string
	callback func(string)
}
