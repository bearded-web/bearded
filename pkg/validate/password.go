package validate

import "fmt"

const (
	PasswordMinLength = 7
	PasswordMaxLength = 100
)

func Password(password string) (bool, string) {
	// TODO (m0sth8): check for weak password
	if len(password) < PasswordMinLength {
		return false, fmt.Sprintf("length must be more then %d symbols", PasswordMinLength)
	}
	if len(password) > PasswordMaxLength {
		return false, fmt.Sprintf("length must be less then %d symbols", PasswordMaxLength)
	}
	return true, ""
}
