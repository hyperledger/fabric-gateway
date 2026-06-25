package scenario

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type ComposeAction string

type DockerManager struct {
	cli *client.Client
}

// NewDockerManager initializes a client compatible with v28+ via API negotiation.
// its created to made the client compadible with previous versions of docker engine
func NewDockerManager() (*DockerManager, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to init docker client: %w", err)
	}
	return &DockerManager{cli: c}, nil
}

func (dm *DockerManager) ContainerExec(containerName string, cmd []string, env []string) (string, error) {
	ctx := context.Background()
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

// uses the docker compose sdk v5
func ComposeCommand(action ComposeAction, fixturesDir string, dockerComposeFile string, projectname string) error {

	configpath := []string{fixturesDir + "/" + dockerComposeFile}

	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		log.Fatalf("Failed to create docker CLI: %v", err)
		return err
	}
	err = dockerCLI.Initialize(&flags.ClientOptions{})
	if err != nil {
		log.Fatalf("Failed to initialize docker CLI: %v", err)
		return err
	}
	service, err := compose.NewComposeService(dockerCLI)
	if err != nil {
		log.Fatalf("Failed to create compose service: %v", err)
		return err
	}
	project, err := service.LoadProject(context.Background(), api.ProjectLoadOptions{
		ConfigPaths: configpath,
		ProjectName: projectname,
	})
	if err != nil {
		log.Fatalf("Failed to load project: %v", err)
		return err
	}

	switch action {
	case "ComposeUp":

		err = service.Up(context.Background(), project, api.UpOptions{
			Create: api.CreateOptions{},
			Start:  api.StartOptions{},
		})
		if err != nil {
			log.Fatalf("Failed to start services: %v", err)
			return err
		}

	case "ComposeDown":
		err = service.Down(context.Background(), projectname, api.DownOptions{})
		if err != nil {
			log.Fatalf("Failed to stop services: %v", err)
			return err
		}
	}

	log.Printf("Successfully %s project: %s", action, project.Name)
	return nil

}
