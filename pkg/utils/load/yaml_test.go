package load

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYaml(t *testing.T) {
	testData := []struct {
		data     string
		expected *TestStruct
		hasError bool
	}{
		{"error format", nil, true},
		{"field1: value1\nfieldname2: 10", &TestStruct{"value1", 10}, false},
	}
	for _, td := range testData {
		actual := &TestStruct{}
		err := LoadYaml(bytes.NewBufferString(td.data), actual)
		if td.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, td.expected, actual)
		}
	}
}
