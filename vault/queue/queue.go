package queue

import "github.com/maansthoernvik/locksmith/protocol"

// The vault queue layer interface specifies a way for the vault to obtain
// a goroutine for a given lock tag. Once the queue layer notifies callback,
// the vault can be sure that no other concurrent process is currently
// accessing the given lock tag.
type QueueLayer interface {
	Enqueue(action protocol.ServerMessageType, lockTag string, client string, callback func(error))
	Waitlist(lockTag string, client string, callback func(error))
	PopWaitlist(lockTag string)
}
