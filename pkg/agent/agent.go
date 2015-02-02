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
	"github.com/bearded-web/bearded/models/report"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/docker"
	"github.com/bearded-web/bearded/pkg/script/mango"
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
		Status: agent.StatusUndefined,
	}

loop:
	for {
		timeout := 0
		logrus.Debugf("Agent status: %s", agnt.Status)
		switch agnt.Status {
		case agent.StatusUndefined:
			err := a.Register(ctx, agnt)
			if err != nil {
				logrus.Errorf("Registration error: %v", err)
				timeout = 5
			}
			logrus.Infof("Agent Id: %s", client.FromId(agnt.Id))
		case agent.StatusRegistered:
			err := a.Retrieve(ctx, agnt)
			if err != nil {
				logrus.Errorf("Retrieve error: %v", err)
				timeout = 5
			}
			if agnt.Status == agent.StatusRegistered {
				timeout = 5
			}
		case agent.StatusApproved:
			err := a.GetJobs(ctx, agnt)
			if err != nil {
				logrus.Errorf("GetJobs error: %v", err)
				timeout = 5
			}
		case agent.StatusBlocked:
			resultErr = fmt.Errorf("Agent is blocked")
			break loop
		default:
			resultErr = fmt.Errorf("Unknown agent status: %s", agnt.Status)
			break loop
		}
		if timeout != 0 {
			logrus.Debugf("Timeout: %d", timeout)
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
		agnt.Status = agent.StatusUndefined
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
	logrus.Debug("Request jobs")
	jobs, err := a.api.Agents.GetJobs(ctx, agnt)
	if err != nil {
		return err
	}
	logrus.Debugf("Got %d jobs", len(jobs))
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
	logrus.Debugf("Job: %s", job)
	if job.Cmd == agent.CmdScan {
		go func() {
			if err := a.HandleScan(ctx, job.Scan); err != nil {
				logrus.Error(stackerr.Wrap(err))
			}
		}()
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
	var cfg docker.ContainerConfig
	switch pl.Type {
	case plugin.Util:
		args := sc.Step.Conf.CommandArgs
		container := pl.Container
		cfg = docker.ContainerConfig{
			Image: container.Image,
			Tty:   true,
			Cmd:   strings.Split(args, " "),
		}
	case plugin.Script:

		mangoServer, err := mango.NewServer("ipc:///tmp/bearded-web.socket")
		if err != nil {
			return stackerr.Wrap(err)
		}
		defer mangoServer.Stop()
	// make mango server with unix socket
	// run script container with unix socket
	// take raw report from logs
	//
	default:
		return setFailed()
	}

	ch := a.dclient.RunImage(ctx, &cfg)
	select {
	case <-ctx.Done():
	case res := <-ch:
		if res.Err != nil {
			logrus.Error(res.Err)
			return setFailed()
		}
		rep := &report.Report{
			Type: report.TypeRaw,
			Raw:  report.Raw{Raw: string(res.Log)},
		}
		_, err := a.api.Scans.SessionReportCreate(ctx, sc, rep)
		if err != nil {
			logrus.Error(err)
			return setFailed()
		}
	}

	logrus.Info("finished")
	sc.Status = scan.StatusFinished
	if sc, err = a.api.Scans.SessionUpdate(ctx, sc); err != nil {
		return err
	}
	return nil
}
