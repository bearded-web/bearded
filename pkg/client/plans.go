package client

import (
	"fmt"
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
	Name string
}

// List plans.
//
//
func (s *PlansService) List(opt *PlansListOpts) (*plan.PlanList, error) {
	planList := &plan.PlanList{}
	return planList, s.client.List(plansUrl, opt, planList)
}

func (s *PlansService) Get(id string) (*plan.Plan, error) {
	plan := &plan.Plan{}
	return plan, s.client.Get(plansUrl, id, plan)
}

func (s *PlansService) Create(src *plan.Plan) (*plan.Plan, error) {
	pl := &plan.Plan{}
	return pl, s.client.Create(plansUrl, src, pl)
}

func (s *PlansService) Update(src *plan.Plan) (*plan.Plan, error) {
	pl := &plan.Plan{}
	id := fmt.Sprintf("%x", string(src.Id))
	return pl, s.client.Update(plansUrl, id, src, pl)
}
