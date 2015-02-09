package transport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessions(t *testing.T) {
	session := NewSession()
	chOut := session.Add(10)
	require.NotNil(t, chOut)

	chIn := session.Pick(10)
	require.NotNil(t, chIn)
	require.NotEqual(t, chIn, chOut)

	msg := &Message{}

	chIn <- msg
	require.Equal(t, msg, <-chOut)

	chIn = session.Pick(10)
	require.Nil(t, chIn)
}
