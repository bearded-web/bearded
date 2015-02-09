//go:generate stringer -type=Cmd
package transport

import (
	"encoding/json"
	"sync"
)

type Cmd int

const (
	CmdRequest Cmd = iota
	CmdResponse
	CmdError
)

type Message struct {
	Id   int              `json:"id"`
	Cmd  Cmd              `json:"cmd"`
	Data *json.RawMessage `json:"data"`
}

func (m *Message) GetData(obj interface{}) error {
	if m.Data == nil {
		return nil
	}
	return json.Unmarshal(*m.Data, obj)
}

func (m *Message) SetData(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	b := json.RawMessage(data)
	m.Data = &b
	return nil
}

func (m *Message) Extract(obj interface{}) error {
	return m.GetData(obj)
}

func NewMessage(cmd Cmd, obj interface{}) (*Message, error) {
	m := &Message{
		Id:  getId(),
		Cmd: cmd,
	}
	m.SetData(obj)
	return m, nil
}

var idLock = sync.Mutex{}
var __messageId = 0

func getId() int {
	// TODO (m0sth8): replace with channel and buffer
	idLock.Lock()
	ret := __messageId
	__messageId += 1
	idLock.Unlock()
	return ret
}
