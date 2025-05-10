package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{"name with number", "3. Some name", "Some name"},
		{"name without number", "Some name", "Some name"},
		{"empty name", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFileName(tt.fileName)
			assert.Equal(t, tt.expected, result, "not expected name value")
		})
	}
}
