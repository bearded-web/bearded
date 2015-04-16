package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MultiLine []string

func (m *MultiLine) UnmarshalJSON(data []byte) error {
	multi := []string{}
	if err := json.Unmarshal(data, &multi); err == nil {
		*m = MultiLine(multi)
		return nil
	}
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		multi = append(multi, single)
		*m = MultiLine(multi)
		return nil
	}
	return fmt.Errorf("MultiLine must be string or string list")
}

func (m MultiLine) String() string {
	return strings.Join([]string(m), " ")
}
