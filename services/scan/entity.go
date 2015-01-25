package scan

import "github.com/bearded-web/bearded/models/scan"

type SessionUpdateEntity struct {
	Status scan.ScanStatus `json:"status" description:"one of [working|finished|failed]"`
}
