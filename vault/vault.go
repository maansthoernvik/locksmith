package vault

import (
	"errors"

	"github.com/maansthoernvik/locksmith/env"
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

var logger *log.Logger

func init() {
	logLevel, _ := env.GetOptionalString(
		env.LOCKSMITH_LOG_LEVEL, env.LOCKSMITH_LOG_LEVEL_DEFAULT,
	)
	logger = log.New(log.Translate(logLevel))
}

type Vault interface {
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

type VaultImpl struct {
	queueLayer        queue.QueueLayer
	state             map[string]lockInfo
	clientLookUpTable map[string][]string
}

type QueueType string

const (
	Single QueueType = "single"
	Multi  QueueType = "multi"
)

type VaultOptions struct {
	QueueType
	QueueConcurrency int
	QueueCapacity    int
}

func NewVault(options *VaultOptions) Vault {
	vaultImpl := &VaultImpl{
		state:             make(map[string]lockInfo),
		clientLookUpTable: make(map[string][]string),
	}
	if options.QueueType == Single {
		vaultImpl.queueLayer = queue.NewSingleQueue(
			options.QueueCapacity, vaultImpl,
		)
	} else {
		vaultImpl.queueLayer = queue.NewMultiQueue(
			options.QueueConcurrency, options.QueueCapacity, vaultImpl,
		)
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
	callback func(error) error,
) {
	logger.Debug("Client", client, "acquiring", lockTag)
	vaultImpl.queueLayer.Enqueue(
		lockTag, vaultImpl.acquireAction(client, callback),
	)
}

func (vaultImpl *VaultImpl) acquireAction(
	client string,
	callback func(error) error,
) func(lockTag string) {
	return func(lockTag string) {
		currentState, ok := vaultImpl.state[lockTag]
		if !ok {
			vaultImpl.state[lockTag] = lockInfo{
				client: "", lockState: UNLOCKED,
			}
			currentState = vaultImpl.state[lockTag]
		}
		// a second acquire is a protocol offense, callback with error and
		// release the lock, pop waitlisted client.
		if currentState.client == client {
			currentState.client = ""
			currentState.lockState = UNLOCKED
			vaultImpl.state[lockTag] = currentState
			_ = callback(UnecessaryAcquireError)

			if lts, ok := vaultImpl.clientLookUpTable[client]; ok {
				newLts := make([]string, len(lts)-1)
				for _, lt := range lts {
					if lt != lockTag {
						newLts = append(newLts, lt)
					}
				}
				vaultImpl.clientLookUpTable[client] = newLts
			}

			vaultImpl.queueLayer.PopWaitlist(lockTag)
			// client didn't match, and the lock state is LOCKED, waitlist the
			// client
		} else if currentState.lockState == LOCKED {
			vaultImpl.queueLayer.Waitlist(
				lockTag, vaultImpl.acquireAction(client, callback),
			)
		} else {
			// This means a write failure occurred and the client that was
			// acquiring the lock has NW issues or something.
			if err := callback(nil); err != nil {
				// don't touch the lock state, pop from waitlist
				vaultImpl.queueLayer.PopWaitlist(lockTag)
			} else {
				currentState.client = client
				currentState.lockState = LOCKED
				vaultImpl.state[lockTag] = currentState
				if _, ok := vaultImpl.clientLookUpTable[client]; !ok {
					vaultImpl.clientLookUpTable[client] = []string{lockTag}
				} else {
					vaultImpl.clientLookUpTable[client] = append(
						vaultImpl.clientLookUpTable[client],
						lockTag,
					)
				}
			}
		}
	}
}

// Release releases a lock, leading to a queued acquire calling the vault
// callback.
func (vaultImpl *VaultImpl) Release(
	lockTag string,
	client string,
	callback func(error) error,
) {
	logger.Debug("Client", client, "releasing", lockTag)
	vaultImpl.queueLayer.Enqueue(lockTag, vaultImpl.releaseAction(client, callback))
}

func (vaultImpl *VaultImpl) releaseAction(
	client string,
	callback func(error) error,
) func(lockTag string) {
	return func(lockTag string) {
		currentState, ok := vaultImpl.state[lockTag]
		if !ok {
			vaultImpl.state[lockTag] = lockInfo{client: "", lockState: UNLOCKED}
			currentState = vaultImpl.state[lockTag]
		}
		// if already unlocked, kill the client for not following the protocol
		if currentState.lockState == UNLOCKED {
			_ = callback(UnecessaryReleaseError)
			// else, the lock is in LOCKED state, so check the owner, if
			// client isn't the owner, it's misbehaving and needs to be killed
		} else if currentState.client != client {
			_ = callback(BadMannersError)
			// else, client is the owner of the lock, release it and call
			// callback
		} else {
			currentState.client = ""
			currentState.lockState = UNLOCKED
			vaultImpl.state[lockTag] = currentState
			_ = callback(nil)

			if lts, ok := vaultImpl.clientLookUpTable[client]; ok {
				newLts := make([]string, len(lts)-1)
				for _, lt := range lts {
					if lt != lockTag {
						newLts = append(newLts, lt)
					}
				}
				vaultImpl.clientLookUpTable[client] = newLts
			}

			vaultImpl.queueLayer.PopWaitlist(lockTag)
		}
	}
}

func (vaultImpl *VaultImpl) Cleanup(client string) {
	logger.Info("Cleaning up after client:", client)
	lockTags := vaultImpl.clientLookUpTable[client]

	for _, lockTag := range lockTags {
		vaultImpl.queueLayer.Enqueue(
			lockTag, func(lockTag string) {
				if currentState, ok := vaultImpl.state[lockTag]; ok &&
					currentState.client == client {
					currentState.client = ""
					currentState.lockState = UNLOCKED
					vaultImpl.state[lockTag] = currentState
					vaultImpl.queueLayer.PopWaitlist(lockTag)
				}
				delete(vaultImpl.clientLookUpTable, client)
			},
		)
	}
}

// This member function is the only function allowed to touch the vault's lock
// states. It is called from the queue layer after a dispatch via Enqueue().
func (vaultImpl *VaultImpl) Synchronized(
	lockTag string,
	action func(string),
) {
	logger.Info("Entering synchronized access block for lock tag", lockTag)
	action(lockTag)
	logger.Debug("Resulting vault state: \n", vaultImpl.state)
}
