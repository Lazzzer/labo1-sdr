package utils

import "net"

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func MapKeysToArray[T net.Conn](m map[int]T) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
