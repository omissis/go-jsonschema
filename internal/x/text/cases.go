package text

import (
	"path/filepath"
	"strings"
	"unicode"
)

func NewCaser(capitalizations, resolveExtensions []string) *Caser {
	return &Caser{
		capitalizations:   capitalizations,
		resolveExtensions: resolveExtensions,
	}
}

type Caser struct {
	capitalizations   []string
	resolveExtensions []string
}

func (c *Caser) IdentifierFromFileName(fileName string) string {
	s := filepath.Base(fileName)
	for _, ext := range c.resolveExtensions {
		trimmed := strings.TrimSuffix(s, ext)
		if trimmed != s {
			s = trimmed

			break
		}
	}

	return c.Identifierize(s)
}

func (c *Caser) Identifierize(s string) string {
	if s == "" {
		return "Blank"
	}

	// FIXME: Better handling of non-identifier chars.
	var sb strings.Builder
	for _, part := range splitIdentifierByCaseAndSeparators(s) {
		_, _ = sb.WriteString(c.Capitalize(part))
	}

	ident := sb.String()

	if !unicode.IsLetter(rune(ident[0])) {
		ident = "A" + ident
	}

	return ident
}

func (c *Caser) Capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}

	for _, c := range c.capitalizations {
		if strings.EqualFold(c, s) {
			return c
		}
	}

	return strings.ToUpper(s[0:1]) + s[1:]
}

func splitIdentifierByCaseAndSeparators(s string) []string {
	if len(s) == 0 {
		return nil
	}

	type state int

	const (
		stateNothing state = iota
		stateLower
		stateUpper
		stateNumber
		stateDelimiter
	)

	var result []string

	currState, j := stateNothing, 0

	for i := 0; i < len(s); i++ {
		var nextState state

		c := rune(s[i])

		switch {
		case unicode.IsLower(c):
			nextState = stateLower

		case unicode.IsUpper(c):
			nextState = stateUpper

		case unicode.IsNumber(c):
			nextState = stateNumber

		default:
			nextState = stateDelimiter
		}

		if nextState != currState {
			if currState == stateDelimiter {
				j = i
			} else if !(currState == stateUpper && nextState == stateLower) {
				if i > j {
					result = append(result, s[j:i])
				}
				j = i
			}

			currState = nextState
		}
	}

	if currState != stateDelimiter && len(s)-j > 0 {
		result = append(result, s[j:])
	}

	return result
}
