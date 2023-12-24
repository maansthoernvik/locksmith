package vault

import (
	"errors"

	"github.com/maansthoernvik/locksmith/log"
	"github.com/maansthoernvik/locksmith/protocol"
	"github.com/maansthoernvik/locksmith/vault/queue"
)

var BadMannersError = errors.New("Client tried to release lock that it did not own")
var UnecessaryReleaseError = errors.New("Client tried to release a lock that had not been acquired")
var UnecessaryAcquireError = errors.New("Client tried to acquire a lock that it already had acquired")

type Vault interface {
	Acquire(lockTag string, client string, callback func(error))
	Release(lockTag string, client string, callback func(error))
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

type VaultImpl struct {
	queueLayer queue.QueueLayer
	state      map[string]lockInfo
}

type QueueType string

const (
	Single QueueType = "single"
	Multi  QueueType = "multi"
)

type VaultOptions struct {
	QueueType
}

func NewVault(vaultOptions *VaultOptions) Vault {
	vaultImpl := &VaultImpl{state: make(map[string]lockInfo)}
	if vaultOptions.QueueType == Single {
		vaultImpl.queueLayer = queue.NewSingleQueue(300, vaultImpl.synchronizedLockTagAccess)
	} else if vaultOptions.QueueType == Multi {
		panic("Multi queue type not implemented")
	} else {
		vaultImpl.queueLayer = queue.NewSingleQueue(300, vaultImpl.synchronizedLockTagAccess)
	}

	return vaultImpl
}

// Acquire attempts to acquire a lock. If the lock is currently busy, the
// request in put on the queue for the lock tag in question, leading to a
// notification once the holder has either released the lock or the lock
// timeout hits.
func (vaultImpl *VaultImpl) Acquire(
	lockTag string,
	client string,
	callback func(error),
) {
	log.GlobalLogger.Debug("Client", client, "acquiring", lockTag)
	vaultImpl.enqueue(protocol.Acquire, lockTag, client, callback)
}

// Release releases a lock, leading to a queued acquire calling the vault
// callback.
func (vaultImpl *VaultImpl) Release(
	lockTag string,
	client string,
	callback func(error),
) {
	log.GlobalLogger.Debug("Client", client, "releasing", lockTag)
	vaultImpl.enqueue(protocol.Release, lockTag, client, callback)
}

func (vaultImpl *VaultImpl) enqueue(
	action protocol.ServerMessageType,
	lockTag string,
	client string,
	callback func(error),
) {
	log.GlobalLogger.Debug("Calling queue layer")
	vaultImpl.queueLayer.Enqueue(action, lockTag, client, callback)
}

// This member function is the only function allowed to touch the vault's lock
// states. It is called from the queue layer after a dispatch via Enqueue().
func (vaultImpl *VaultImpl) synchronizedLockTagAccess(
	action protocol.ServerMessageType,
	client string,
	lockTag string,
	callback func(error),
) {
	log.GlobalLogger.Debug("Entering synchronized access block for lock tag", lockTag, "on behalf of client", client)
	currentState, ok := vaultImpl.state[lockTag]
	if !ok {
		vaultImpl.state[lockTag] = lockInfo{client: "", lockState: UNLOCKED}
		currentState = vaultImpl.state[lockTag]
	}

	switch action {
	case protocol.Acquire:
		// a second acquire is a protocol offense, callback with error and
		// release the lock, pop waitlisted client.
		if currentState.client == client {
			callback(UnecessaryAcquireError)
			// client didn't match, and the lock state is LOCKED, waitlist the
			// client
		} else if currentState.lockState == LOCKED {
			vaultImpl.queueLayer.Waitlist(lockTag, client, callback)
		} else {
			currentState.client = client
			currentState.lockState = LOCKED
			vaultImpl.state[lockTag] = currentState
			callback(nil)
		}
	case protocol.Release:
		// if already unlocked, kill the client for not following the protocol
		if currentState.lockState == UNLOCKED {
			callback(UnecessaryReleaseError)
			// else, the lock is in LOCKED state, so check the owner, if
			// client isn't the owner, it's misbehaving and needs to be killed
		} else if currentState.client != client {
			callback(BadMannersError)
			// else, client is the owner of the lock, release it and call
			// callback
		} else {
			currentState.client = ""
			currentState.lockState = UNLOCKED
			vaultImpl.state[lockTag] = currentState
			callback(nil)
		}
	}
	log.GlobalLogger.Debug("Resulting vault state: \n", vaultImpl.state)
}
