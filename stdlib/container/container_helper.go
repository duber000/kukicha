package container

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	types "github.com/docker/docker/api/types"
)

// newClient creates a Docker client with functional options.
// If host is empty, it uses the DOCKER_HOST env var or the default socket.
// Always enables API version negotiation.
func newClient(host string) (*client.Client, error) {
	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("container connect: %w", err)
	}
	return cli, nil
}

// buildImage creates a tar archive from contextPath, calls ImageBuild,
// and parses the JSON stream for the final image ID.
func buildImage(cli *client.Client, contextPath string, tag string) (string, string, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	contextPath, err := filepath.Abs(contextPath)
	if err != nil {
		return "", "", fmt.Errorf("container build: %w", err)
	}

	err = filepath.WalkDir(contextPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(contextPath, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		return "", "", fmt.Errorf("container build context: %w", err)
	}
	if err := tw.Close(); err != nil {
		return "", "", fmt.Errorf("container build tar: %w", err)
	}

	opts := types.ImageBuildOptions{
		Tags:       []string{tag},
		Remove:     true,
		Dockerfile: "Dockerfile",
	}

	resp, err := cli.ImageBuild(context.Background(), &buf, opts)
	if err != nil {
		return "", "", fmt.Errorf("container build: %w", err)
	}
	defer resp.Body.Close()

	var output strings.Builder
	var imageID string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line)
		output.WriteString("\n")

		var msg struct {
			Stream string `json:"stream"`
			Aux    struct {
				ID string `json:"ID"`
			} `json:"aux"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Error != "" {
			return "", output.String(), fmt.Errorf("container build: %s", msg.Error)
		}
		if msg.Aux.ID != "" {
			imageID = msg.Aux.ID
		}
	}
	if err := scanner.Err(); err != nil {
		return "", output.String(), fmt.Errorf("container build stream: %w", err)
	}

	return imageID, output.String(), nil
}

// containerLogs retrieves logs using stdcopy.StdCopy to demux stdout/stderr.
// The tail parameter controls how many lines to return ("" for all, or a number string).
func containerLogs(cli *client.Client, containerID string, tail string) (string, error) {
	opts := dockercontainer.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	if tail != "" {
		opts.Tail = tail
	}

	reader, err := cli.ContainerLogs(context.Background(), containerID, opts)
	if err != nil {
		return "", fmt.Errorf("container logs: %w", err)
	}
	defer reader.Close()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	if err != nil {
		// If stdcopy fails (e.g., TTY mode), fall back to raw read
		raw, readErr := io.ReadAll(reader)
		if readErr != nil {
			return "", fmt.Errorf("container logs: %w", err)
		}
		return string(raw), nil
	}

	combined := stdout.String()
	if stderr.Len() > 0 {
		combined += stderr.String()
	}
	return combined, nil
}

// loadDockerAuth reads ~/.docker/config.json and resolves credentials
// for the given registry server address.
// Returns (username, password, serverAddress, error).
func loadDockerAuth(serverAddress string) (string, string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", fmt.Errorf("container auth: %w", err)
	}

	configPath := filepath.Join(home, ".docker", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", "", fmt.Errorf("container auth: %w", err)
	}

	var config struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", "", "", fmt.Errorf("container auth parse: %w", err)
	}

	auth, ok := config.Auths[serverAddress]
	if !ok {
		variations := []string{
			"https://" + serverAddress,
			"http://" + serverAddress,
			strings.TrimPrefix(serverAddress, "https://"),
			strings.TrimPrefix(serverAddress, "http://"),
		}
		for _, v := range variations {
			if a, found := config.Auths[v]; found {
				auth = a
				ok = true
				break
			}
		}
	}

	if !ok {
		return "", "", "", fmt.Errorf("container auth: no credentials found for %s", serverAddress)
	}

	if auth.Auth == "" {
		return "", "", "", fmt.Errorf("container auth: empty credentials for %s", serverAddress)
	}

	decoded, err := base64.StdEncoding.DecodeString(auth.Auth)
	if err != nil {
		return "", "", "", fmt.Errorf("container auth decode: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("container auth: invalid credential format for %s", serverAddress)
	}

	return parts[0], parts[1], serverAddress, nil
}

// --- Bridge functions called from Kukicha ---

// Connect creates an Engine using auto-detected socket or DOCKER_HOST.
func Connect() (Engine, error) {
	// Auto-detect Podman socket first, then Docker
	socketPaths := []string{
		// Podman rootless
		fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid()),
		// Docker default
		"/var/run/docker.sock",
		// Podman root
		"/run/podman/podman.sock",
	}

	host := ""
	for _, path := range socketPaths {
		if _, err := os.Stat(path); err == nil {
			host = "unix://" + path
			break
		}
	}

	cli, err := newClient(host)
	if err != nil {
		return Engine{}, err
	}
	return Engine{cli: cli}, nil
}

// ConnectRemote creates an Engine connected to a specific Docker host.
func ConnectRemote(host string) (Engine, error) {
	cli, err := newClient(host)
	if err != nil {
		return Engine{}, err
	}
	return Engine{cli: cli}, nil
}

// Open creates an Engine from the builder configuration.
func Open(cfg Config) (Engine, error) {
	opts := []client.Opt{
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	}
	if cfg.host != "" {
		opts = append(opts, client.WithHost(cfg.host))
	}
	if cfg.apiVersion != "" {
		opts = append(opts, client.WithVersion(cfg.apiVersion))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return Engine{}, fmt.Errorf("container open: %w", err)
	}
	return Engine{cli: cli}, nil
}

// Pull pulls an image from a registry. Returns the image digest.
func Pull(engine Engine, ref string) (string, error) {
	reader, err := engine.cli.ImagePull(context.Background(), ref, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("container pull: %w", err)
	}
	defer reader.Close()

	// Consume the stream to complete the pull; extract digest from status messages
	var digest string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var msg struct {
			Status string `json:"status"`
			ID     string `json:"id"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
			if strings.HasPrefix(msg.Status, "Digest:") {
				digest = strings.TrimPrefix(msg.Status, "Digest: ")
			}
		}
	}
	if digest == "" {
		digest = ref
	}
	return digest, nil
}

// PullAuth pulls an image using registry credentials.
func PullAuth(engine Engine, ref string, auth Auth) (string, error) {
	authJSON, err := json.Marshal(map[string]string{
		"username":      auth.username,
		"password":      auth.password,
		"serveraddress": auth.serverAddress,
	})
	if err != nil {
		return "", fmt.Errorf("container pull auth: %w", err)
	}
	encoded := base64.URLEncoding.EncodeToString(authJSON)

	reader, err := engine.cli.ImagePull(context.Background(), ref, image.PullOptions{
		RegistryAuth: encoded,
	})
	if err != nil {
		return "", fmt.Errorf("container pull: %w", err)
	}
	defer reader.Close()

	var digest string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var msg struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil {
			if strings.HasPrefix(msg.Status, "Digest:") {
				digest = strings.TrimPrefix(msg.Status, "Digest: ")
			}
		}
	}
	if digest == "" {
		digest = ref
	}
	return digest, nil
}

// Run creates and starts a container. Returns the container ID.
func Run(engine Engine, img string, cmd []string) (string, error) {
	resp, err := engine.cli.ContainerCreate(
		context.Background(),
		&dockercontainer.Config{
			Image: img,
			Cmd:   cmd,
		},
		nil, nil, nil, "",
	)
	if err != nil {
		return "", fmt.Errorf("container run create: %w", err)
	}

	err = engine.cli.ContainerStart(context.Background(), resp.ID, dockercontainer.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("container run start: %w", err)
	}

	return resp.ID, nil
}

// Build builds a Docker image from a directory. Returns imageID and build output.
func Build(engine Engine, path string, tag string) (BuildOutput, error) {
	imageID, output, err := buildImage(engine.cli, path, tag)
	if err != nil {
		return BuildOutput{}, err
	}
	return BuildOutput{imageID: imageID, output: output}, nil
}

// Logs retrieves all logs from a container.
func Logs(engine Engine, containerID string) (string, error) {
	return containerLogs(engine.cli, containerID, "")
}

// LogsTail retrieves the last N lines of logs from a container.
func LogsTail(engine Engine, containerID string, lines int64) (string, error) {
	return containerLogs(engine.cli, containerID, fmt.Sprintf("%d", lines))
}

// LoginFromConfig loads registry credentials from ~/.docker/config.json.
func LoginFromConfig(server string) (Auth, error) {
	username, password, addr, err := loadDockerAuth(server)
	if err != nil {
		return Auth{}, err
	}
	return Auth{username: username, password: password, serverAddress: addr}, nil
}

