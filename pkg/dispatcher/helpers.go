package dispatcher

import "github.com/bearded-web/bearded/pkg/manager"

// Get or create token by default agent email. The token is used only by internal agent.
func getAgentToken(mgr *manager.Manager) (string, error) {
	u, err := mgr.Users.GetByEmail(manager.AgentEmail)
	if err != nil {
		return "", err
	}
	token, err := mgr.Tokens.GetOrCreate(u.Id)
	if err != nil {
		return "", err
	}
	return token.Hash, nil
}
