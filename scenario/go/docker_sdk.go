// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scenario

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type ComposeAction string

type DockerManager struct {
	cli *client.Client
}

const (
	ComposeUp   ComposeAction = "up"
	ComposeDown ComposeAction = "down"
)

// NewDockerManager initializes a client compatible with v28+ via API negotiation.
func NewDockerManager() (*DockerManager, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to init docker client: %w", err)
	}
	return &DockerManager{cli: c}, nil
}

// ContainerExec executes a command inside a container, demultiplexing stdout/stderr.
func (dm *DockerManager) ContainerExec(containerName string, cmd []string, env []string) (string, error) {
	ctx := context.Background()

	// v28: ExecConfig moved to container.ExecOptions
	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Env:          env,
		Cmd:          cmd,
	}

	execID, err := dm.cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return "", fmt.Errorf("exec create failed [%s]: %w", containerName, err)
	}

	resp, err := dm.cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("exec attach failed [%s]: %w", containerName, err)
	}
	defer resp.Close()

	var outBuf, errBuf bytes.Buffer
	if _, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader); err != nil && err != io.EOF {
		return "", fmt.Errorf("stream copy failed [%s]: %w", containerName, err)
	}

	inspect, err := dm.cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return "", fmt.Errorf("exec inspect failed [%s]: %w", containerName, err)
	}

	combined := outBuf.String() + errBuf.String()
	if inspect.ExitCode != 0 {
		return combined, fmt.Errorf("command failed (exit %d) in %s: %s", inspect.ExitCode, containerName, combined)
	}

	return combined, nil
}

func (dm *DockerManager) ContainerStart(containerName string) error {
	return dm.cli.ContainerStart(context.Background(), containerName, container.StartOptions{})
}

func (dm *DockerManager) ContainerStop(containerName string) error {
	timeout := 10
	return dm.cli.ContainerStop(context.Background(), containerName, container.StopOptions{Timeout: &timeout})
}

func (dm *DockerManager) ComposeCommand(action ComposeAction, projectDir, file, projectName string) (string, error) {
	composeArgs := []string{"-f", file, "-p", projectName, string(action)}

	switch action {
	case ComposeUp:
		composeArgs = append(composeArgs, "-d")
	case ComposeDown:
		// no extra flags
	default:
		return "", fmt.Errorf("unsupported compose action: %q", action)
	}

	binary, args, err := resolveComposeBinary(composeArgs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve compose binary: %w", err)
	}

	fmt.Println("\033[1m", ">", binary, strings.Join(args, " "), "\033[0m")

	cmd := exec.Command(binary, args...) //#nosec G204
	cmd.Dir = projectDir

	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	if len(out) > 0 {
		fmt.Println(outputStr)
	}

	if err != nil {
		return outputStr, fmt.Errorf("compose %s failed: %w\n%s", action, err, outputStr)
	}

	return outputStr, nil
}

func resolveComposeBinary(composeArgs []string) (binary string, args []string, err error) {
	// Check docker-compose standalone (covers v1, v2, v5+)
	out, e := exec.Command("docker-compose", "version").Output()
	if e == nil && len(out) > 0 {
		return "docker-compose", composeArgs, nil
	}

	// Fall back to docker compose plugin (v2 built into docker CLI)
	out, e = exec.Command("docker", "compose", "version").Output()
	if e == nil && len(out) > 0 {
		return "docker", append([]string{"compose"}, composeArgs...), nil
	}

	return "", nil, fmt.Errorf("neither 'docker-compose' nor 'docker compose' found in PATH")
}

var (
	dmInstance *DockerManager
	dmOnce     sync.Once
	dmErr      error
)

func getDockerManager() (*DockerManager, error) {
	dmOnce.Do(func() {
		dmInstance, dmErr = NewDockerManager()
	})
	return dmInstance, dmErr
}

// dockerCommand preserves the existing call signature used throughout scenario code.
func dockerCommand(args ...string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("insufficient docker args: %v", args)
	}

	dm, err := getDockerManager()
	if err != nil {
		return "", fmt.Errorf("docker manager unavailable: %w", err)
	}

	fmt.Println("\033[1m", ">", "docker", strings.Join(args, " "), "\033[0m")

	switch args[0] {
	case "start":
		return "", dm.ContainerStart(args[1])
	case "stop":
		return "", dm.ContainerStop(args[1])
	case "exec":
		containerName, cmd, env := parseExecArgs(args[1:])
		if containerName == "" || len(cmd) == 0 {
			return "", fmt.Errorf("invalid exec args: %v", args)
		}
		return dm.ContainerExec(containerName, cmd, env)
	default:
		return "", fmt.Errorf("unsupported docker action %q — use exec.Command for compose ops", args[0])
	}
}

// dockerCommandWithTLS appends TLS peer flags and delegates to dockerCommand.
func dockerCommandWithTLS(args ...string) (string, error) {
	tlsArgs := []string{
		"--tls",
		"--cafile",
		"/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem",
	}
	return dockerCommand(append(args, tlsArgs...)...)
}

func parseExecArgs(args []string) (containerName string, cmd []string, env []string) {
	i := 0

	for i < len(args) {
		if args[i] == "-e" && i+1 < len(args) {
			env = append(env, args[i+1])
			i += 2
			continue
		}
		break
	}

	if i >= len(args) {
		return "", nil, env
	}

	containerName = args[i]
	cmd = args[i+1:]
	return containerName, cmd, env
}
