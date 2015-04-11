package agent

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/bearded-web/bearded/models/plan"
	"github.com/bearded-web/bearded/models/report"
	"github.com/bearded-web/bearded/models/scan"
	"github.com/bearded-web/bearded/pkg/agent/api"
	"github.com/bearded-web/bearded/pkg/client"
	"github.com/bearded-web/bearded/pkg/transport"
	"golang.org/x/net/context"
)

type RemoteServer struct {
	transp transport.Transport
	Rep    *report.Report
	sess   *scan.Session
	api    *client.Client
}

// Server connects agent with plugin script
func NewRemoteServer(transp transport.Transport, api *client.Client, sess *scan.Session) (*RemoteServer, error) {
	server := &RemoteServer{
		transp: transp,
		sess:   sess,
		api:    api,
	}
	return server, nil
}

func (s *RemoteServer) Handle(ctx context.Context, msg transport.Extractor) (interface{}, error) {
	req := api.RequestV1{}
	resp := api.ResponseV1{}
	err := msg.Extract(&req)
	if err != nil {
		return nil, err
	}

	switch req.Method {
	case api.GetConfig:
		if data, err := s.GetConfig(ctx); err != nil {
			return nil, err
		} else {
			resp.GetConfig = data
		}

	case api.GetPluginVersions:
		if data, err := s.GetPluginVersions(ctx, req.GetPluginVersions); err != nil {
			return nil, err
		} else {
			resp.GetPluginVersions = data
		}

	case api.RunPlugin:
		if data, err := s.RunPlugin(ctx, req.RunPlugin); err != nil {
			return nil, err
		} else {
			resp.RunPlugin = data
		}

	case api.SendReport:
		if err := s.SendReport(ctx, req.SendReport); err != nil {
			return nil, err
		}
	case api.DownloadFile:
		if data, err := s.DownloadFile(ctx, req.DownloadFile); err != nil {
			return nil, err
		} else {
			resp.DownloadFile = data
		}
	default:
		return nil, fmt.Errorf("Unknown method requested %s", req.Method)
	}

	return resp, nil
}

func (s *RemoteServer) Connect(ctx context.Context) error {
	req := api.RequestV1{Method: api.Connect}
	resp := api.ResponseV1{}
	return s.transp.Request(ctx, req, &resp)
}

// apiV1 methods

func (s *RemoteServer) GetPluginVersions(ctx context.Context, name string) ([]string, error) {
	// TODO (m0sth8): add sort by -version and increase limit
	plugins, err := s.api.Plugins.List(ctx, &client.PluginsListOpts{Name: name})
	if err != nil {
		return nil, err
	}
	versions := []string{}
	for _, pl := range plugins.Results {
		versions = append(versions, pl.Version)
	}
	return versions, nil
}

func (s *RemoteServer) GetConfig(ctx context.Context) (*plan.Conf, error) {
	return s.sess.Step.Conf, nil
}

func (s *RemoteServer) RunPlugin(ctx context.Context, step *plan.WorkflowStep) (*report.Report, error) {
	// create session inside current session
	child := &scan.Session{
		Scan:   s.sess.Scan,
		Parent: s.sess.Id,
		Step:   step,
	}
	sess, err := s.api.Scans.SessionAddChild(ctx, child)
	if err != nil {
		return nil, err
	}
	// wait for session status finished
loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		sess2, err := s.api.Scans.SessionGet(ctx, client.FromId(sess.Scan), client.FromId(sess.Id))
		if err != nil {
			if client.IsNotFound(err) {
				// why session is not found???
				return nil, err
			}
			logrus.Error(err)
			// TODO (m0sth8): add errors counter and fail if counter is more then maximum
			// TODO (m0sth8): add exponential sleep on errors
			time.Sleep(time.Second * 15)
			continue loop
		}
		switch sess2.Status {
		case scan.StatusFinished:
			break loop
		case scan.StatusFailed:
			return nil, fmt.Errorf("session was failed")
		case scan.StatusPaused:
			time.Sleep(time.Second * 30)
			continue loop
		}
		time.Sleep(time.Second * 2)
	}

	rep, err := s.api.Scans.SessionReportGet(ctx, sess)
	if err != nil {
		return nil, err
	}
	return rep, nil
}

func (s *RemoteServer) SendReport(ctx context.Context, rep *report.Report) error {
	s.Rep = rep
	return nil
}

func (s *RemoteServer) DownloadFile(ctx context.Context, fileId string) ([]byte, error) {
	buf, err := s.api.Files.Download(ctx, fileId)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(buf)
}
