package scan

import "encoding/json"

type ScanStatus string

const (
	StatusCreated  ScanStatus = "created"
	StatusQueued              = "queued"
	StatusWorking             = "working"
	StatusPaused              = "paused"
	StatusFinished            = "finished"
	StatusFailed              = "failed"
)

// It's a hack to show custom type as string in swagger
func (t ScanStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}
