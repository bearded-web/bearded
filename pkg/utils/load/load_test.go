package load

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SubStruct struct {
	FieldName3 string
}

type TestStruct struct {
	Field1     string
	FieldName2 int
	Sub        SubStruct
}

func TestLoaderFromReader(t *testing.T) {
	l := New()
	assert.Len(t, l.Loaders, 0)
	assert.Len(t, l.ExtToFormat, 0)
	data := `{"field1": "value1","fieldname2": 10}`

	dst := &TestStruct{}
	err := l.FromReader(bytes.NewBufferString(data), dst, JsonFormat)
	assert.Error(t, err)
	assert.Equal(t, "Unknown format json", err.Error())

	l.Loaders[JsonFormat] = LoadJson
	err = l.FromReader(bytes.NewBufferString(data), dst, JsonFormat)
	assert.NoError(t, err)
	assert.Equal(t, &TestStruct{"value1", 10, SubStruct{}}, dst)
}

func TestLoaderFromFile(t *testing.T) {
	l := New()
	assert.Len(t, l.Loaders, 0)
	assert.Len(t, l.ExtToFormat, 0)
	data := `{"field1": "value1","fieldname2": 10}`

	// prepare file
	expected := &TestStruct{"value1", 10, SubStruct{}}
	dst := &TestStruct{}
	tmpDir, err := ioutil.TempDir("", "test_bearded")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	filename := filepath.Join(tmpDir, "file.json")
	err = ioutil.WriteFile(filename, []byte(data), 0660)
	if err != nil {
		t.Fatal(err)
	}

	err = l.FromFile(filename, dst)
	assert.Error(t, err)
	assert.Equal(t, "Cant't recognize format from extension .json", err.Error())

	l.Loaders[JsonFormat] = LoadJson
	l.ExtToFormat[".json"] = JsonFormat
	err = l.FromFile(filename, dst)
	assert.NoError(t, err)
	assert.Equal(t, expected, dst)

	err = l.FromFile(filename, dst, Opts{Format: TomlFormat})
	assert.Error(t, err) // format error
	assert.Equal(t, "Unknown format toml", err.Error())

	err = l.FromFile("/bad"+filename, dst)
	assert.Error(t, err)

	err = FromFile(filename, dst)
	assert.NoError(t, err)
	assert.Equal(t, expected, dst)

}

func TestDefaultLoader(t *testing.T) {
	testData := []struct {
		data     string
		format   Format
		expected *TestStruct
		hasError bool
	}{
		{"error format", JsonFormat, nil, true},
		{`{"field1": "value1","fieldname2": 10}`, JsonFormat, &TestStruct{"value1", 10, SubStruct{}}, false},

		{"error format", TomlFormat, nil, true},
		{`
		field1 = "value1"
		fieldname2 = 10
		`, TomlFormat, &TestStruct{"value1", 10, SubStruct{}}, false},

		{"error format", YamlFormat, nil, true},
		{"field1: value1\nfieldname2: 10", YamlFormat, &TestStruct{"value1", 10, SubStruct{}}, false},
	}
	for _, td := range testData {
		actual := &TestStruct{}
		err := FromReader(bytes.NewBufferString(td.data), actual, td.format)
		if td.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, td.expected, actual)
		}
	}
}
