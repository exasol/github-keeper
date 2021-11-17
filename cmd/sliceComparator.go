package cmd

import "sort"

func stringSlicesEqualIgnoringOrder(sliceA, sliceB []string) bool {
	if len(sliceA) != len(sliceB) {
		return false
	}
	sort.Strings(sliceA)
	sort.Strings(sliceB)
	for index := range sliceA {
		if sliceA[index] != sliceB[index] {
			return false
		}
	}
	return true
}
