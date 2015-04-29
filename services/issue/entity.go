package issue

import (
	"github.com/bearded-web/bearded/models/issue"
)

type StatusEntity struct {
	Confirmed *bool `json:"confirmed,omitempty"`
	False     *bool `json:"false,omitempty"`
	Muted     *bool `json:"muted,omitempty"`
	Resolved  *bool `json:"resolved,omitempty"`
}

type IssueEntity struct {
	Summary    *string            `json:"summary,omitempty" creating:"nonzero,min=3,max=120"`
	VulnType   *int               `json:"vulnType,omitempty" bson:"vulnType" description:"vulnerability type from vulndb"`
	Severity   *issue.Severity    `json:"severity,omitempty" description:"one of [high medium low info]"`
	References []*issue.Reference `json:"references,omitempty" bson:"references" description:"information about vulnerability"`
	Desc       *string            `json:"desc,omitempty"`
	Vector     *issue.Vector      `json:"vector,omitempty"`
}

type TargetIssueEntity struct {
	Target string `json:"target,omitempty" creating:"nonzero,bsonId"`

	StatusEntity `json:",inline"`
	IssueEntity  `json:",inline"`
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

// Update all fields for dst with entity data if they present
func updateTargetIssue(raw *TargetIssueEntity, dst *issue.TargetIssue) {
	if raw.Summary != nil {
		dst.Summary = *raw.Summary
	}
	if raw.Desc != nil {
		dst.Desc = *raw.Desc
	}
	if raw.References != nil {
		dst.References = raw.References
	}
	if raw.VulnType != nil {
		dst.VulnType = *raw.VulnType
	}
	if raw.Vector != nil {
		dst.Vector = raw.Vector
	}
	if raw.Confirmed != nil {
		dst.Confirmed = *raw.Confirmed
	}
	if raw.False != nil {
		dst.False = *raw.False
	}
	if raw.Resolved != nil {
		dst.Resolved = *raw.Resolved
	}
	if raw.Muted != nil {
		dst.Muted = *raw.Muted
	}
	if raw.Severity != nil {
		if isValidSeverity(*raw.Severity) {
			dst.Severity = *raw.Severity
		}
	}
}
