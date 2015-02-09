package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {

	// get data
	m, err := NewMessage(CmdRequest, "data")
	require.NoError(t, err)
	require.Equal(t, CmdRequest, m.Cmd)
	recvStr := ""
	err = m.GetData(&recvStr)
	require.NoError(t, err)

	assert.Equal(t, "data", recvStr)

	err = m.SetData("data2")
	require.NoError(t, err)
	m.GetData(&recvStr)
	assert.Equal(t, "data2", recvStr)

	recvInt := 1
	err = m.GetData(&recvInt)
	require.Error(t, err)

	m.Data = nil
	err = m.GetData(recvInt)
	assert.NoError(t, err)

	err = m.SetData(make(chan struct{}))
	require.Error(t, err)

}
