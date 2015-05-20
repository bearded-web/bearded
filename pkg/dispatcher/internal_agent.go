package dispatcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"golang.org/x/net/context"

	"github.com/bearded-web/bearded/pkg/agent"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/config"
	"github.com/bearded-web/bearded/pkg/utils/async"
)

// Run agent inside current process
// Start httptest local server
func RunInternalAgent(ctx context.Context, app http.Handler, token string, cfg *config.Agent) <-chan error {
	ts := httptest.NewServer(app)
	api := client.NewClient(fmt.Sprintf("%s/api/", ts.URL), nil)
	api.Token = token
	if cfg.Name == "" {
		cfg.Name = "internal"
	}
	return async.Promise(func() error {
		return agent.ServeAgent(ctx, cfg, api)
	})
}
