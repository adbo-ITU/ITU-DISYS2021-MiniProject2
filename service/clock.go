package service

import (
	"fmt"
	"sort"
	"strings"
)

type VectorClock map[string]uint32

func MergeClocks(a VectorClock, b VectorClock) VectorClock {
	newClock := make(VectorClock)
	for ak, av := range a {
		newClock[ak] = av
	}

	for bk, bv := range b {
		// This is clever merging:
		// if newClock[bk] doesn't exist, then the default value `0` is returned
		// and we can get the maximum value of `bv` or `0`, saving precious
		// SLOCs on silly `if` statements and such
		newClock[bk] = max(bv, newClock[bk])
	}

	return newClock
}

func max(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func FormatVectorClockAsString(clock VectorClock) string {
	parts := make([]string, 0)
	for k, v := range clock {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	sort.Strings(parts)
	return fmt.Sprintf("<%s>", strings.Join(parts, ", "))
}
