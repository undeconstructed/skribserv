package lib

import "math/rand"

const symbolsForIDs = "abcdefghijklmnopqrstuvwxyz0123456789"

func MakeRandomID(prefix string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = symbolsForIDs[rand.Int63()%int64(len(symbolsForIDs))]
	}

	return prefix + "-" + string(b)
}
