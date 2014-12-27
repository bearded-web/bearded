package docker

// helper to work with docker

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.google.com/p/go.net/context"
	"github.com/samalba/dockerclient"
	"strings"
)

type Docker struct {
	client *dockerclient.DockerClient
	log    *log.Logger
}

type ContainerConfig dockerclient.ContainerConfig

func (c *ContainerConfig) String() string {
	return fmt.Sprintf("%s: %s", c.Image, strings.Join(c.Cmd, " "))
}

type ContainerResponse struct {
	Err error
	Log []byte
}

func New(log *log.Logger) (*Docker, error) {
	client, err := getDockerClient()
	if err != nil {
		return nil, err
	}
	return &Docker{
		client: client,
		log:    log,
	}, nil
}

func (d *Docker) PullImages(images []string) error {
	for _, image := range images {
		d.log.Printf("Pull image %s \n", image)
		if err := d.client.PullImage(image, nil); err != nil {
			return err
		}
		d.log.Print("successful \n")
	}
	return nil
}

func (d *Docker) RunImage(ctx context.Context, config *ContainerConfig) <-chan ContainerResponse {
	ch := make(chan ContainerResponse, 1)
	go func(ch chan<- ContainerResponse, docker *dockerclient.DockerClient, config *ContainerConfig) {
		resp := ContainerResponse{}
		// Create a container
		d.log.Printf("Create container %s \n", config)
		cfg := dockerclient.ContainerConfig(*config)
		containerId, err := docker.CreateContainer(&cfg, "")
		if err != nil {
			d.log.Printf("Failed: %v", err)
			resp.Err = err
			ch <- resp
			return
		}
		cprint := func(format string, opt ...interface{}) {
			d.log.Printf("[%s] %s\n", containerId[:6], fmt.Sprintf(format, opt...))
		}
		cprint("Created with config: %s", config)

		defer func() {
			docker.RemoveContainer(containerId, false)
			cprint("Removed")
		}()

		// Start the container
		err = docker.StartContainer(containerId, &dockerclient.HostConfig{})
		if err != nil {
			resp.Err = err
			ch <- resp
			return
		}
		cprint("Started")
		defer func(){
			docker.StopContainer(containerId, 1)
			cprint("Stopped")
		}()
		tmpCh := make(chan ContainerResponse, 1)
		go func(tmpCh chan<-ContainerResponse) {
			resp := ContainerResponse{}
			cprint("Waiting for log")
			logs, err := docker.ContainerLogs(containerId, &dockerclient.LogOptions{Follow: true, Stdout: true})
			if err != nil {
				resp.Err = err
				tmpCh <- resp
				return
			}
			defer logs.Close()
			body, err := ioutil.ReadAll(logs)
			if err != nil {
				resp.Err = err
				tmpCh <- resp
				return
			}
			resp.Log = body
			tmpCh <- resp
		}(tmpCh)
		select {
		case <- ctx.Done():
			cprint("Context done")
			resp.Err = ctx.Err()
		case tmpResp := <- tmpCh:
			resp.Err = tmpResp.Err
			resp.Log = tmpResp.Log
		}
		ch <- resp
		return

	}(ch, d.client, config)
	return ch
}

func getDockerClient() (*dockerclient.DockerClient, error) {
	var tlsConfig *tls.Config
	if os.Getenv("DOCKER_TLS_VERIFY") == "1" {
		certPath := os.Getenv("DOCKER_CERT_PATH")
		cert := filepath.Join(certPath, "cert.pem")
		key := filepath.Join(certPath, "key.pem")
		ca := filepath.Join(certPath, "ca.pem")
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{tlsCert}}
		if ca == "" {
			tlsConfig.InsecureSkipVerify = true
		} else {
			cert, err := ioutil.ReadFile(ca)
			if err != nil {
				return nil, err
			}
			caPool := x509.NewCertPool()
			if !caPool.AppendCertsFromPEM(cert) {
				return nil, fmt.Errorf("Could not add RootCA pem")
			}
			tlsConfig.RootCAs = caPool
		}
	}
	docker, err := dockerclient.NewDockerClient(os.Getenv("DOCKER_HOST"), tlsConfig)
	if err != nil {
		return nil, err
	}
	return docker, nil
}
