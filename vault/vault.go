// Package vault solves the handling of mutexes.
package vault

import (
	"errors"
	"fmt"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/vault/queue"
)

var BadMannersError = errors.New(
	"Client tried to release lock that it did not own",
)
var UnecessaryReleaseError = errors.New(
	"Client tried to release a lock that had not been acquired",
)
var UnecessaryAcquireError = errors.New(
	"Client tried to acquire a lock that it already had acquired",
)

// The Vault interface specifies high level functions to implement in order to
// handle the acquisition and release of mutexes.
type Vault interface {
	// Lock tag is a string identifying the lock to acquire, client the requesting party,
	// and the callback a function which will be called to either confirm acquisition or
	// including an error in case the client is misbehaving. The callback may return an
	// error in case feedback handling encounters an error.
	Acquire(lockTag string, client string, callback func(error) error)
	Release(lockTag string, client string, callback func(error) error)
	Cleanup(client string)
}

type lockState bool

const (
	LOCKED   lockState = true
	UNLOCKED lockState = false
)

type lockInfo struct {
	client string
	lockState
}

func newLockInfo() *lockInfo {
	return &lockInfo{client: "", lockState: UNLOCKED}
}

func (li *lockInfo) isOwner(client string) bool {
	return li.client == client
}

func (li *lockInfo) isLocked() bool {
	return li.lockState == LOCKED
}

func (li *lockInfo) unlock() {
	li.lockState = UNLOCKED
	li.client = ""
}

func (li *lockInfo) lock(client string) {
	li.lockState = LOCKED
	li.client = client
}

func (li *lockInfo) String() string {
	return fmt.Sprintf("&lockInfo{c: %s, s: %v}", li.client, li.lockState)
}

// Implementation of the Vault interface. By use of a queue layer, the vault ensures
// lock states are only manipulated from one Go-routine at a time. Read more in the
// QueueLayer interface description.
type vaultImpl struct {
	queueLayer queue.QueueLayer
	state      map[string]*lockInfo
	// Used to keep track of which locks a client owns without having to iterate over
	// all of them. Used when clients disconnect to release locks held by them.
	clientLookUpTable map[string][]string
}

type QueueType string

const (
	Single QueueType = "single"
	Multi  QueueType = "multi"
)

type VaultOptions struct {
	// Single queue mode should only be used for testing.
	QueueType

	// Only for multi-mode queues, determines the number of
	// supporting Go-routines able to handle work given to the
	// queueing layer.
	QueueConcurrency int

	// Sets the capacity of the underlying queue(s), the max amount
	// of buffered work for a queue. In a multi queue setting, the
	// capacity indicates the buffer size per queue.
	QueueCapacity int
}

func NewVault(options *VaultOptions) Vault {
	vault := &vaultImpl{
		state:             make(map[string]*lockInfo),
		clientLookUpTable: make(map[string][]string),
	}
	if options.QueueType == Single {
		vault.queueLayer = queue.NewSingleQueue(
			options.QueueCapacity, vault,
		)
	} else {
		vault.queueLayer = queue.NewMultiQueue(
			options.QueueConcurrency, options.QueueCapacity, vault,
		)
	}

	return vault
}

// Acquire attempts to acquire a lock. If the lock is currently busy, the
// request in put on the queue for the lock tag in question, leading to a
// notification once the holder has either released the lock or the lock
// timeout hits.
func (vault *vaultImpl) Acquire(
	lockTag string,
	client string,
	callback func(error) error,
) {
	log.Info("Client ", client, " acquiring ", lockTag)
	vault.queueLayer.Enqueue(
		lockTag, vault.acquireAction(client, callback),
	)
}

// Returns a callback to call once the vault has gotten hold of a
// synchronization Go-routine. The returned action callback contains
// handling for what should happen when a client requests to acquire
// a lock. The returned callback is the only piece of code allowed to
// handle acquiring locks.
func (vault *vaultImpl) acquireAction(
	client string,
	callback func(error) error,
) queue.SynchronizedAction {
	return func(lockTag string) {
		currentState := vault.fetch(lockTag)
		// a second acquire is a protocol offense, callback with error and
		// release the lock, pop waitlisted client.
		if currentState.isOwner(client) {
			currentState.unlock()
			_ = callback(UnecessaryAcquireError)

			vault.queueLayer.PopWaitlist(lockTag)
			// client didn't match, and the lock state is LOCKED, waitlist the
			// client
		} else if currentState.isLocked() {
			vault.queueLayer.Waitlist(
				lockTag, vault.acquireAction(client, callback),
			)
		} else {
			// This means a write failure occurred and the client that was
			// acquiring the lock has NW issues or something.
			if err := callback(nil); err != nil {
				// don't touch the lock state, pop from waitlist
				vault.queueLayer.PopWaitlist(lockTag)
			} else {
				currentState.lock(client)
				vault.appendClientLookupTable(client, lockTag)
			}
		}
	}
}

// Release releases a lock, leading to a queued acquire calling the vault
// callback.
func (vault *vaultImpl) Release(
	lockTag string,
	client string,
	callback func(error) error,
) {
	log.Info("Client ", client, " releasing ", lockTag)
	vault.queueLayer.Enqueue(lockTag, vault.releaseAction(client, callback))
}

// Returns a callback that handles the release of locks. This is the only piece
// of code allowed to touch release-handling, similarily to the acquireAction
// function. The returned function must only be called from the scope of a
// synchronization Go-routine.
func (vault *vaultImpl) releaseAction(
	client string,
	callback func(error) error,
) queue.SynchronizedAction {
	return func(lockTag string) {
		currentState := vault.fetch(lockTag)
		// if already unlocked, kill the client for not following the protocol
		if !currentState.isLocked() {
			_ = callback(UnecessaryReleaseError)
			// else, the lock is in LOCKED state, so check the owner, if
			// client isn't the owner, it's misbehaving and needs to be killed
		} else if !currentState.isOwner(client) {
			_ = callback(BadMannersError)
			// else, client is the owner of the lock, release it and call
			// callback
		} else {
			currentState.unlock()
			_ = callback(nil) // We don't care about release errors

			vault.cleanClientLookupTable(client, lockTag)

			vault.queueLayer.PopWaitlist(lockTag)
		}
	}
}

// Cleans up all information associated with a given client.
func (vault *vaultImpl) Cleanup(client string) {
	log.Info("Cleaning up after client: ", client)
	lockTags := vault.clientLookUpTable[client]

	for _, lockTag := range lockTags {
		vault.queueLayer.Enqueue(
			lockTag, vault.cleanupAction(client),
		)
	}
	delete(vault.clientLookUpTable, client)
}

// Returns a callback that handles the cleanup of a client for a given lock tag.
// This function must only be called from the scope of a synchronization
// Go-routine, because just like the acquire- and releaseAction functions, it
// handles the vault's lock states.
func (vault *vaultImpl) cleanupAction(client string) queue.SynchronizedAction {
	return func(lockTag string) {
		if currentState := vault.fetch(lockTag); currentState.isOwner(client) {
			currentState.unlock()
			vault.queueLayer.PopWaitlist(lockTag)
		}
	}
}

// This member function is the only function allowed to touch the vault's lock
// states. It is called from the queue layer after a dispatch via Enqueue().
func (vault *vaultImpl) Synchronized(
	lockTag string,
	action queue.SynchronizedAction,
) {
	log.Debug("Entering synchronized access block for lock tag ", lockTag)
	action(lockTag)
	log.Debug("Resulting vault state: \n", vault.state)
}

func (vault *vaultImpl) fetch(lockTag string) *lockInfo {
	li, ok := vault.state[lockTag]
	if !ok {
		li = newLockInfo()
		vault.state[lockTag] = li
	}

	return li
}

// Add a lock to a client's lookup table.
func (vault *vaultImpl) appendClientLookupTable(client, lockTag string) {
	if _, ok := vault.clientLookUpTable[client]; !ok {
		vault.clientLookUpTable[client] = []string{lockTag}
	} else {
		vault.clientLookUpTable[client] = append(vault.clientLookUpTable[client], lockTag)
	}
}

// Remove a lock from a client's lookup table.
func (vault *vaultImpl) cleanClientLookupTable(client, lockTag string) {
	if lts, ok := vault.clientLookUpTable[client]; ok {
		newLts := make([]string, 0, len(lts)-1)
		for _, lt := range lts {
			if lt != lockTag {
				newLts = append(newLts, lt)
			}
		}
		vault.clientLookUpTable[client] = newLts
	}
}
