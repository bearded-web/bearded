// Helper for golang templates
// Helps to load templates from different directories and assets
// Based on https://github.com/unrolled/render
package template

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Renderer interface {
	Render(wr io.Writer, name string, binding interface{}, opts ...RenderOptions) error
}

// Included helper functions for use when rendering HTML.
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
}

// Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}

type Opts struct {
	// Directory to load templates. Default is "extra/templates".
	Directory string
	// Asset function to use in place of directory. Defaults to nil.
	Asset func(name string) ([]byte, error)
	// AssetNames function to use in place of directory. Defaults to nil.
	AssetNames func() []string
	// Layout template name. Will not render a layout if blank (""). Defaults to blank ("").
	Layout string
	// Extensions to parse template files from. Defaults to [".html"].
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	Funcs []template.FuncMap
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims Delims

	// reload templates automatically
	ReloadTemplates bool
}

// RenderOptions is a struct for overriding some rendering Options for specific Render call.
type RenderOptions struct {
	// Layout template name. Overrides Options.Layout.
	Layout string
}

type Template struct {
	opt       *Opts
	templates *template.Template
}

func New(opt *Opts) *Template {
	opt = setToDefault(opt)
	tmpl := &Template{
		opt: opt,
	}
	tmpl.compileTemplates()
	return tmpl
}

func (t *Template) compileTemplates() {
	if t.opt.Asset == nil || t.opt.AssetNames == nil {
		t.compileTemplatesFromDir()
		return
	}
	t.compileTemplatesFromAsset()
}

func (t *Template) compileTemplatesFromDir() {
	dir := t.opt.Directory
	t.templates = template.New(dir)
	t.templates.Delims(t.opt.Delims.Left, t.opt.Delims.Right)

	// Walk the supplied directory and compile any files that match our extension list.
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = "." + strings.Join(strings.Split(rel, ".")[1:], ".")
		}

		for _, extension := range t.opt.Extensions {
			if ext == extension {

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				name := (rel[0 : len(rel)-len(ext)])
				tmpl := t.templates.New(filepath.ToSlash(name))

				// Add our funcmaps.
				for _, funcs := range t.opt.Funcs {
					tmpl.Funcs(funcs)
				}

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}

		return nil
	})
}

func (t *Template) compileTemplatesFromAsset() {
	dir := t.opt.Directory
	t.templates = template.New(dir)
	t.templates.Delims(t.opt.Delims.Left, t.opt.Delims.Right)

	for _, path := range t.opt.AssetNames() {
		if !strings.HasPrefix(path, dir) {
			continue
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			panic(err)
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = "." + strings.Join(strings.Split(rel, ".")[1:], ".")
		}

		for _, extension := range t.opt.Extensions {
			if ext == extension {

				buf, err := t.opt.Asset(path)
				if err != nil {
					panic(err)
				}

				name := (rel[0 : len(rel)-len(ext)])
				tmpl := t.templates.New(filepath.ToSlash(name))

				// Add our funcmaps.
				for _, funcs := range t.opt.Funcs {
					tmpl.Funcs(funcs)
				}

				// Break out if this parsing fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}
	}
}

func (t *Template) Render(wr io.Writer, name string, binding interface{}, opts ...RenderOptions) error {
	// If we are in development mode, recompile the templates on every HTML request.
	if t.opt.ReloadTemplates {
		t.compileTemplates()
	}

	opt := t.prepareRenderOptions(opts)

	// Assign a layout if there is one.
	if len(opt.Layout) > 0 {
		t.addYield(name, binding)
		name = opt.Layout
	}
	return t.templates.ExecuteTemplate(wr, name, binding)
}

func (t *Template) execute(name string, binding interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	return buf, t.templates.ExecuteTemplate(buf, name, binding)
}

func (t *Template) addYield(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf, err := t.execute(name, binding)
			// Return safe HTML here since we are rendering our own template.
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
	}
	t.templates.Funcs(funcs)
}

func (t *Template) prepareRenderOptions(htmlOpt []RenderOptions) RenderOptions {
	if len(htmlOpt) > 0 {
		return htmlOpt[0]
	}

	return RenderOptions{
		Layout: t.opt.Layout,
	}
}

func setToDefault(opt *Opts) *Opts {
	if opt == nil {
		opt = &Opts{}
	}
	if len(opt.Directory) == 0 {
		opt.Directory = "extra/templates"
	}
	if len(opt.Extensions) == 0 {
		opt.Extensions = []string{".html"}
	}
	return opt
}
