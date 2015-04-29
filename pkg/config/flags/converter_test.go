package flags

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFlagToCamel(t *testing.T) {
	data := []struct {
		Src string
		Exp string
	}{
		{"value-value2-value3", "ValueValue2Value3"},
		{"value", "Value"},
	}
	for _, d := range data {
		assert.Equal(t, d.Exp, FlagToCamel(d.Src))
	}
}

func TestCamelToFlag(t *testing.T) {
	data := []struct {
		Src string
		Exp string
	}{
		{"ValueValue2Value3", "value-value2-value3"},
		{
			"ValueValue2Value3Value4Value5Value6Value7",
			"value-value2-value3-value4-value5-value6-value7",
		},
		{"Value", "value"},
	}
	for _, d := range data {
		assert.Equal(t, d.Exp, CamelToFlag(d.Src))
	}
}

func TestFlagToEnv(t *testing.T) {
	data := []struct {
		Src string
		Exp string
	}{
		{"value-value2-value3", "VALUE_VALUE2_VALUE3"},
		{"value", "VALUE"},
	}
	for _, d := range data {
		assert.Equal(t, d.Exp, FlagToEnv(d.Src))
	}
}
