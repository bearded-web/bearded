package docker

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stackerr"
	dockerclient "github.com/fsouza/go-dockerclient"
)

type ContainerResponse struct {
	Container *dockerclient.Container
	Err       error
	Log       []byte
	Files     map[string]io.Reader
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

func (d *Docker) PullImage(name string) error {
	_, err := d.Client.InspectImage(name)
	if err == dockerclient.ErrNoSuchImage {
		logrus.Infof("pull image: %s", name)
		err = d.Client.PullImage(dockerclient.PullImageOptions{
			Repository:   name,
			OutputStream: os.Stdout,
		}, dockerclient.AuthConfiguration{})
	}
	return err
}

// RunImage returns 2 response, first with created container object, second with logs.
// I know it's kind of stupid. But I'll rewrite it later.
func (d *Docker) RunImage(ctx context.Context, config *dockerclient.Config,
	hostCfg *dockerclient.HostConfig, takeFiles []string) <-chan ContainerResponse {

	// TODO (m0sth8): rewrite this, please
	ch := make(chan ContainerResponse, 2)
	go func(ch chan<- ContainerResponse, config *dockerclient.Config, hostCfg *dockerclient.HostConfig) {
		resp := ContainerResponse{}
		// pull container
		if err := d.PullImage(config.Image); err != nil {
			logrus.Errorf("Failed: %v", err)
			resp.Err = err
			ch <- resp
			return
		}
		// Create a container
		logrus.Infof("Create container %s: %v", config.Image, config.Cmd)
		opts := dockerclient.CreateContainerOptions{
			Config:     config,
			HostConfig: hostCfg,
		}
		container, err := d.Client.CreateContainer(opts)
		if err != nil {
			logrus.Errorf("Failed: %v", err)
			resp.Err = err
			ch <- resp
			return
		}

		cprint := func(format string, opt ...interface{}) {
			logrus.Infof("[%s] %s", container.ID[:6], fmt.Sprintf(format, opt...))
		}
		cprint("Created")

		defer func() {
			d.Client.RemoveContainer(dockerclient.RemoveContainerOptions{
				ID: container.ID,
			})
			cprint("Removed")
		}()

		// Start the container
		err = d.Client.StartContainer(container.ID, opts.HostConfig)
		if err != nil {
			resp.Err = stackerr.Wrap(err)
			ch <- resp
			return
		}
		cprint("Started")
		defer func() {
			d.Client.StopContainer(container.ID, 5)
			cprint("Stopped")
		}()
		container, err = d.Client.InspectContainer(container.ID)
		if err != nil {
			logrus.Errorf("Failed: %v", err)
			resp.Err = err
			ch <- resp
			return
		}

		resp.Container = container
		ch <- resp

		respCh := make(chan ContainerResponse, 1)
		go func(respCh chan<- ContainerResponse) {
			resp := ContainerResponse{}
			cprint("Waiting for log")
			stdout := bytes.NewBuffer(nil)
			stderr := bytes.NewBuffer(nil)
			err := d.Client.Logs(dockerclient.LogsOptions{
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
				respCh <- resp
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
			respCh <- resp
		}(respCh)
		select {
		case <-ctx.Done():
			cprint("Context done")
			resp.Err = ctx.Err()
		case tmpResp := <-respCh:
			resp.Err = tmpResp.Err
			resp.Log = tmpResp.Log
		}
		if resp.Err == nil && takeFiles != nil && len(takeFiles) > 0 {
			files := map[string]io.Reader{}
			for _, fPath := range takeFiles {
				cprint("take file %s from container", fPath)
				// try to copy files from container
				buf := new(bytes.Buffer)
				err := d.Client.CopyFromContainer(dockerclient.CopyFromContainerOptions{
					Container:    container.ID,
					Resource:     fPath,
					OutputStream: buf,
				})
				if err != nil {
					logrus.Error(stackerr.Wrap(err))
					continue
				}
				tr := tar.NewReader(buf)
				// Iterate through the files in the archive.
				fName := path.Base(fPath)
				for {
					hdr, err := tr.Next()
					if err == io.EOF {
						// end of tar archive
						break
					}
					if err != nil {
						logrus.Error(stackerr.Wrap(err))
						break
					}
					if hdr.Typeflag != tar.TypeReg {
						continue
					}
					if hdr.Name != fName {
						continue
					}
					fBuf := new(bytes.Buffer)
					_, err = io.Copy(fBuf, tr)
					if err != nil {
						logrus.Error(stackerr.Wrap(err))
						break
					}
					files[fPath] = fBuf
				}
			}
			if len(files) > 0 {
				resp.Files = files
			}
		}

		ch <- resp
		return

	}(ch, config, hostCfg)
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
