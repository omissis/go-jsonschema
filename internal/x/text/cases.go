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

	rIdent := []rune(ident)
	if len(rIdent) > 0 {
		if !unicode.IsLetter(rIdent[0]) || isNotCaseSensitiveLetter(rIdent[0]) {
			ident = "A" + ident
		}
	}

	if ident == "" {
		return "Undefined"
	}

	return ident
}

func isNotCaseSensitiveLetter(r rune) bool {
	return !unicode.IsUpper(r) && !unicode.IsLower(r)
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

	r := []rune(s)

	return string(unicode.ToUpper(r[0])) + string(r[1:])
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
		stateNoCase
		stateNumber
		stateDelimiter
	)

	var result [][]rune

	currState, j := stateNothing, 0

	runes := []rune(s)

	for i, r := range runes {
		var nextState state

		switch {
		case unicode.IsLower(r):
			nextState = stateLower

		case unicode.IsUpper(r):
			nextState = stateUpper

		case unicode.IsNumber(r):
			nextState = stateNumber

		case !unicode.IsLetter(r): // Non-letter characters.
			nextState = stateDelimiter

		default: // Non-case sensitive letters.
			nextState = stateNoCase
		}

		if nextState != currState {
			if currState == stateDelimiter {
				j = i
			} else if !(currState == stateUpper && nextState == stateLower) {
				if i > j {
					result = append(result, runes[j:i])
				}

				j = i
			}

			currState = nextState
		}
	}

	if currState != stateDelimiter && len(runes)-j > 0 {
		result = append(result, runes[j:])
	}

	return runesToStrings(result)
}

func runesToStrings(runes [][]rune) []string {
	result := make([]string, len(runes))
	for i, r := range runes {
		result[i] = string(r)
	}

	return result
}
