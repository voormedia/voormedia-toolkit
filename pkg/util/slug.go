package util

import (
	"regexp"
	"strings"
)

var regexpNonAuthorizedChars = regexp.MustCompile("[^a-z0-9-_]")
var regexpMultipleDashes = regexp.MustCompile("-+")

func Slugify(str string) string {
	str = strings.ToLower(strings.TrimSpace(str))
	str = regexpNonAuthorizedChars.ReplaceAllString(str, "-")
	str = regexpMultipleDashes.ReplaceAllString(str, "-")
	return strings.Trim(str, "-")
}
