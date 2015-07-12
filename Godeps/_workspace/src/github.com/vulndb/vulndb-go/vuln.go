package vulndb

type Reference struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

type VulnFix struct {
	Guidance MultiLine `json:"guidance"`
	Effort   int       `json:"effort"`
}

type Vuln struct {
	Id          int              `json:"id"`
	Title       string           `json:"title"`
	Description MultiLine        `json:"description"`
	Severity    string           `json:"severity"`
	Tags        []string         `json:"tags,omitempty"`
	References  []Reference      `json:"references"`
	Wasc        []string         `json:"wasc,omitempty"`
	Cwe         []string         `json:"cwe,omitempty"`
	OwaspTop10  map[string][]int `json:"owasp_top_10,omitempty"`
	Fix         VulnFix          `json:"fix"`

	// filename is taken from file name
	Filename string
}

type VulnList []*Vuln

// Len is part of sort.Interface.
func (s VulnList) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s VulnList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface. It is implemented by comparing Id
func (s VulnList) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

// Get all vulnerabilites filtered by tag
func (s VulnList) FilterByTag(tag string) VulnList {
	if s == nil {
		return nil
	}
	vulns := VulnList{}
	if tag != "" {
		for _, vuln := range s {
			if vuln.Tags != nil {
				for _, vulnTag := range vuln.Tags {
					if vulnTag == tag {
						vulns = append(vulns, vuln)
						break
					}
				}
			}
		}
	}
	return vulns
}

// Get all vulnerabilites filtered by severity
func (s VulnList) FilterBySeverity(severity string) VulnList {
	if s == nil {
		return nil
	}
	vulns := VulnList{}
	if severity != "" {
		for _, vuln := range s {
			if vuln.Severity == severity {
				vulns = append(vulns, vuln)
			}
		}
	}
	return vulns
}

// Get vulnerability by id
func (s VulnList) GetById(id int) *Vuln {
	if s == nil {
		return nil
	}
	for _, vuln := range s {
		if vuln.Id == id {
			return vuln
		}
	}
	return nil
}
