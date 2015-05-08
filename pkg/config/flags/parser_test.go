package flags

import (
	"os"
	"testing"

	"github.com/m0sth8/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlag(t *testing.T) {
	type Sub struct {
		Name  string
		Name2 string `env:"NAME_TWO"`
	}
	type Cfg1 struct {
		Name      string `desc:"description" desc2:"description2"`
		Name2     string
		IntVal    int
		IntVal2   int  `env:"-"`
		BoolVal   bool `env:"-"`
		BoolVal2  bool
		BoolVal3  bool
		Sub       Sub
		StrSlice1 []string
		StrSlice2 []string
		StrSlice3 []string
		IntSlice1 []int
		IntSlice2 []int
		IntSlice3 []int
	}
	os.Clearenv()
	os.Setenv("NAME2", "10")
	os.Setenv("SUB_NAME_TWO", "sub name 2")
	os.Setenv("SUB_NAME_TWO", "sub name 2")
	os.Setenv("INT_VAL", "10")
	os.Setenv("INT_VAL2", "10")
	os.Setenv("BOOL_VAL2", "true")
	os.Setenv("STR_SLICE3", "value3,value33")
	os.Setenv("INT_SLICE3", "3,33")
	a := cli.NewApp()
	a.Flags = GenerateFlags(&Cfg1{})
	a.Action = func(ctx *cli.Context) {
		cfg := &Cfg1{
			StrSlice1: []string{"value1"},
			StrSlice2: []string{"value2"},
			IntSlice1: []int{1},
			IntSlice2: []int{2},
		}
		err := ParseFlags(cfg, ctx)
		require.NoError(t, err)
		assert.Equal(t, "bla", cfg.Name)
		assert.Equal(t, "sub", cfg.Sub.Name)
		assert.Equal(t, "sub name 2", cfg.Sub.Name2)
		assert.Equal(t, "10", cfg.Name2)
		assert.Equal(t, true, cfg.BoolVal)
		assert.Equal(t, true, cfg.BoolVal2)
		assert.Equal(t, 10, cfg.IntVal)
		assert.Equal(t, 0, cfg.IntVal2)
		assert.Equal(t, []string{"value1"}, cfg.StrSlice1)
		assert.Equal(t, []string{"value22"}, cfg.StrSlice2)
		assert.Equal(t, []string{"value3", "value33"}, cfg.StrSlice3)
		assert.Equal(t, []int{1}, cfg.IntSlice1)
		assert.Equal(t, []int{22}, cfg.IntSlice2)
		assert.Equal(t, []int{3, 33}, cfg.IntSlice3)
	}
	a.Run([]string{"run", "--name", "bla",
		"--sub-name", "sub", "--bool-val", "--str-slice2", "value22", "--int-slice2", "22"})
}
