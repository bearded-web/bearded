package issue

import (
	"github.com/bearded-web/bearded/models/issue"
)

type Status struct {
	Confirmed *bool           `json:"confirmed,omitempty"`
	False     *bool           `json:"false,omitempty"`
	Muted     *bool           `json:"muted,omitempty"`
	Resolved  *bool           `json:"resolved,omitempty"`
	Severity  *issue.Severity `json:"severity,omitempty" description:"one of [high medium low info]"`
}

type TargetIssueEntity struct {
	Status `json:",inline"`
}

func isValidSeverity(sev issue.Severity) bool {
	if sev == issue.SeverityHigh ||
		sev == issue.SeverityMedium ||
		sev == issue.SeverityLow ||
		sev == issue.SeverityInfo {

		return true
	}
	return false
}
