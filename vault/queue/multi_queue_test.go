package queue

import "testing"

func Test_queueIndexFromHash(t *testing.T) {
	mq := &multiQueue{queues: make([]chan *queueItem, 10)}

	var hash uint16 = 0
	index := mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 0 {
		t.Error("Expected hash 0 to have index 0, but it has index:", index)
	}

	hash = 65535
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 9 {
		t.Error("Expected hash 65535 to have index 9, but it has index:", index)
	}

	hash = 65535 / 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 1 {
		t.Error("Expected hash", hash, "to have index 1, but it has index:", index)
	}

	hash = 65535/10 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 1 {
		t.Error("Expected hash", hash, "to have index 1, but it has index:", index)
	}

	hash = 65535/10*2 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 2 {
		t.Error("Expected hash", hash, "to have index 2, but it has index:", index)
	}

	hash = 65535/10*3 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 3 {
		t.Error("Expected hash", hash, "to have index 3, but it has index:", index)
	}

	hash = 65535/10*4 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 4 {
		t.Error("Expected hash", hash, "to have index 4, but it has index:", index)
	}

	hash = 65535/10*5 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 5 {
		t.Error("Expected hash", hash, "to have index 1, but it has index:", index)
	}

	hash = 65535/10*6 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 6 {
		t.Error("Expected hash", hash, "to have index 6, but it has index:", index)
	}

	hash = 65535/10*7 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 7 {
		t.Error("Expected hash", hash, "to have index 7, but it has index:", index)
	}

	hash = 65535/10*8 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 8 {
		t.Error("Expected hash", hash, "to have index 8, but it has index:", index)
	}

	hash = 65535/10*9 + 10
	index = mq.queueIndexFromHash(hash)
	t.Log("Hash", hash, "got index", index)
	if index != 9 {
		t.Error("Expected hash", hash, "to have index 9, but it has index:", index)
	}
}
