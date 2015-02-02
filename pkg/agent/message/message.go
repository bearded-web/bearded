//go:generate stringer -type=Cmd
package message

import (
	"encoding/json"
	"sync"
)

type Cmd int

const (
	CmdResponse Cmd = iota
	CmdError
	CmdGetPluginVersions
	CmdRunPlugin
)

type Message struct {
	Id   int              `json:"id"`
	Cmd  Cmd              `json:"cmd"`
	Data *json.RawMessage `json:"data"`
}

func (m *Message) GetData(obj interface{}) error {
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

func New(cmd Cmd, obj interface{}) (*Message, error) {
	m := &Message{
		Id:  getId(),
		Cmd: cmd,
	}
	m.SetData(obj)
	return m, nil
}

var idLock = sync.Mutex{}
var id = 0

func getId() int {
	idLock.Lock()
	ret := id
	id += 1
	idLock.Unlock()
	return ret
}
