package docker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/go.net/context"
	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stackerr"
	dockerclient "github.com/fsouza/go-dockerclient"
)

type ContainerConfig dockerclient.Config

func (c *ContainerConfig) String() string {
	return fmt.Sprintf("%s: %s", c.Image, strings.Join(c.Cmd, " "))
}

type ContainerResponse struct {
	Err error
	Log []byte
}

type Docker struct {
	Client *dockerclient.Client
}

func NewDocker() (*Docker, error) {
	client, err := GetDockerClient()
	if err != nil {
		return nil, err
	}
	return &Docker{
		Client: client,
	}, nil
}

// Ping the docker server
//
// See http://goo.gl/stJENm for more details.
func (d *Docker) Ping() error {
	return d.Client.Ping()
}

//func (d *Docker) PullImages(images []string) error {
//	for _, image := range images {
//		logrus.Debugf("Pull image %s", image)
//		if err := d.Client.PullImage(image, nil); err != nil {
//			return err
//		}
//		d.log.Print("successful \n")
//	}
//	return nil
//}

func (d *Docker) RunImage(ctx context.Context, config *ContainerConfig) <-chan ContainerResponse {
	ch := make(chan ContainerResponse, 1)
	go func(ch chan<- ContainerResponse, docker *dockerclient.Client, config *ContainerConfig) {
		resp := ContainerResponse{}
		// Create a container
		logrus.Infof("Create container %s \n", config)
		cfg := dockerclient.Config(*config)
		opts := dockerclient.CreateContainerOptions{
			Config: &cfg,
		}
		container, err := docker.CreateContainer(opts)
		if err != nil {
			logrus.Errorf("Failed: %v", err)
			resp.Err = err
			ch <- resp
			return
		}
		cprint := func(format string, opt ...interface{}) {
			logrus.Infof("[%s] %s\n", container.ID[:6], fmt.Sprintf(format, opt...))
		}
		cprint("Created with config: %s", config)

		defer func() {
			docker.RemoveContainer(dockerclient.RemoveContainerOptions{
				ID: container.ID,
			})
			cprint("Removed")
		}()

		// Start the container
		err = docker.StartContainer(container.ID, opts.HostConfig)
		if err != nil {
			resp.Err = stackerr.Wrap(err)
			ch <- resp
			return
		}
		cprint("Started")
		defer func() {
			docker.StopContainer(container.ID, 5)
			cprint("Stopped")
		}()
		tmpCh := make(chan ContainerResponse, 1)
		go func(tmpCh chan<- ContainerResponse) {
			resp := ContainerResponse{}
			cprint("Waiting for log")
			stdout := bytes.NewBuffer(nil)
			stderr := bytes.NewBuffer(nil)
			err := docker.Logs(dockerclient.LogsOptions{
				OutputStream: stdout,
				ErrorStream:  stderr,
				Container:    container.ID,
				Follow:       true,
				Stdout:       true,
				Stderr:       true,
				RawTerminal:  true,
			})
			if err != nil {
				resp.Err = stackerr.Wrap(err)
				tmpCh <- resp
				return
			}
			serr := stderr.Bytes()
			sout := stdout.Bytes()
			if serr != nil {
				cprint("STDERR: %s", string(serr))
			}
			if sout != nil {
				cprint("STDOUT: %s", string(sout))
			}
			resp.Log = sout
			tmpCh <- resp
		}(tmpCh)
		select {
		case <-ctx.Done():
			cprint("Context done")
			resp.Err = ctx.Err()
		case tmpResp := <-tmpCh:
			resp.Err = tmpResp.Err
			resp.Log = tmpResp.Log
		}
		ch <- resp
		return

	}(ch, d.Client, config)
	return ch
}

func GetDockerClient() (*dockerclient.Client, error) {
	var (
		client *dockerclient.Client
		err    error
	)
	endpoint := os.Getenv("DOCKER_HOST")
	if os.Getenv("DOCKER_TLS_VERIFY") == "1" {
		certPath := os.Getenv("DOCKER_CERT_PATH")
		cert := filepath.Join(certPath, "cert.pem")
		key := filepath.Join(certPath, "key.pem")
		ca := filepath.Join(certPath, "ca.pem")
		client, err = dockerclient.NewTLSClient(endpoint, cert, key, ca)
	} else {
		client, err = dockerclient.NewClient(endpoint)
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}
