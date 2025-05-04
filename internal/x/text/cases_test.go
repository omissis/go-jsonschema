package text_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atombender/go-jsonschema/internal/x/text"
)

func TestNewCaser(t *testing.T) {
	t.Parallel()

	capitalizations := []string{"ID", "URL"}
	resolveExtensions := []string{".go", ".txt"}

	caser := text.NewCaser(capitalizations, resolveExtensions)

	assert.NotNil(t, caser)
}

func TestIdentifierFromFileName(t *testing.T) {
	t.Parallel()

	caser := text.NewCaser(nil, []string{".go", ".txt"})

	tests := []struct {
		fileName string
		expected string
	}{
		{"example.go", "Example"},
		{"example.txt", "Example"},
		{"example.md", "ExampleMd"},
		{"example", "Example"},
	}

	for _, test := range tests {
		result := caser.IdentifierFromFileName(test.fileName)
		assert.Equal(t, test.expected, result)
	}
}

func TestIdentifierize(t *testing.T) {
	t.Parallel()

	caser := text.NewCaser(nil, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"", "Blank"},
		{"*", "Wildcard"},
		{"example", "Example"},
		{"example_test", "ExampleTest"},
		{"exampleTest", "ExampleTest"},
		{"ExampleTest", "ExampleTest"},
		{"123example", "A123Example"},
		{"example123Test", "Example123Test"},
	}

	for _, test := range tests {
		result := caser.Identifierize(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestCapitalize(t *testing.T) {
	t.Parallel()

	caser := text.NewCaser([]string{"ID", "URL"}, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"id", "ID"},
		{"url", "URL"},
		{"example", "Example"},
	}

	for _, test := range tests {
		result := caser.Capitalize(test.input)
		assert.Equal(t, test.expected, result)
	}
}
