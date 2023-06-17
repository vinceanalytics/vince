package names

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

// Valid returns true if s is a valid username. Username starts with
// alphanumeric and contains only alphanumeric and hyphen.
//
// Must be a valid unicode text.
func Valid(s string) bool {
	// At least 4 characters
	if len(s) < 4 {
		return false
	}
	r := strings.NewReader(s)
	start, _, err := r.ReadRune()
	if err != nil {
		return false
	}
	switch {
	case unicode.IsLetter(start), unicode.IsNumber(start):
	default:
		return false
	}
	for {
		o, _, err := r.ReadRune()
		if err != nil {
			return errors.Is(err, io.EOF)
		}
		switch {
		case unicode.IsLetter(o),
			unicode.IsNumber(o),
			o == '-':
		default:
			return false
		}
	}
}
