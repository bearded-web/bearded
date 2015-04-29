package flags

import (
	"bytes"
	"strings"
	"unicode"
)

const (
	flagDivider = "-"
	envDivider  = "_"
)

// transform s from camel-case to CamelCase
func FlagToCamel(s string) string {
	var result string
	words := strings.Split(s, flagDivider)
	for _, word := range words {
		w := []rune(word)
		w[0] = unicode.ToUpper(w[0])
		result += string(w)
	}
	return result
}

// transform s from CamelCase to camel-case
func CamelToFlag(s string) string {
	rs := []rune(s)
	buf := bytes.NewBuffer(make([]byte, 0, len(rs)+5))
	for i := 0; i < len(rs); i++ {
		if unicode.IsUpper(rs[i]) && i > 0 {
			buf.WriteString(flagDivider)
		}
		buf.WriteRune(unicode.ToLower(rs[i]))
	}
	return strings.ToLower(buf.String())
}

// transform s from camel-case to CAMEL_CASE
func FlagToEnv(s string) string {
	return strings.ToUpper(strings.Join(strings.Split(s, flagDivider), envDivider))
}
