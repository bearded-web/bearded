package script

import (
	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/plugin"
)

type Scripter interface {
	Handle(context.Context, ClientV1, *plugin.Conf) error
}
