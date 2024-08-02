package vault

import (
	"errors"
	"sync"
	"testing"

	"github.com/maansthoernvik/locksmith/pkg/vault/queue"
)

type testQueueLayer struct {
	vault *vaultImpl
}

func (tql *testQueueLayer) Enqueue(lockTag string, action queue.SynchronizedAction) {
	tql.vault.Synchronized(lockTag, action)
}

func (tql *testQueueLayer) Waitlist(lockTag string, action func(string)) {
	// noop
}

func (tql *testQueueLayer) PopWaitlist(lockTag string) {
	// noop
}

func Test_Acquire(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	called := false
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire callback called!")
		called = true
		return nil
	})

	if !called {
		t.Error("Acquire callback wasn't called")
	}

	if li := v.fetch("lt"); !(li.isOwner("client") && li.isLocked()) {
		t.Error("Expected client to be the owner and the lock to be locked")
	}
}

func Test_Release(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire callback called!")
		return nil
	})

	called := false
	v.Release("lt", "client", func(err error) error {
		t.Log("Release callback called!")
		called = true
		return nil
	})

	if !called {
		t.Error("Release callback wasn't called")
	}

	if li := v.fetch("lt"); li.isLocked() {
		t.Error("Expected the lock to not be.... well locked")
	}
}

func Test_Waitlist(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		waitList:          make(map[string][]*func(string)),
		clientLookUpTable: make(map[string][]string),
	}
	// Use single queue for waitlist functionality
	v.queueLayer = &testQueueLayer{vault: v}

	order := make([]string, 0, 3)

	wg := sync.WaitGroup{}
	wg.Add(3)
	v.Acquire("lt", "client1", func(err error) error {
		t.Log("Acquire client1 callback called!")
		wg.Done()
		order = append(order, "client1")

		return nil
	})
	v.Acquire("lt", "client2", func(err error) error {
		t.Log("Acquire client2 callback called!")
		wg.Done()
		order = append(order, "client2")

		return nil
	})
	v.Release("lt", "client1", func(err error) error {
		t.Log("Release client1 callback called!")
		wg.Done()
		order = append(order, "client1")

		return nil
	})
	wg.Wait()

	t.Log(order)

	// Check order of operations...
	if order[0] != "client1" {
		t.Error("First operation was not client1")
	}
	if order[1] != "client1" {
		t.Error("Second operation was not client1")
	}
	if order[2] != "client2" {
		t.Error("Third operation was not client2")
	}
}

func Test_ReleaseBadManners(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	v.Acquire("lt", "client1", func(err error) error {
		t.Log("Acquire client1 callback called!")
		return nil
	})
	v.Release("lt", "client2", func(err error) error {
		t.Log("Release client2 callback called with error:", err)
		if !errors.Is(err, ErrBadManners) {
			t.Error("Expected BadMannersError")
		}
		return nil
	})
}

func Test_UnecessaryRelease(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	v.Release("lt", "client", func(err error) error {
		t.Log("Release client callback called with error:", err)
		if !errors.Is(err, ErrUnnecessaryRelease) {
			t.Error("Expected UnecessaryReleaseError")
		}

		return nil
	})
}

func Test_UnecessaryAcquire(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		return nil
	})
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		if !errors.Is(err, ErrUnnecessaryAcquire) {
			t.Error("Expected UnecesasryAcquireError")
		}

		return nil
	})
}

func Test_CallbackError(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		// Because of the returned error, another client is able to acquire the lock
		return errors.New("some kind of error")
	})
	if l, ok := v.state["lt"]; ok {
		if l.owner != "" || l.state != UNLOCKED {
			t.Error("Unexpected lock state")
		}
	}

	v.Acquire("lt", "client2", func(err error) error {
		t.Log("Acquire client2 callback called with error:", err)
		return nil
	})

	if l, ok := v.state["lt"]; ok {
		if l.owner != "client2" || l.state != LOCKED {
			t.Error("Expected client2 to have acquired the lock")
		}
	}
}

func Test_Cleanup(t *testing.T) {
	v := &vaultImpl{
		state:             make(map[string]*lock),
		clientLookUpTable: make(map[string][]string),
	}
	v.queueLayer = &testQueueLayer{vault: v}

	t.Log("Initial lookup table state: ", v.clientLookUpTable)

	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire lt client callback called with error:", err)
		return nil
	})
	t.Log(v.clientLookUpTable)

	v.Acquire("lt2", "client", func(err error) error {
		t.Log("Acquire lt2 client callback called with error:", err)
		return nil
	})
	t.Log(v.clientLookUpTable)

	v.Acquire("lt3", "client", func(err error) error {
		t.Log("Acquire lt3 client callback called with error:", err)
		return nil
	})
	t.Log(v.clientLookUpTable)

	t.Log("Checking who owns 'lt'...")
	l := v.state["lt"]
	if l.owner != "client" && l.state != LOCKED {
		t.Fatal("client does not have acquired lock")
	}

	lts, ok := v.clientLookUpTable["client"]
	if !ok {
		t.Fatal("Could not find client in client lookup table")
	}

	if len(lts) != 3 {
		t.Fatal("Unexpected length of lock tags owner by client")
	}
	t.Log("Lock states before release: ", v.state)

	v.Release("lt3", "client", func(err error) error {
		t.Log("Release lt3 client callback called with error:", err)
		return nil
	})
	t.Log(v.clientLookUpTable)

	lts, ok = v.clientLookUpTable["client"]
	if !ok {
		t.Fatal("Could not find client in client lookup table")
	}

	if len(lts) != 2 {
		t.Fatal("Unexpected length of lock tags owned by client")
	}

	v.Cleanup("client")
	if v.state["lt"].owner != "" && v.state["lt"].state != UNLOCKED {
		t.Error("Cleanup wasn't successful")
	}
	if v.state["lt2"].owner != "" && v.state["lt"].state != UNLOCKED {
		t.Error("Cleanup wasn't successful")
	}
	if v.state["lt3"].owner != "" && v.state["lt"].state != UNLOCKED {
		t.Error("Cleanup wasn't successful")
	}

	_, ok = v.clientLookUpTable["client"]
	if ok {
		t.Fatal("Client lookup table should have been cleared")
	}
	t.Log("Resulting lookup table after cleanup: ", v.clientLookUpTable)
	t.Log("Resulting state after cleanup: ", v.state)
}
