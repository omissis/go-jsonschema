package tests_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/atombender/go-jsonschema/pkg/types"
)

func TestSerializableDateMarshalsToJSON(t *testing.T) {
	t.Parallel()

	date := types.SerializableDate{
		Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	output, err := json.Marshal(date)
	if err != nil {
		t.Fatalf("Unable to marshal SerializableDate as JSON: %v", err)
	}

	stringifiedOutput := string(output)

	expected := "\"2023-01-02\""
	if stringifiedOutput != expected {
		t.Fatalf("Expected SerializableDate to marshal to %s but got %s", expected, stringifiedOutput)
	}
}

func TestSerializableDateUnmarshalsFromJSON(t *testing.T) {
	t.Parallel()

	input := "\"2023-01-02\""

	var date types.SerializableDate
	if err := json.Unmarshal([]byte(input), &date); err != nil {
		t.Fatalf("Unable to unmarshal %s to a SerializableDate: %s", input, err)
	}

	expected := types.SerializableDate{
		Time: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	if date != expected {
		t.Fatalf("Expected SerializableDate to unmarshal to %s but got %s", expected, date)
	}
}

func TestSerializableDateUnmarshalJSONReturnsErrorForInvalidString(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"",
		"2023-01-02",
		"\"2023\"",
		"\"2023-01\"",
		"\"2023-01-02T03:04:05Z\"",
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(fmt.Sprintf("Given '%s' expected UnmarshalJSON to return an error", testCase), func(t *testing.T) {
			t.Parallel()

			var date types.SerializableDate
			if err := date.UnmarshalJSON([]byte(testCase)); err == nil {
				t.Fatalf("Expected an error but got '%s'", date)
			}
		})
	}
}

func TestSerializableDateUnmarshalJSONDoesNothingWhenGivenNull(t *testing.T) {
	t.Parallel()

	var date types.SerializableDate
	if err := date.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("Given 'null' expected UnmarshalJSON to be no-op but got an error: %v", err)
	}

	var zeroValue types.SerializableDate
	if date != zeroValue {
		t.Fatalf("Given 'null' expected to stay at zero value but got %s", date)
	}
}
