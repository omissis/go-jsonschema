package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTimeNotJSONString = errors.New("cannot parse non-string value as a time")

//nolint:recvcheck // json marshal/unmarshal require value and pointer receivers
type SerializableTime struct {
	time.Time
}

func (t SerializableTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + t.Format(time.TimeOnly) + "\""), nil
}

func (t *SerializableTime) UnmarshalJSON(data []byte) error {
	stringifiedData := string(data)
	if stringifiedData == "null" {
		return nil
	}

	if !strings.HasPrefix(stringifiedData, "\"") || !strings.HasSuffix(stringifiedData, "\"") {
		return ErrTimeNotJSONString
	}

	timeWithoutQuotes := stringifiedData[1 : len(stringifiedData)-1]

	// RFC 3339 full-time layouts.
	layouts := []string{
		"15:04:05",       // Time without offset.
		"15:04:05Z07:00", // Time with offset. Works with ..Z, ..+09:00 ...
	}

	var parsedTime time.Time

	var err error

	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, timeWithoutQuotes)
		if err == nil {
			t.Time = parsedTime

			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("unable to parse time from JSON: %w", err)
	}

	return nil
}
