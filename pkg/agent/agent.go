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
	"github.com/bearded-web/bearded/pkg/transport/mango"
	"github.com/davecgh/go-spew/spew"
	dockerclient "github.com/fsouza/go-dockerclient"
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
			scanCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			if err := a.HandleScan(scanCtx, job.Scan); err != nil {
				logrus.Error(stackerr.Wrap(err))
			}
		}()
	}
	return nil
}

func (a *Agent) HandleScan(ctx context.Context, sess *scan.Session) error {
	// take a plugin
	pl, err := a.api.Plugins.Get(ctx, client.FromId(sess.Plugin))
	if err != nil {
		return stackerr.Wrap(err)
	}
	logrus.Info("plugin: %s", pl)
	logrus.Info("set session to working state")
	sess.Status = scan.StatusWorking
	if sess, err = a.api.Scans.SessionUpdate(ctx, sess); err != nil {
		return err
	}

	setFailed := func() error {
		logrus.Info("set session to failed state")
		sess.Status = scan.StatusFailed
		if sess, err = a.api.Scans.SessionUpdate(ctx, sess); err != nil {
			return err
		}
		return nil
	}
	hostCfg := &dockerclient.HostConfig{}
	args := sess.Step.Conf.CommandArgs
	cfg := &dockerclient.Config{
		Image: pl.Container.Image,
		Tty:   true,
		Cmd:   strings.Split(args, " "),
	}

	switch pl.Type {
	case plugin.Util:
	case plugin.Script:
		if hostCfg.PortBindings == nil {
			hostCfg.PortBindings = map[dockerclient.Port][]dockerclient.PortBinding{}
		}
		hostCfg.PortBindings["9238/tcp"] = []dockerclient.PortBinding{dockerclient.PortBinding{HostIP: "127.0.0.1"}}
	default:
		return setFailed()
	}

	ch := a.dclient.RunImage(ctx, cfg, hostCfg)
	// creating container
	var container *dockerclient.Container
	select {
	case <-ctx.Done():
		return setFailed()
	case res := <-ch:
		// container info
		if res.Err != nil {
			logrus.Error(res.Err)
			return setFailed()
		}
		container = res.Container
	}
	spew.Dump(container)
	// get ports
	var serv *RemoteServer
	if pl.Type == plugin.Script {
		// setup transport between agent and script
		// script should expose 9238
		port := container.NetworkSettings.Ports["9238/tcp"][0].HostPort
		logrus.Infof("container port is %v", port)
		if port == "" {
			return setFailed()
		}
		//		transp, err := mango.NewDialer(fmt.Sprintf("tcp://127.0.0.1:%s", port))
		transp, err := mango.NewClient(fmt.Sprintf("tcp://127.0.0.1:9238"))
		if err != nil {
			logrus.Error(err)
			return setFailed()
		}
		// setup remote server
		serv, _ = NewRemoteServer(transp, a.api, sess)
		go transp.Serve(ctx, serv)
		err = serv.Connect(ctx)
		if err != nil {
			logrus.Error(err)
			return setFailed()
		}
	}

	var (
		res    docker.ContainerResponse
		closed bool
	)
	// running
	select {
	case <-ctx.Done():
		logrus.Error(stackerr.Wrap(ctx.Err()))
		return setFailed()
	case res = <-ch:
		if closed {
			// TODO (m0sth8): handle closed channel from docker container
			return setFailed()
		}
	}
	if res.Err != nil {
		logrus.Error(res.Err)
		return setFailed()
	}
	var rep *report.Report
	raw := &report.Report{
		Type: report.TypeRaw,
		Raw:  report.Raw{Raw: string(res.Log)},
	}
	rep = raw
	switch pl.Type {
	case plugin.Script:
		if serv != nil && serv.Rep != nil {
			if serv.Rep.Type != report.TypeMulti {
				rep = &report.Report{
					Type: report.TypeMulti,
				}
				rep.Multi = append(rep.Multi, serv.Rep)
			} else {
				rep = serv.Rep
			}

			rep.Multi = append(rep.Multi, raw)
		}
	}

	_, err = a.api.Scans.SessionReportCreate(ctx, sess, rep)
	if err != nil {
		logrus.Error(err)
		return setFailed()
	}

	logrus.Info("finished")
	sess.Status = scan.StatusFinished
	if sess, err = a.api.Scans.SessionUpdate(ctx, sess); err != nil {
		return err
	}
	return nil
}
