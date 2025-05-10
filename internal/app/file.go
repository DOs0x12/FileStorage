package app

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var nameRegex = regexp.MustCompile(`^(?:\d+\.\s)?(.+)$`)

func ExtractFileName(val string) string {
	matches := nameRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

var numRegex = regexp.MustCompile(`^(\d+)\.\s`)
var ErrWrongFormat = errors.New("string is in wrong format")

func ExtractNumber(val string) (int64, error) {
	matches := numRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		num, err := strconv.ParseInt(matches[1], 0, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to convert extracted number %v to int64: %w", matches[1], err)
		}

		return num, nil
	}

	return 0, ErrWrongFormat
}
