package tests_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/atombender/go-jsonschema/types"
)

func TestSerializableTimeMarshalsToJSON(t *testing.T) {
	t.Parallel()

	now := time.Now()
	timeValue := types.SerializableTime{
		Time: now,
	}

	output, err := json.Marshal(timeValue)
	if err != nil {
		t.Fatalf("Unable to marshal SerializableTime as JSON: %v", err)
	}

	stringifiedOutput := string(output)

	expected := "\"" + now.Format("15:04:05") + "\""
	if stringifiedOutput != expected {
		t.Fatalf("Expected SerializableTime to marshal to %s but got %s", expected, stringifiedOutput)
	}
}

func TestSerializableTimeUnmarshalsFromJSON(t *testing.T) {
	t.Parallel()

	now := time.Date(0, 1, 1, 15, 0o4, 0o5, 0, time.UTC)

	input := "\"" + now.Format("15:04:05") + "\""

	var timeValue types.SerializableTime
	if err := json.Unmarshal([]byte(input), &timeValue); err != nil {
		t.Fatalf("Unable to unmarshal %s to a SerializableTime: %s", input, err)
	}

	expected := types.SerializableTime{
		Time: now,
	}

	if timeValue != expected {
		t.Fatalf("Expected SerializableTime to unmarshal to %s but got %s", expected, timeValue)
	}
}

func TestSerializableTimeUnmarshalJSONReturnsErrorForInvalidString(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"",
		"15:04:05",
		"\"15\"",
		"\"15:04\"",
		"\"2023-01-02T03:04:05Z\"",
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(fmt.Sprintf("Given '%s' expected UnmarshalJSON to return an error", testCase), func(t *testing.T) {
			t.Parallel()

			var timeValue types.SerializableTime
			if err := timeValue.UnmarshalJSON([]byte(testCase)); err == nil {
				t.Fatalf("Expected an error but got '%s'", timeValue)
			}
		})
	}
}

func TestSerializableTimeUnmarshalJSONDoesNothingWhenGivenNull(t *testing.T) {
	t.Parallel()

	var timeValue types.SerializableTime
	if err := timeValue.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("Given 'null' expected UnmarshalJSON to be no-op but got an error: %v", err)
	}

	var zeroValue types.SerializableTime
	if timeValue != zeroValue {
		t.Fatalf("Given 'null' expected to stay at zero value but got %s", timeValue)
	}
}
