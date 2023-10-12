package utils

import "strings"

func IsAffirmative(input string) bool {
	lowercaseInput := strings.ToLower(input)
	affirmativeSet := make(map[string]struct{})
	affirmativeSet["y"] = struct{}{}
	affirmativeSet["yes"] = struct{}{}
	_, isAffirmative := affirmativeSet[lowercaseInput]
	return isAffirmative
}
