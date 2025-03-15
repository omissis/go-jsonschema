package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrDateNotJSONString = errors.New("cannot parse non-string value as a date")

//nolint:recvcheck // json marshal/unmarshal require value and pointer receivers
type SerializableDate struct {
	time.Time
}

func (date SerializableDate) MarshalJSON() ([]byte, error) {
	return []byte("\"" + date.Format(time.DateOnly) + "\""), nil
}

func (date *SerializableDate) UnmarshalJSON(data []byte) error {
	stringifiedData := string(data)
	if stringifiedData == "null" {
		return nil
	}

	if !strings.HasPrefix(stringifiedData, "\"") || !strings.HasSuffix(stringifiedData, "\"") {
		return ErrDateNotJSONString
	}

	dataWithoutQuotes := stringifiedData[1 : len(stringifiedData)-1]

	parsedDate, err := time.Parse(time.DateOnly, dataWithoutQuotes)
	if err != nil {
		return fmt.Errorf("unable to parse date from JSON: %w", err)
	}

	date.Time = parsedDate

	return nil
}
