package script

import (
	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/models/plan"
)

type Scripter interface {
	Handle(context.Context, ClientV1, *plan.Conf) error
}
