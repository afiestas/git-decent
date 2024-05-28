package internal

import "math/rand"

var testRandom bool = false

func randomMinute() int {
	if testRandom {
		return 5
	}

	return rand.Intn(10)
}
