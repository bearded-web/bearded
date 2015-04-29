package flags

import (
	"github.com/m0sth8/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestParseFlag(t *testing.T) {
	type Sub struct {
		Name  string
		Name2 string `env:"NAME_TWO"`
	}
	type Cfg1 struct {
		Name     string `desc:"description" desc2:"description2"`
		Name2    string
		IntVal   int
		IntVal2  int  `env:"-"`
		BoolVal  bool `env:"-"`
		BoolVal2 bool
		BoolVal3 bool
		Sub      Sub
	}
	os.Clearenv()
	os.Setenv("NAME2", "10")
	os.Setenv("SUB_NAME_TWO", "sub name 2")
	os.Setenv("SUB_NAME_TWO", "sub name 2")
	os.Setenv("INT_VAL", "10")
	os.Setenv("INT_VAL2", "10")
	os.Setenv("BOOL_VAL2", "true")
	a := cli.NewApp()
	a.Flags = GenerateFlags(&Cfg1{})
	a.Action = func(ctx *cli.Context) {
		cfg := &Cfg1{}
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
	}
	a.Run([]string{"run", "--name", "bla",
		"--sub-name", "sub", "--bool-val", "true"})
}
