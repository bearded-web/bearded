package script

import (
	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/plan"
)

type Scripter interface {
	Handle(context.Context, ClientV1, *plan.Conf) error
}
