package queue

import (
	"math/rand"
	"testing"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

const RANGES = 4
const SAMPLE_SIZE = 100000
const SEQUENCE_SIZE = 100
const MAX = 65535

func TestRangeDistribution_crc16(t *testing.T) {
	t.Skip()
	distributionResult := make([]uint32, RANGES)
	for i := 0; i < SAMPLE_SIZE; i++ {
		n := crc16(randSeq(SEQUENCE_SIZE))
		if n < MAX/RANGES {
			distributionResult[0]++
		} else if n >= MAX/RANGES && n <= MAX/2 {
			distributionResult[1]++
		} else if n > MAX/2 && n < MAX-(MAX/RANGES) {
			distributionResult[2]++
		} else {
			distributionResult[3]++
		}
	}
	var previous uint32 = distributionResult[0]
	for _, val := range distributionResult[1:] {
		if val > (previous+1000) || val < (previous-1000) {
			t.Error("Distribution outside the allowed bounds")
		}
	}
	t.Log(distributionResult)
}

func TestRangeDistribution_customHash(t *testing.T) {
	t.Skip()
	distributionResult := make([]uint32, RANGES)
	for i := 0; i < SAMPLE_SIZE; i++ {
		n := customHash(randSeq(SEQUENCE_SIZE))
		if n < MAX/RANGES {
			distributionResult[0]++
		} else if n >= MAX/RANGES && n <= MAX/2 {
			distributionResult[1]++
		} else if n > MAX/2 && n < MAX-(MAX/RANGES) {
			distributionResult[2]++
		} else {
			distributionResult[3]++
		}
	}
	var previous uint32 = distributionResult[0]
	for _, val := range distributionResult[1:] {
		if val > (previous+1000) || val < (previous-1000) {
			t.Error("Distribution outside the allowed bounds")
		}
	}
	t.Log(distributionResult)
}

func TestRangeDistribution_fnv1aHash(t *testing.T) {
	distributionResult := make([]uint32, RANGES)
	for i := 0; i < SAMPLE_SIZE; i++ {
		n := fnv1aHash(randSeq(SEQUENCE_SIZE))
		if n < MAX/RANGES {
			distributionResult[0]++
		} else if n >= MAX/RANGES && n <= MAX/2 {
			distributionResult[1]++
		} else if n > MAX/2 && n < MAX-(MAX/RANGES) {
			distributionResult[2]++
		} else {
			distributionResult[3]++
		}
	}

	var previous uint32 = distributionResult[0]
	for _, val := range distributionResult[1:] {
		if val > (previous+1000) || val < (previous-1000) {
			t.Error("Distribution outside the allowed bounds")
		}
	}

	t.Log(distributionResult)
}

func TestRangeDistribution_fnv1Hash(t *testing.T) {
	t.Skip()
	distributionResult := make([]uint32, RANGES)
	for i := 0; i < SAMPLE_SIZE; i++ {
		n := fnv1Hash(randSeq(SEQUENCE_SIZE))
		if n < MAX/RANGES {
			distributionResult[0]++
		} else if n >= MAX/RANGES && n <= MAX/2 {
			distributionResult[1]++
		} else if n > MAX/2 && n < MAX-(MAX/RANGES) {
			distributionResult[2]++
		} else {
			distributionResult[3]++
		}
	}

	var previous uint32 = distributionResult[0]
	for _, val := range distributionResult[1:] {
		if val > (previous+1000) || val < (previous-1000) {
			t.Error("Distribution outside the allowed bounds")
		}
	}
	t.Log(distributionResult)
}

func Benchmark_crc16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		crc16(randSeq(52))
	}
}

func Benchmark_customHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		customHash(randSeq(52))
	}
}

func Benchmark_fnv1ahash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fnv1aHash(randSeq(52))
	}
}

func Benchmark_fnv1hash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fnv1Hash(randSeq(52))
	}
}
