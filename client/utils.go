package main

import "regexp"

// A simple regex to distinguish UI event strings from typed letters.
// I.e. "<Enter>" is illegal, and "y" is legal
func IsLegalCharacter(c string) bool {
	matched, err := regexp.Match(`^<[\w-]+>`, []byte(c))
	return err == nil && !matched
}
