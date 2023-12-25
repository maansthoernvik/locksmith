package vault

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/maansthoernvik/locksmith/vault/queue"
)

func Test_AcquireCallback(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(1)
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire callback called!")
		wg.Done()

		return nil
	})
	wg.Wait()
}

func Test_ReleaseCallback(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(1)
	v.Release("lt", "client", func(err error) error {
		t.Log("Release callback called!")
		wg.Done()

		return nil
	})
	wg.Wait()
}

func Test_AcquireWaitlist(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(3)
	v.Acquire("lt", "client1", func(err error) error {
		t.Log("Acquire client1 callback called!")
		wg.Done()

		return nil
	})
	v.Acquire("lt", "client2", func(err error) error {
		t.Log("Acquire client2 callback called!")
		wg.Done()

		return nil
	})
	v.Release("lt", "client1", func(err error) error {
		t.Log("Release client1 callback called!")
		wg.Done()

		return nil
	})
	wg.Wait()
}

func Test_ReleaseBadManners(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(2)
	v.Acquire("lt", "client1", func(err error) error {
		t.Log("Acquire client1 callback called!")
		wg.Done()

		return nil
	})
	v.Release("lt", "client2", func(err error) error {
		t.Log("Release client2 callback called with error:", err)
		if errors.Is(err, BadMannersError) {
			wg.Done()
		}
		return nil
	})
	wg.Wait()
}

func Test_UnecessaryRelease(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(1)
	v.Release("lt", "client", func(err error) error {
		t.Log("Release client callback called with error:", err)
		if errors.Is(err, UnecessaryReleaseError) {
			wg.Done()
		}

		return nil
	})
	wg.Wait()
}

func Test_UnecessaryAcquire(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(2)
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		if err == nil {
			wg.Done()
		}

		return nil
	})
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		if errors.Is(err, UnecessaryAcquireError) {
			wg.Done()
		}

		return nil
	})
	wg.Wait()
}

func Test_CallbackError(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(1)
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		if err == nil {
			wg.Done()
		}

		// Because of the returned error, another client is able to acquire the lock
		return errors.New("some kind of error")
	})
	wg.Wait()

	wg.Add(1)
	v.Acquire("lt", "client2", func(err error) error {
		t.Log("Acquire client2 callback called with error:", err)
		if err == nil {
			wg.Done()
		}

		return nil
	})

	wg.Wait()
}

func Test_Cleanup(t *testing.T) {
	v := &VaultImpl{state: make(map[string]lockInfo), clientLookUpTable: make(map[string][]string)}
	v.queueLayer = queue.NewSingleQueue(1, v)
	wg := sync.WaitGroup{}
	wg.Add(1)
	v.Acquire("lt", "client", func(err error) error {
		t.Log("Acquire client callback called with error:", err)
		if err == nil {
			wg.Done()
		}

		return nil
	})
	wg.Wait()

	li := v.state["lt"]
	t.Log(li)
	t.Log(v.clientLookUpTable)
	if li.client != "client" {
		t.Error("client does not have acquired lock")
	}

	v.Cleanup("client")

	for {
		time.Sleep(50 * time.Millisecond)
		if v.state["lt"].client == "" {
			break
		}
	}
	t.Log(li)
	t.Log(v.clientLookUpTable)
}
