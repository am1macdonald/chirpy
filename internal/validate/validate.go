package validate

import (
	"errors"
	"fmt"
	"strings"
)

const (
	tweetMaxChar int = 140
)

func Validate(s string) (string, error) {
	fmt.Println("Validating chirp")
	if !checkLength(s) {
		return "", errors.New("Chirp is too long")
	}

	return filterProfanity(s), nil
}

func checkLength(s string) bool {
	return len(s) <= tweetMaxChar
}

func filterProfanity(s string) string {
	replacements := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}

	sa := strings.Fields(s)
	for i, word := range sa {
		redacted, ok := replacements[strings.ToLower(word)]
		if ok {
			sa[i] = redacted
		}
	}
	return strings.Join(sa, " ")
}
