package manager

import (
	"github.com/bearded-web/bearded/models/issue"
	"github.com/bearded-web/bearded/models/vuln"
	vulndb "github.com/vulndb/vulndb-go"
	"github.com/vulndb/vulndb-go/bindata"
)

type VulndbManager struct {
	manager *Manager

	vulnList []*vuln.Vuln
}

func (m *VulndbManager) Init() error {
	vulnRawList, err := bindata.LoadFromBin()
	if err != nil {
		return err
	}
	vulnList := make([]*vuln.Vuln, 0, len(vulnRawList))
	for _, rawVuln := range vulnRawList {
		vulnList = append(vulnList, convertRawVuln(rawVuln))
	}
	m.vulnList = vulnList
	return nil
}

func (m *VulndbManager) GetVulns() []*vuln.Vuln {
	return m.vulnList
}

func (m *VulndbManager) GetById(id int) *vuln.Vuln {
	for _, vuln := range m.vulnList {
		if vuln.Id == id {
			return vuln
		}
	}
	return nil
}

func (m *VulndbManager) Copy(new *VulndbManager) {
	new.vulnList = m.vulnList
}

func convertRawVuln(rawVuln *vulndb.Vuln) *vuln.Vuln {
	v := &vuln.Vuln{
		Id:          rawVuln.Id,
		Title:       rawVuln.Title,
		Description: rawVuln.Description.String(),
		Severity:    convertVulnSeverity(rawVuln.Severity),
		Tags:        rawVuln.Tags,
		Wasc:        rawVuln.Wasc,
		Cwe:         rawVuln.Cwe,
		OwaspTop10:  rawVuln.OwaspTop10,
		Fix:         vuln.VulnFix{Guidance: rawVuln.Fix.Guidance.String(), Effort: rawVuln.Fix.Effort},
	}
	for _, ref := range rawVuln.References {
		v.References = append(v.References, vuln.Reference{Url: ref.Url, Title: ref.Title})
	}
	return v
}

var severityMap = map[string]issue.Severity{
	"info":          issue.SeverityInfo,
	"informational": issue.SeverityInfo,
	"low":           issue.SeverityLow,
	"medium":        issue.SeverityMedium,
	"high":          issue.SeverityHigh,
}

func convertVulnSeverity(sev string) issue.Severity {
	if severity, ok := severityMap[sev]; ok {
		return severity
	}
	return issue.SeverityError
}
