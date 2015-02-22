package manager

import (
	"testing"
	//	c "github.com/smartystreets/goconvey/convey"
	"bytes"
	"github.com/bearded-web/bearded/models/file"
	"github.com/bearded-web/bearded/pkg/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileManager(t *testing.T) {
	mongo, dbName, err := tests.RandomTestMongoUp()
	if err != nil {
		t.Fatal(err)
	}
	defer tests.RandomTestMongoDown(mongo, dbName)
	println("Up test mongodb", dbName)

	mgr := New(mongo.DB(dbName))

	data := []byte("data")

	meta, err := mgr.Files.Create(bytes.NewBuffer(data), &file.Meta{Name: "file1.txt", ContentType: "text/plain"})

	require.NoError(t, err)
	require.NotNil(t, meta)
	assert.Equal(t, len(data), meta.Size)
	assert.Equal(t, "file1.txt", meta.Name)
	assert.Equal(t, "text/plain", meta.ContentType)

	f, err := mgr.Files.GetById(meta.Id)
	require.NoError(t, err)
	defer f.Close()
	require.NotNil(t, f)
	require.Equal(t, meta, f.Meta)

	_, err = mgr.Files.GetById("bad id")
	require.Error(t, err)
}
