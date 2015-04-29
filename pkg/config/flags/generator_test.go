package flags

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateFlags(t *testing.T) {
	type Sub struct {
		Name  string
		Name2 string `env:"NAME_TWO"`
		Name3 string `env:"~NAME_THREE"`
	}
	type Cfg1 struct {
		Name    string `desc:"description" desc2:"description2"`
		IntVal  int    `env:"INTEGER_VAL"`
		BoolVal bool   `env:"-"`
		Sub     Sub
	}
	flags := GenerateFlags(&Cfg1{Name: "default name"})
	require.Len(t, flags, 6)
	assert.Equal(t, "--name \"default name\"\tdescription [$NAME]", flags[0].String())
	assert.Equal(t, "--int-val \"0\"\t [$INTEGER_VAL]", flags[1].String())
	assert.Equal(t, "--bool-val\t", flags[2].String())
	assert.Equal(t, "--sub-name \t [$SUB_NAME]", flags[3].String())
	assert.Equal(t, "--sub-name2 \t [$SUB_NAME_TWO]", flags[4].String())
	assert.Equal(t, "--sub-name3 \t [$NAME_THREE]", flags[5].String())

	flags = GenerateFlags(&Cfg1{Name: "default name"},
		Opts{DescTag: "desc2", Prefix: "api", EnvPrefix: "BEARDED"})
	require.Len(t, flags, 6)
	assert.Equal(t, "--api-name \"default name\"\tdescription2 [$BEARDED_API_NAME]", flags[0].String())
	assert.Equal(t, "--api-sub-name \t [$BEARDED_API_SUB_NAME]", flags[3].String())
	assert.Equal(t, "--api-sub-name2 \t [$BEARDED_API_SUB_NAME_TWO]", flags[4].String())
	assert.Equal(t, "--api-sub-name3 \t [$NAME_THREE]", flags[5].String())

}
