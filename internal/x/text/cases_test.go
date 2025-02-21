package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCaser(t *testing.T) {
	capitalizations := []string{"ID", "URL"}
	resolveExtensions := []string{".go", ".txt"}

	caser := NewCaser(capitalizations, resolveExtensions)

	assert.NotNil(t, caser)
	assert.Equal(t, capitalizations, caser.capitalizations)
	assert.Equal(t, resolveExtensions, caser.resolveExtensions)
}

func TestIdentifierFromFileName(t *testing.T) {
	caser := NewCaser(nil, []string{".go", ".txt"})

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
	caser := NewCaser(nil, nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"", "Blank"},
		{"*", "Wildcard"},
		{"example", "Example"},
		{"example_test", "ExampleTest"},
		{"exampleTest", "ExampleTest"},
		{"123example", "A123Example"},
	}

	for _, test := range tests {
		result := caser.Identifierize(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestCapitalize(t *testing.T) {
	caser := NewCaser([]string{"ID", "URL"}, nil)

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

func TestSplitIdentifierByCaseAndSeparators(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"exampleTest", []string{"example", "Test"}},
		{"example_test", []string{"example", "test"}},
		{"ExampleTest", []string{"Example", "Test"}},
		{"example123Test", []string{"example", "123", "Test"}},
	}

	for _, test := range tests {
		result := splitIdentifierByCaseAndSeparators(test.input)
		assert.Equal(t, test.expected, result)
	}
}
