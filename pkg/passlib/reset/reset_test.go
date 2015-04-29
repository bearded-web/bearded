package reset

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenVerification(t *testing.T) {
	pwdHash := []byte("passwordHash")
	email := "my@email.com"
	secret := []byte("1234567891023123123")
	token := NewToken(email, time.Second*10, pwdHash, secret)
	assert.True(t, len(token) > decodedMinLength)
	assert.True(t, len(token) < decodedMaxLength)
	emailA, err := VerifyToken(token, func(email string) ([]byte, error) { return pwdHash, nil }, secret)
	require.NoError(t, err)
	assert.Equal(t, emailA, email)
}
