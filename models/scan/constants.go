package scan

import "encoding/json"

type ScanStatus string

const (
	StatusCreated  ScanStatus = "created"
	StatusQueued   ScanStatus = "queued"  // put scan to queue
	StatusWorking  ScanStatus = "working" // scan was taken by agent
	StatusPaused   ScanStatus = "paused"
	StatusFinished ScanStatus = "finished"
	StatusFailed   ScanStatus = "failed"
)

var scanStatuses = []interface{}{
	StatusCreated,
	StatusQueued,
	StatusWorking,
	StatusPaused,
	StatusFinished,
	StatusFailed,
}

// It's a hack to show custom type as string in swagger
func (t ScanStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t ScanStatus) Enum() []interface{} {
	return scanStatuses
}

func (t ScanStatus) Convert(text string) (interface{}, error) {
	return ScanStatus(text), nil
}
