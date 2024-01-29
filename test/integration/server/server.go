//go:build integration

package server

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bastionzero/opk-ssh/internal/projectpath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ServerContainer struct {
	testcontainers.Container
	Host     string
	Port     int
	User     string
	Password string
}

func RunOpkSshContainer(ctx context.Context, issuerHostIp string, issuerPort string, networkName string) (*ServerContainer, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       projectpath.Root,
			Dockerfile:    filepath.Join("test", "integration", "server", "debian_opkssh.Dockerfile"),
			PrintBuildLog: true,
			KeepImage:     true,
			BuildArgs:     make(map[string]*string),
		},
		ExposedPorts: []string{"22/tcp"},
		Networks: []string{
			networkName,
		},
		ExtraHosts:    []string{fmt.Sprintf("oidc.local:%v", issuerHostIp)},
		ImagePlatform: "linux/amd64",
		// Wait for SSH server to be running by attempting to connect
		//
		// https://stackoverflow.com/a/54364978
		WaitingFor: wait.ForExec(strings.Split("echo -e '\\x1dclose\\x0d' | telnet localhost 22", " ")).
			WithStartupTimeout(time.Second * 10).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}),
	}
	if issuerPort != "" {
		req.BuildArgs["ISSUER_PORT"] = &issuerPort
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "22")
	if err != nil {
		return nil, err
	}
	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}
	return &ServerContainer{
		Container: container,
		Host:      hostIP,
		Port:      mappedPort.Int(),
		User:      "test",
		Password:  "test",
	}, nil
}

func RunUbuntuContainer(ctx context.Context) (*ServerContainer, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    projectpath.Root,
			Dockerfile: filepath.Join("test", "integration", "server", "ubuntu.Dockerfile"),
			KeepImage:  true,
		},
		ExposedPorts:  []string{"22/tcp"},
		ImagePlatform: "linux/amd64",
		WaitingFor:    wait.ForExposedPort(),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "22")
	if err != nil {
		return nil, err
	}
	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}
	return &ServerContainer{
		Container: container,
		Host:      hostIP,
		Port:      mappedPort.Int(),
		User:      "test",
		Password:  "test",
	}, nil
}
