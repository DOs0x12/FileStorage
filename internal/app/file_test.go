package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{"name with number", "3. Some name", "Some name"},
		{"name with big number", "30000. Some name", "Some name"},
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

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected int64
		wantErr  bool
		expErr   error
	}{
		{"name with number", "3. Some name", 3, false, nil},
		{"name with big number", "30000. Some name", 30000, false, nil},
		{"name without number", "Some name", 0, true, ErrWrongFormat},
		{"empty name", "", 0, true, ErrWrongFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractNumber(tt.fileName)
			if tt.wantErr {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expErr.Error(), "not expected error")

				return
			}

			assert.Equal(t, tt.expected, result, "not expected number")
			assert.NoError(t, err)
		})
	}
}
