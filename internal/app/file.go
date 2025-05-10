package app

import "regexp"

var nameRegex = regexp.MustCompile(`^(?:\d\.\s)?(.+)$`)

func ExtractFileName(val string) string {
	matches := nameRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func ExtractNumber(val string) int64 {
	return 0
}
