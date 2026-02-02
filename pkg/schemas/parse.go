package schemas

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/atombender/go-jsonschema/pkg/yamlutils"
)

func FromJSONFile(fileName string) (*Schema, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	return FromJSONReader(f)
}

func FromJSONReader(r io.Reader) (*Schema, error) {
	var schema Schema
	if err := json.NewDecoder(r).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &schema, nil
}

func FromYAMLFile(fileName string) (*Schema, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	return FromYAMLReader(f)
}

func FromYAMLReader(r io.Reader) (*Schema, error) {
	// Marshal to JSON first because YAML decoder doesn't understand JSON tags.
	var m map[string]any

	if err := yaml.NewDecoder(r).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	yamlutils.FixMapKeys(m)

	value, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	var schema Schema

	if err = json.Unmarshal(value, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &schema, nil
}
