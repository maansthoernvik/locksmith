package queue

import "hash/fnv"

const MAX_HASH uint16 = 65535

// Do not use, not finished
func customHash(str string) uint16 {
	var hash uint16 = 0

	for i := 0; i < len(str); i++ {
		hash = uint16(str[i]) << 8 & hash
	}

	return hash
}

func fnv1aHash(str string) uint16 {
	alg := fnv.New32a()
	alg.Write([]byte(str))
	return uint16(alg.Sum32() % 65535)
}

func fnv1Hash(str string) uint16 {
	alg := fnv.New32()
	alg.Write([]byte(str))
	return uint16(alg.Sum32() % 65535)
}
