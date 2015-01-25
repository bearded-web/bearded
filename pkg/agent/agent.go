package agent

import (
	"fmt"
	"strings"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/agent"
	"github.com/bearded-web/bearded/models/plugin"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/docker"
)

type Agent struct {
	// client helps to communicate with bearded api
	api     *client.Client
	name    string
	dclient *docker.Docker
}

func New(api *client.Client, dclient *docker.Docker, name string) (*Agent, error) {
	a := &Agent{
		api:     api,
		name:    name,
		dclient: dclient,
	}
	return a, nil
}

func (a *Agent) Serve(ctx context.Context) error {
	var resultErr error

	agnt := &agent.Agent{
		Name:   a.name,
		Type:   agent.System,
		Status: agent.Undefined,
	}

loop:
	for {
		timeout := 0
		logrus.Infof("Agent status: %s", agnt.Status)
		switch agnt.Status {
		case agent.Undefined:
			err := a.Register(ctx, agnt)
			if err != nil {
				logrus.Errorf("Registration error: %v", err)
				timeout = 10
			}
			logrus.Infof("Agent Id: %s", client.FromId(agnt.Id))
		case agent.Registered:
			err := a.Retrieve(ctx, agnt)
			if err != nil {
				logrus.Errorf("Retrieve error: %v", err)
				timeout = 10
			}
			if agnt.Status == agent.Registered {
				timeout = 10
			}
		case agent.Approved:
			err := a.GetJobs(ctx, agnt)
			if err != nil {
				logrus.Errorf("GetJobs error: %v", err)
			}
			timeout = 10
		case agent.Blocked:
			resultErr = fmt.Errorf("Agent is blocked")
			break loop
		default:
			resultErr = fmt.Errorf("Unknown agent status: %s", agnt.Status)
			break loop
		}
		if timeout != 0 {
			logrus.Infof("Timeout: %d", timeout)
		}
		select {
		case <-ctx.Done():
			if ctx.Err() != context.Canceled {
				resultErr = ctx.Err()
			}
			break loop
		case <-time.After(time.Duration(timeout) * time.Second):

		}
	}
	return resultErr
}

func (a *Agent) Register(ctx context.Context, agnt *agent.Agent) error {
	logrus.Info("Register")
	if created, err := a.api.Agents.Create(ctx, agnt); err != nil {
		if !client.IsConflicted(err) {
			logrus.Error(err)
			return err
		}
		logrus.Info("Already existed")
		// agent is already existed
		// retrieve existed
		logrus.Info("Get by name and type")
		agentList, err := a.api.Agents.List(ctx, &client.AgentsListOpts{Name: agnt.Name, Type: agnt.Type})
		if err != nil {
			return err
		}
		if agentList.Count != 1 {
			err := fmt.Errorf("expected 1 agent, but actual is %d", agentList.Count)
			return err
		}
		*agnt = *agentList.Results[0]
	} else {
		*agnt = *created
	}
	return nil
}

func (a *Agent) Retrieve(ctx context.Context, agnt *agent.Agent) error {
	logrus.Info("Retrieve")
	if agnt.Id == "" {
		agnt.Status = agent.Undefined
		return fmt.Errorf("ageng.Id shouldn't be empty")
	}
	got, err := a.api.Agents.Get(ctx, client.FromId(agnt.Id))
	if err != nil {
		return err
	}
	*agnt = *got
	return nil
}

func (a *Agent) GetJobs(ctx context.Context, agnt *agent.Agent) error {
	logrus.Info("Request jobs")
	jobs, err := a.api.Agents.GetJobs(ctx, agnt)
	if err != nil {
		return err
	}
	logrus.Infof("Got %d jobs", len(jobs))
	for _, job := range jobs {
		if err := a.HandleJob(ctx, job); err != nil {
			// TODO (m0sth8): return scan failed status
			// what should I do if backend server is unavailable?
			logrus.Error(err)
		}
	}
	return nil
}

func (a *Agent) HandleJob(ctx context.Context, job *agent.Job) error {
	logrus.Infof("Job: %s", job)
	if job.Cmd == agent.CmdScan {
		if err := a.HandleScan(ctx, job.Scan); err != nil {
			return stackerr.Wrap(err)
		}
	}
	return nil
}

func (a *Agent) HandleScan(ctx context.Context, sc *scan.Session) error {
	// take a plugin
	pl, err := a.api.Plugins.Get(ctx, client.FromId(sc.Plugin))
	if err != nil {
		return stackerr.Wrap(err)
	}
	logrus.Info("plugin: %s", pl)
	logrus.Info("set session to working state")
	sc.Status = scan.StatusWorking
	if sc, err = a.api.Scans.SessionUpdate(ctx, sc); err != nil {
		return err
	}

	setFailed := func() error {
		logrus.Info("set session to failed state")
		sc.Status = scan.StatusFailed
		if sc, err = a.api.Scans.SessionUpdate(ctx, sc); err != nil {
			return err
		}
		return nil
	}

	if pl.Type == plugin.Util {
		args := sc.Step.Conf.CommandArgs
		container := pl.Container
		cfg := docker.ContainerConfig{
			Image: container.Image,
			Tty:   true,
			Cmd:   strings.Split(args, " "),
		}
		ch := a.dclient.RunImage(ctx, &cfg)
		select {
		case <-ctx.Done():
		case res := <-ch:
			if res.Err != nil {
				logrus.Error(res.Err)
				return setFailed()
			}
			//			t.Report = &report.Report{
			//				Type: report.TypeRaw,
			//				Raw: string(res.Log),
			//			}
		}

	}

	time.Sleep(time.Second * 20)
	logrus.Info("finished")
	sc.Status = scan.StatusFinished
	if sc, err = a.api.Scans.SessionUpdate(ctx, sc); err != nil {
		return err
	}
	return nil
}
