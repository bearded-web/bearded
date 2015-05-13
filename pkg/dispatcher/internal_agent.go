package dispatcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bearded-web/bearded/pkg/agent"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/config"
)

// Run agent inside current process
// Start httptest local server
func RunInternalAgent(app http.Handler, token string, cfg *config.Agent) error {
	ts := httptest.NewServer(app)
	api := client.NewClient(fmt.Sprintf("%s/api/", ts.URL), nil)
	api.Token = token
	if cfg.Name == "" {
		cfg.Name = "internal"
	}
	go func() {
		agent.ServeAgent(cfg, api)
	}()
	return nil
}
