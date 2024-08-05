package types

import (
	"fmt"
	"time"

	"github.com/sosodev/duration"
)

// TODO: Can we just call this "Duration"?
type SerializableDuration struct {
	time.Duration
}

func (date SerializableDuration) MarshalJSON() ([]byte, error) {
	//TODO: Implement this later
	return []byte(""), nil
}

func (date *SerializableDuration) UnmarshalJSON(data []byte) error {
	stringifiedData := string(data)
	if stringifiedData == "null" {
		return nil
	}

	d, err := duration.Parse(stringifiedData)
	if err != nil {
		return fmt.Errorf("unable to parse duration from JSON: %w", err)
	}

	date.Duration = d.ToTimeDuration()

	return nil
}
