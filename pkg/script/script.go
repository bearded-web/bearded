package script

import (
	"github.com/bearded-web/bearded/models/plan"
	"golang.org/x/net/context"
)

type Scripter interface {
	Handle(context.Context, ClientV1, *plan.Conf) error
}
