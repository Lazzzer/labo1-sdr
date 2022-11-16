package utils

import (
	"net"
	"strconv"
	"strings"
)

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

func IntToString(numArray []int) string {
	numArrayStr := make([]string, len(numArray))
	for i, v := range numArray {
		numArrayStr[i] = "S" + strconv.Itoa(v)
	}

	return strings.Trim(strings.Join(numArrayStr, ","), ",")
}
