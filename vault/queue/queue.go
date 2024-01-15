// Package queue implements a way for the vault to ensure single-threaded handling of mutexes.
package queue

// The vault queue layer interface specifies a way for the vault to obtain
// a Go-routine for a given lock tag. Once the queue layer notifies the vault,
// the vault knows that for a given lock tag, it is safe to operate as the
// thread assigned to that lock tag is the one making the notifying call. The
// notifying call is made to an entity implementing the Synchronized interface
// below. One implementor being the vault.
//
// The queueing layer also exposes a waiting list functionality. The vault is
// able to waitlist operations if need be, for example in case a lock tag is
// already taken. The vault may also pop from the wait list, telling the queue
// layer to remove an item from the front of the waiting list. Calls to Waitlist
// and PopWaitlist must be made from a synchronization thread as they alter the
// state of a lock-specific storage of waiting clients.
type QueueLayer interface {
	// Request a Go-routine for the given lock tag.
	Enqueue(lockTag string, callback func(lockTag string))
	// Put callback in the waiting list for the given lock tag.
	Waitlist(lockTag string, callback func(lockTag string))
	// Pop from the waiting list, making a previously waitlisted callback be called.
	PopWaitlist(lockTag string)
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
