package client

import (
	"fmt"
	"github.com/bearded-web/bearded/models/agent"
)

const agentsUrl = "agents"

type AgentsService struct {
	client *Client
}

func (s *AgentsService) String() string {
	return Stringify(s)
}

type AgentsListOpts struct {
	Name string     `url:"name"`
	Type agent.Type `url:"type"`
}

// List agents.
//
//
func (s *AgentsService) List(opt *AgentsListOpts) (*agent.AgentList, error) {
	agentList := &agent.AgentList{}
	return agentList, s.client.List(agentsUrl, opt, agentList)
}

func (s *AgentsService) Get(id string) (*agent.Agent, error) {
	agent := &agent.Agent{}
	return agent, s.client.Get(agentsUrl, id, agent)
}

func (s *AgentsService) Create(src *agent.Agent) (*agent.Agent, error) {
	pl := &agent.Agent{}
	return pl, s.client.Create(agentsUrl, src, pl)
}

func (s *AgentsService) Update(src *agent.Agent) (*agent.Agent, error) {
	pl := &agent.Agent{}
	id := fmt.Sprintf("%x", string(src.Id))
	return pl, s.client.Update(agentsUrl, id, src, pl)
}
