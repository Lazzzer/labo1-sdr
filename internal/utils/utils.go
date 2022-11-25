// Auteurs: Jonathan Friedli, Lazar Pavicevic
// Labo 2 SDR

package utils

import (
	"net"
	"strconv"
	"strings"
)

// Max retourne le maximum entre deux entiers
func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// MapKeysToArray retourne les clés d'une map sous forme de tableau
func MapKeysToArray[T net.Conn](m map[int]T) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// IntToString retourne un affichage spécifique identifiant un serveur avec son numéro
func IntToString(numArray []int) string {
	numArrayStr := make([]string, len(numArray))
	for i, v := range numArray {
		numArrayStr[i] = "S" + strconv.Itoa(v)
	}
	return strings.Trim(strings.Join(numArrayStr, ","), ",")
}
