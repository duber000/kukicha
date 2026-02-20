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
	ctxpkg "github.com/duber000/kukicha/stdlib/ctx"

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

		// Skip symlinks to prevent including files outside the build context
		if d.Type()&os.ModeSymlink != 0 {
			return nil
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
// An optional ctx.Handle can be passed for cancellation support.
func Pull(engine Engine, ref string, handles ...ctxpkg.Handle) (string, error) {
	ctx := context.Background()
	if len(handles) > 0 {
		ctx = ctxpkg.Value(handles[0])
	}
	reader, err := engine.cli.ImagePull(ctx, ref, image.PullOptions{})
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

// Build builds a Docker image from a directory. Returns imageID and build output.
func Build(engine Engine, path string, tag string) (BuildOutput, error) {
	imageID, output, err := buildImage(engine.cli, path, tag)
	if err != nil {
		return BuildOutput{}, err
	}
	return BuildOutput{imageID: imageID, output: output}, nil
}

// LoginFromConfig loads registry credentials from ~/.docker/config.json.
func LoginFromConfig(server string) (Auth, error) {
	username, password, addr, err := loadDockerAuth(server)
	if err != nil {
		return Auth{}, err
	}
	return Auth{username: username, password: password, serverAddress: addr}, nil
}

// CopyFrom copies files from a container path to a local destination path.
// An optional ctx.Handle can be passed for cancellation support.
func CopyFrom(engine Engine, containerID string, sourcePath string, destPath string, handles ...ctxpkg.Handle) error {
	ctx := context.Background()
	if len(handles) > 0 {
		ctx = ctxpkg.Value(handles[0])
	}
	return copyFromWithContext(engine, ctx, containerID, sourcePath, destPath)
}

func copyFromWithContext(engine Engine, ctx context.Context, containerID string, sourcePath string, destPath string) error {
	reader, _, err := engine.cli.CopyFromContainer(ctx, containerID, sourcePath)
	if err != nil {
		return fmt.Errorf("container copy from: %w", err)
	}
	defer reader.Close()
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return fmt.Errorf("container copy from mkdir: %w", err)
	}
	if err := extractTar(reader, destPath); err != nil {
		return fmt.Errorf("container copy from extract: %w", err)
	}
	return nil
}

// CopyTo copies a local file or directory into a container destination directory.
// An optional ctx.Handle can be passed for cancellation support.
func CopyTo(engine Engine, containerID string, sourcePath string, destPath string, handles ...ctxpkg.Handle) error {
	ctx := context.Background()
	if len(handles) > 0 {
		ctx = ctxpkg.Value(handles[0])
	}
	return copyToWithContext(engine, ctx, containerID, sourcePath, destPath)
}

func copyToWithContext(engine Engine, ctx context.Context, containerID string, sourcePath string, destPath string) error {
	archive, err := createTarFromPath(sourcePath)
	if err != nil {
		return err
	}
	opts := dockercontainer.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}
	if err := engine.cli.CopyToContainer(ctx, containerID, destPath, archive, opts); err != nil {
		return fmt.Errorf("container copy to: %w", err)
	}
	return nil
}

func createTarFromPath(sourcePath string) (io.Reader, error) {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("container copy to abs path: %w", err)
	}
	// Use Lstat to detect symlinks at the top level
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("container copy to stat: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("container copy to: source path is a symlink: %s", sourcePath)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	baseDir := filepath.Base(sourcePath)

	addFile := func(path string, fi os.FileInfo) error {
		// Skip symlinks to prevent archiving files outside the source tree
		if fi.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(filepath.Dir(sourcePath), path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if fi.IsDir() {
			header.Name += "/"
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if fi.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	}

	if info.IsDir() {
		if err := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			// Skip symlinks to prevent traversal outside the source tree
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			fi, err := d.Info()
			if err != nil {
				return err
			}
			return addFile(path, fi)
		}); err != nil {
			return nil, fmt.Errorf("container copy to walk: %w", err)
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil, fmt.Errorf("container copy to header: %w", err)
		}
		header.Name = filepath.ToSlash(baseDir)
		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("container copy to write header: %w", err)
		}
		f, err := os.Open(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("container copy to open: %w", err)
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return nil, fmt.Errorf("container copy to write file: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("container copy to close tar: %w", err)
	}
	return &buf, nil
}

func extractTar(reader io.Reader, destPath string) error {
	tr := tar.NewReader(reader)
	cleanDest := filepath.Clean(destPath)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		cleanName := filepath.Clean(header.Name)
		target := filepath.Join(destPath, cleanName)
		if !strings.HasPrefix(target, cleanDest+string(filepath.Separator)) && filepath.Clean(target) != cleanDest {
			return fmt.Errorf("invalid archive path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("archive contains unsupported link entry: %s", header.Name)
		default:
			return fmt.Errorf("archive contains unsupported entry type %d: %s", header.Typeflag, header.Name)
		}
	}
	return nil
}
