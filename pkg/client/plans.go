package client

import (
	"fmt"

	"code.google.com/p/go.net/context"
	"github.com/bearded-web/bearded/models/plan"
)

const plansUrl = "plans"

type PlansService struct {
	client *Client
}

func (s *PlansService) String() string {
	return Stringify(s)
}

type PlansListOpts struct {
	Name string `url:"name"`
}

// List plans.
//
//
func (s *PlansService) List(ctx context.Context, opt *PlansListOpts) (*plan.PlanList, error) {
	planList := &plan.PlanList{}
	return planList, s.client.List(ctx, plansUrl, opt, planList)
}

func (s *PlansService) Get(ctx context.Context, id string) (*plan.Plan, error) {
	plan := &plan.Plan{}
	return plan, s.client.Get(ctx, plansUrl, id, plan)
}

func (s *PlansService) Create(ctx context.Context, src *plan.Plan) (*plan.Plan, error) {
	pl := &plan.Plan{}
	return pl, s.client.Create(ctx, plansUrl, src, pl)
}

func (s *PlansService) Update(ctx context.Context, src *plan.Plan) (*plan.Plan, error) {
	pl := &plan.Plan{}
	id := fmt.Sprintf("%x", string(src.Id))
	return pl, s.client.Update(ctx, plansUrl, id, src, pl)
}
