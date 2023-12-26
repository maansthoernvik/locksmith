package queue

import (
	"testing"
)

func Test_queueIndexDistribution(t *testing.T) {
	numQueues := 1000
	mq := &multiQueue{
		queues:   make([]chan *queueItem, numQueues),
		hashFunc: fnv1aHash,
	}

	numberOfTags := 1000000
	result := map[uint16]int{}

	for i := 0; i < numberOfTags; i++ {
		tag := randSeq(50)
		hash := mq.hashFunc(tag)
		index := mq.queueIndexFromHash(hash)

		if index == uint16(numQueues) {
			t.Error("Somehow hash", hash, "gave an index outside the range...")
		}

		v, ok := result[index]
		if !ok {
			result[index] = 1
		} else {
			result[index] = v + 1
		}
	}

	if len(result) != numQueues {
		t.Fatal("Length of result below number of queues, something is seriously wrong:", len(result))
	}

	previous := 0
	for _, v := range result {
		if previous == 0 {
			previous = v
			continue
		}

		// Tolerance is +- previous/10
		if v > (previous+(previous/5)) || v < (previous-(previous/5)) {
			t.Fatal("Distribution is outside tolerances, "+
				"something is wrong with the index calculator:\nprevious=", previous, "\nnew=", v)
		}
	}
}
