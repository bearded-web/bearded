package issue

type Status struct {
	Confirmed *bool `json:"confirmed"`
	False     *bool `json:"false"`
	Muted     *bool `json:"muted"`
	Resolved  *bool `json:"resolved"`
}

type TargetIssueEntity struct {
	Status `json:",inline"`
}
