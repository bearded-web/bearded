package agent

import (
	"fmt"
	"github.com/bearded-web/bearded/models/scan"
)

type JobCmd string

const (
	CmdRepeat JobCmd = "repeat" // just repeat request
	CmdScan          = "scan"
)

type Job struct {
	Cmd JobCmd `json:"cmd" description:"one of [repeat|scan]"`

	Scan *scan.Session
}

func (j *Job) String() string {
	return fmt.Sprintf("%s", j.Cmd)
}
