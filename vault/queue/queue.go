package queue

// The vault queue layer interface specifies a way for the vault to obtain
// a goroutine for a given lock tag. Once the queue layer notifies callback,
// the vault can be sure that no other concurrent process is currently
// accessing the given lock tag.
type QueueLayer interface {
	Enqueue(lockTag string, callback func(lockTag string))
	Waitlist(lockTag string, callback func(lockTag string))
	PopWaitlist(lockTag string)
}

// Interface for calls made by the synchronization thread in the queue layer.
type Synchronized interface {
	Synchronized(lockTag string, callback func(lockTag string))
}

type queueItem struct {
	lockTag  string
	callback func(string)
}
