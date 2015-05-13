package utils

import (
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/plugin"
)

type ExtraData struct {
	Plugins []*plugin.Plugin
	Plans   []*plan.Plan
}
