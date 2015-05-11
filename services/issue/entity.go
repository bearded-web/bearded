package issue

import (
	"net/http"

	"github.com/bearded-web/bearded/models/issue"
)

type StatusEntity struct {
	Confirmed *bool `json:"confirmed,omitempty"`
	False     *bool `json:"false,omitempty"`
	Muted     *bool `json:"muted,omitempty"`
	Resolved  *bool `json:"resolved,omitempty"`
}

type HeaderMyEntity struct {
	Key    string   `json:"key"`
	Values []string `json:"values,omitempty"`
}

type CommentEntity struct {
	Text string `json:"text" description:"raw markdown text"`
}

func TransformHeader(he []*HeaderMyEntity) http.Header {
	h := http.Header{}
	for _, heElem := range he {
		for _, val := range heElem.Values {
			h.Add(heElem.Key, val)
		}
	}
	return h
}

type HttpMyEntity struct {
	Status string `json:"status"`
	// The problem with header is that map isn't supported in swagger api
	// So we transform map[string][]string to []struct{key:string,values:[]string}
	Header []*HeaderMyEntity `json:"header"`
	Body   *issue.HttpBody   `json:"body,omitempty"`
}

func (he *HttpMyEntity) Transform() *issue.HttpEntity {
	if he == nil {
		return nil
	}
	dst := &issue.HttpEntity{
		Status: he.Status,
		Body:   he.Body,
		Header: TransformHeader(he.Header),
	}
	return dst
}

type HttpTransactionEntity struct {
	Id       int           `json:"id,omitempty"`
	Url      string        `json:"url,omitempty"`
	Params   []string      `json:"params,omitempty"`
	Method   string        `json:"method"`
	Request  *HttpMyEntity `json:"request,omitempty"`
	Response *HttpMyEntity `json:"response,omitempty"`
}

type VectorEntity struct {
	Url              string                   `json:"url,omitempty" description:"where this issue is happened"`
	HttpTransactions []*HttpTransactionEntity `json:"httpTransactions,omitempty" bson:"httpTransactions"`
}

func (v *VectorEntity) Transform() *issue.Vector {
	if v == nil {
		return nil
	}
	vec := &issue.Vector{
		Url: v.Url,
	}
	for _, tr := range v.HttpTransactions {
		httpTr := &issue.HttpTransaction{
			Id:       tr.Id,
			Url:      tr.Url,
			Params:   tr.Params,
			Method:   tr.Method,
			Request:  tr.Request.Transform(),
			Response: tr.Response.Transform(),
		}

		vec.HttpTransactions = append(vec.HttpTransactions, httpTr)
	}
	return vec
}

type IssueEntity struct {
	Summary    *string            `json:"summary,omitempty" creating:"nonzero,min=3,max=120"`
	VulnType   *int               `json:"vulnType,omitempty" bson:"vulnType" description:"vulnerability type from vulndb"`
	Severity   *issue.Severity    `json:"severity,omitempty" description:"one of [high medium low info]"`
	References []*issue.Reference `json:"references,omitempty" bson:"references" description:"information about vulnerability"`
	Desc       *string            `json:"desc,omitempty"`
	Vector     *VectorEntity      `json:"vector,omitempty"`
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
// Return true if target should rebuild summary for issues
func updateTargetIssue(raw *TargetIssueEntity, dst *issue.TargetIssue) bool {
	rebuildSummary := false
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
		dst.Vector = raw.Vector.Transform()
	}
	if raw.Confirmed != nil {
		dst.Confirmed = *raw.Confirmed
	}
	if raw.False != nil {
		rebuildSummary = true
		dst.False = *raw.False
	}
	if raw.Resolved != nil {
		rebuildSummary = true
		dst.Resolved = *raw.Resolved
	}
	if raw.Muted != nil {
		rebuildSummary = true
		dst.Muted = *raw.Muted
	}
	if raw.Severity != nil {
		if isValidSeverity(*raw.Severity) {
			rebuildSummary = true
			dst.Severity = *raw.Severity
		}
	}
	return rebuildSummary
}
