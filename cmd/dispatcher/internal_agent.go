package dispatcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/bearded-web/bearded/cmd/agent"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/utils"
)

// Run agent inside current process
// Start httptest local server
func RunInternalAgent(app http.Handler, token string) error {
	ts := httptest.NewServer(app)
	hostname, err := utils.GetHostname()
	if err != nil {
		return err
	}
	api := client.NewClient(fmt.Sprintf("%s/api/", ts.URL), nil)
	api.Token = token
	go func() {
		agent.ServeAgent(hostname, api)
	}()
	return nil
}
