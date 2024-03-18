package testing

import (
	"github.com/andyzhou/tinylib/algorithm"
	"testing"
)

var (
	hashRing *algorithm.HashRing
)

func init() {
	hashRing = algorithm.NewHashRingDefault()
}

func TestCHash(t *testing.T) {
	//add nods
	hashRing.Add("8", "6", "4", "2")

	//setup test cases
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"25": "6",
		"27": "8",
	}

	//test case
	for k, v := range testCases {
		hr := hashRing.Get(k)
		t.Logf("Part1. Asking for %s, should have %v, yielded %s\n", k, v, hr)
	}

	// Adds 8, 18, 28
	hashRing.Add("9")
	hashRing.Add("10")

	// 27 should now map to 8.
	testCases["28"] = "8"
	testCases["30"] = "10"
	testCases["31"] = "10"
	testCases["1263"] = "2"

	t.Logf("\n\n")
	for k, v := range testCases {
		hr := hashRing.Get(k)
		t.Logf("Part2. Asking for %s, should have %v, yielded %s\n", k, v, hr)
	}
}
