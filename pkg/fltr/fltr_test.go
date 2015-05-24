package fltr

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilterQuery(t *testing.T) {
	f := struct {
		Field1 string `fltr:"field1"`
		Field2 string `fltr:"field2"`
		Field3 int
		Field4 int `fltr:","`
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

func TestGetParams(t *testing.T) {
	ws := &restful.WebService{}
	f := struct {
		Field1 string `fltr:"field1,in"`
		Field2 string `fltr:"field222"`
		Field3 int
		Field4 *bool
	}{}
	params := GetParams(ws, f)
	require.NotNil(t, params)
	require.Len(t, params, 5)

	assert.Equal(t, "field1", params[0].Data().Name)
	assert.Equal(t, restful.QueryParameterKind, params[0].Data().Kind)
	assert.Equal(t, "string", params[0].Data().DataType)
	assert.Equal(t, false, params[0].Data().Required)

	assert.Equal(t, "field1_in", params[1].Data().Name)
	assert.Equal(t, restful.QueryParameterKind, params[1].Data().Kind)
	assert.Equal(t, "string", params[1].Data().DataType)
	assert.Equal(t, false, params[1].Data().Required)

	assert.Equal(t, "field222", params[2].Data().Name)

	assert.Equal(t, "Field3", params[3].Data().Name)
	assert.Equal(t, restful.QueryParameterKind, params[3].Data().Kind)
	assert.Equal(t, "integer", params[3].Data().DataType)
	assert.Equal(t, false, params[3].Data().Required)

	assert.Equal(t, "Field4", params[4].Data().Name)
	assert.Equal(t, restful.QueryParameterKind, params[4].Data().Kind)
	assert.Equal(t, "boolean", params[4].Data().DataType)
	assert.Equal(t, false, params[4].Data().Required)

}
