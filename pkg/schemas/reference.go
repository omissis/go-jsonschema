package schemas

import (
	"errors"
	"fmt"
	"net/url"
)

var (
	ErrEmptyReference       = errors.New("reference is empty")
	ErrUnsupportedRefFormat = errors.New("unsupported $ref format")
	ErrCannotParseRef       = errors.New("cannot parse $ref")
	ErrUnsupportedRefSchema = errors.New("unsupported $ref schema")
	ErrGetRefType           = errors.New("cannot get $ref type")
)

type RefType string

const (
	RefTypeFile    RefType = "file"
	RefTypeHTTP    RefType = "http"
	RefTypeHTTPS   RefType = "https"
	RefTypeUnknown RefType = "unknown"
)

func GetRefType(ref string) (RefType, error) {
	urlRef, err := url.Parse(ref)
	if err != nil {
		return RefTypeUnknown, fmt.Errorf("%w: %w", ErrGetRefType, err)
	}

	switch urlRef.Scheme {
	case "http":
		return RefTypeHTTP, nil

	case "https":
		return RefTypeHTTPS, nil

	case "file", "":
		return RefTypeFile, nil

	default:
		return RefTypeUnknown, fmt.Errorf("%w: %w", ErrGetRefType, ErrUnsupportedRefSchema)
	}
}
