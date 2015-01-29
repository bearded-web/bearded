package fltr

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetFilterQuery(t *testing.T) {
	f := struct {
		Field1 string `fltr:"field1"`
		Field2 string `fltr:"field2"`
		Field3 int
		Field4 int `fltr:",`
	}{
		Field1: "value1",
		Field3: 10,
		Field4: 20,
	}

	data := GetQuery(f)
	require.NotNil(t, data)
	require.Len(t, data, 3)
	spew.Dump(data)
	assert.Equal(t, "value1", data["field1"])
	assert.Equal(t, 10, data["Field3"])
}
