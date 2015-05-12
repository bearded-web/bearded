package load

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToml(t *testing.T) {
	testData := []struct {
		data     string
		expected *TestStruct
		hasError bool
	}{
		{"error format", nil, true},
		{`
		Field1 = "value1"
		FieldName2 = 10
		[sub]
		   FieldName3 = "value3"
		`, &TestStruct{"value1", 10, SubStruct{FieldName3: "value3"}}, false},
	}
	for _, td := range testData {
		actual := &TestStruct{}
		err := LoadToml(bytes.NewBufferString(td.data), actual)
		if td.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, td.expected, actual)
		}
	}
}
