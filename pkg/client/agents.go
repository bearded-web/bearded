package client

import (
	"fmt"

	"golang.org/x/net/context"
	"github.com/bearded-web/bearded/models/agent"
)

const (
	agentsUrl     = "agents"
	agentsJobsUrl = "jobs"
)

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
func (s *AgentsService) List(ctx context.Context, opt *AgentsListOpts) (*agent.AgentList, error) {
	agentList := &agent.AgentList{}
	return agentList, s.client.List(ctx, agentsUrl, opt, agentList)
}

func (s *AgentsService) Get(ctx context.Context, id string) (*agent.Agent, error) {
	agent := &agent.Agent{}
	return agent, s.client.Get(ctx, agentsUrl, id, agent)
}

func (s *AgentsService) Create(ctx context.Context, src *agent.Agent) (*agent.Agent, error) {
	pl := &agent.Agent{}
	return pl, s.client.Create(ctx, agentsUrl, src, pl)
}

func (s *AgentsService) Update(ctx context.Context, src *agent.Agent) (*agent.Agent, error) {
	pl := &agent.Agent{}
	id := fmt.Sprintf("%x", string(src.Id))
	return pl, s.client.Update(ctx, agentsUrl, id, src, pl)
}

func (s *AgentsService) GetJobs(ctx context.Context, src *agent.Agent) ([]*agent.Job, error) {
	jobs := []*agent.Job{}
	url := fmt.Sprintf("%s/%s/%s", agentsUrl, FromId(src.Id), agentsJobsUrl)
	return jobs, s.client.List(ctx, url, nil, &jobs)
}
