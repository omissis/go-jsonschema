package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrTimeNotJSONString = errors.New("cannot parse non-string value as a time")

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

	parsedTime, err := time.Parse(time.TimeOnly, timeWithoutQuotes)
	if err != nil {
		return fmt.Errorf("unable to parse time from JSON: %w", err)
	}

	t.Time = parsedTime

	return nil
}
