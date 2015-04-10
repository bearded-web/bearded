package issue

import "encoding/json"

type Severity string

const (
	SeverityInfo   = Severity("info")
	SeverityLow    = Severity("low")
	SeverityMedium = Severity("medium")
	SeverityHigh   = Severity("high")
	SeverityError  = Severity("error")
)

var severities = []interface{}{
	SeverityInfo,
	SeverityLow,
	SeverityMedium,
	SeverityHigh,
	SeverityError,
}

// It's a hack to show custom type as string in swagger
func (t Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t Severity) Enum() []interface{} {
	return severities
}

func (t Severity) Convert(text string) (interface{}, error) {
	return Severity(text), nil
}

//
//type Affect string
//
//const (
//	AffectServer = Affect("server")
//	AffectUser   = Affect("user")
//)

type ActivityType string

const (
	ActivityReported  = ActivityType("reported")  // the issue was reported by plugin
	ActivityConfirmed = ActivityType("confirmed") // the issue was confirmed by someone
	ActivityMuted     = ActivityType("muted")     // go away! I'll fix you later
	ActivityUnmuted   = ActivityType("unmuted")
	ActivityFalse     = ActivityType("false") // set to false
	ActivityTrue      = ActivityType("true")  // set to true
	ActivityResolved  = ActivityType("resolved")
	ActivityReopened  = ActivityType("reopened")
)

var activities = []interface{}{
	ActivityReported,
	ActivityConfirmed,
	ActivityMuted,
	ActivityUnmuted,
	ActivityFalse,
	ActivityTrue,
}

// It's a hack to show custom type as string in swagger
func (t ActivityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t ActivityType) Enum() []interface{} {
	return activities
}

func (t ActivityType) Convert(text string) (interface{}, error) {
	return ActivityType(text), nil
}
