package template

import (
	"bytes"
	"testing"

	"fmt"
	"html/template"

	"github.com/stretchr/testify/assert"
)

func TestRenderBadHTML(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "nope", nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "html/template: \"nope\" is undefined")
}

func TestRender(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "hello", "gophers")
	assert.NoError(t, err)
	assert.Equal(t, "<h1>Hello gophers</h1>\n", buf.String())
}

func TestRenderFuncs(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/custom_funcs",
		Extensions: []string{".tmpl"},
		Funcs: []template.FuncMap{
			{
				"myCustomFunc": func() string {
					return "My custom function"
				},
			},
		},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "index", "gophers")
	assert.NoError(t, err)
	assert.Equal(t, "My custom function\n", buf.String())
}

func TestRenderLayout(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
		Layout:     "layout",
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "content", "gophers")
	assert.NoError(t, err)
	assert.Equal(t, "head\n<h1>gophers</h1>\n\nfoot\n", buf.String())
}

func TestRenderLayoutCurrent(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
		Layout:     "current_layout",
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "content", "gophers")
	assert.NoError(t, err)
	assert.Equal(t, "content head\n<h1>gophers</h1>\n\ncontent foot\n", buf.String())
}

func TestRenderNested(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "admin/index", "gophers")
	assert.NoError(t, err)
	assert.Equal(t, "<h1>Admin gophers</h1>\n", buf.String())
}

func TestRenderBadPath(t *testing.T) {
	render := New(&Opts{
		Directory:  "../../../../../../../../../../../../../../../../testdata/basic",
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "hello", "gophers")
	assert.Error(t, err)
	assert.EqualError(t, err, "html/template: \"hello\" is undefined")
}

func TestRenderDelimeters(t *testing.T) {
	render := New(&Opts{
		Delims:     Delims{"{[{", "}]}"},
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "delims", "gophers")
	assert.NoError(t, err)
	assert.Equal(t, "<h1>Hello gophers</h1>", buf.String())
}

func TestRenderOverrideLayout(t *testing.T) {
	render := New(&Opts{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "content", "gophers",
		RenderOptions{Layout: "another_layout"})
	assert.NoError(t, err)
	assert.Equal(t, "another head\n<h1>gophers</h1>\n\nanother foot\n", buf.String())
}

func TestRenderFromAsset(t *testing.T) {
	render := New(&Opts{
		Directory: "testdata/asset",
		Asset: func(file string) ([]byte, error) {
			switch file {
			case "testdata/asset/test.tmpl":
				return []byte("<h1>gophers</h1>\n"), nil
			case "testdata/asset/layout.tmpl":
				return []byte("head\n{{ yield }}\nfoot\n"), nil
			default:
				return nil, fmt.Errorf("file not found: %s", file)
			}
		},
		AssetNames: func() []string {
			return []string{"testdata/asset/test.tmpl", "testdata/asset/layout.tmpl"}
		},
		Extensions: []string{".tmpl"},
	})

	buf := bytes.NewBuffer(nil)
	err := render.Render(buf, "test", "gophers",
		RenderOptions{Layout: "layout"})
	assert.NoError(t, err)
	assert.Equal(t, "head\n<h1>gophers</h1>\n\nfoot\n", buf.String())
}
