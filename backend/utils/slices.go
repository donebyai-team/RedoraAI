package utils

import "strings"

// SlicesAny returns true if any element in the slice satisfies the predicate function.
func SlicesAll[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if !f(v) {
			return false
		}
	}
	return true
}

// SlicesAllEquals returns true if all elements in the slice are equal to each other.
func SlicesAllEquals[T comparable](s []T) bool {
	if len(s) == 0 {
		return true
	}

	return SlicesAll(s, func(v T) bool { return v == s[0] })
}

func Contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if strings.EqualFold(a, needle) {
			return true
		}
	}
	return false
}
