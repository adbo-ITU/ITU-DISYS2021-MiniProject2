package main

import "regexp"

func IsLegalCharacter(c string) bool {
	matched, err := regexp.Match(`^<\w+>`, []byte(c))
	return err == nil && !matched
}
