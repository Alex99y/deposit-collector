package utils

import "strings"

func ContainsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && (len(s) >= len(sub)) && (strings.Contains(s, sub)) {
			return true
		}
	}
	return false
}
