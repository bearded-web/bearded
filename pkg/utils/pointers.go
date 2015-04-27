package utils

func BoolP(v bool) *bool {
	return &v
}

func StringP(s string) *string {
	return &s
}
