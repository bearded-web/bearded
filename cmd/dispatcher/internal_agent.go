package dispatcher

import (
	"net/http/httptest"
	"net/http"
	"github.com/bearded-web/bearded/cmd/agent"
	"github.com/bearded-web/bearded/pkg/utils"
	"github.com/bearded-web/bearded/pkg/client"
	"fmt"
)


func RunInternalAgent(app http.Handler) error {
	ts := httptest.NewServer(app)
	hostname, err := utils.GetHostname()
	if err != nil {
		return err
	}
	api := client.NewClient(fmt.Sprintf("%s/api/", ts.URL), nil)
	api.Token = "agent-token"
	go func() {
		agent.ServeAgent(hostname, api)
	}()
	return nil
}
