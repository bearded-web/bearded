package tech

import (
	"github.com/bearded-web/bearded/models/tech"
)

type TechEntity struct {
	Categories *[]tech.Category `json:"categories,omitempty"`
	Name       *string          `json:"name"`
	Version    *string          `json:"version"`
	Confidence *int             `json:"confidence"`
	Url        *string          `json:"url" description:"url to technology"`
}

type TargetTechEntity struct {
	Target string          `json:"target,omitempty" creating:"nonzero,bsonId"`
	Status tech.StatusType `json:"status,omitempty"`

	TechEntity `json:",inline"`
}

// Update all fields for dst with entity data if they present
func updateTargetTech(raw *TargetTechEntity, dst *tech.TargetTech) {
	if raw.Name != nil {
		dst.Name = *raw.Name
	}
	if raw.Url != nil {
		dst.Url = *raw.Url
	}
	if raw.Version != nil {
		dst.Version = *raw.Version
	}
	if raw.Confidence != nil {
		dst.Confidence = *raw.Confidence
	}
	if raw.Categories != nil {
		dst.Categories = *raw.Categories
	}
	if raw.Status != "" && (raw.Status == tech.StatusCorrect ||
		raw.Status == tech.StatusIncorrect ||
		raw.Status == tech.StatusUnknown) {

		dst.Status = raw.Status
	}
}
