package service

import (
	"fmt"
	"sort"
	"strings"
)

//Defines a VectorClock as a map from a username string to a timestamp integer
type VectorClock map[string]uint32

//Method that merges two VectorClocks according to the rules for Lamport timestamps
//For each clock in the map, the maximum timestamp is saved in the merged newClock
//If a clock only exists in one of the two incoming clocks, it is added to the merged
//clock with its existing value
func MergeClocks(a VectorClock, b VectorClock) VectorClock {
	newClock := make(VectorClock)
	for ak, av := range a {
		newClock[ak] = av
	}

	for bk, bv := range b {
		// We abues Go here, and we are loving it:
		// if newClock[bk] doesn't exist, then the default value `0` is returned
		// and we can get the maximum value of `bv` or `0`, saving precious
		// SLOCs on silly `if` statements and such. And yes, this comment takes
		// up more space than those if statements would have, but so what??
		newClock[bk] = max(bv, newClock[bk])
	}

	return newClock
}

//Utility function to compare to integers and return the maximum
func max(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

//Formats vector clocks for printing as a string with comma-seperated pairs of username: timestamp
func FormatVectorClockAsString(clock VectorClock) string {
	parts := make([]string, 0)
	for k, v := range clock {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	sort.Strings(parts)
	return fmt.Sprintf("<%s>", strings.Join(parts, ", "))
}
