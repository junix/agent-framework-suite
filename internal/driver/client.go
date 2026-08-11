package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/junix/agent-framework-suite/internal/suite"
)

const streamLimit = 64 << 10

type Client struct {
	Path        string
	Participant string
}

func (client Client) ParticipantName() string { return client.Participant }

func (client Client) Version(ctx context.Context) (suite.ParticipantVersion, error) {
	var version suite.ParticipantVersion
	stdout, stderr, err := client.invoke(ctx, "version")
	if err != nil {
		return version, fmt.Errorf("%s version probe failed: %w; stderr=%s", client.Participant, err, stderr)
	}
	if err := decodeOne(stdout, &version); err != nil {
		return version, fmt.Errorf("%s version response: %w", client.Participant, err)
	}
	if version.Participant != client.Participant || version.Driver == "" || version.SubjectVersion == "" {
		return version, fmt.Errorf("%s version identity mismatch: %+v", client.Participant, version)
	}
	return version, nil
}

func (client Client) Run(ctx context.Context, caseID string) (suite.Observation, error) {
	var observation suite.Observation
	stdout, stderr, err := client.invoke(ctx, "run", caseID)
	if err != nil {
		return observation, fmt.Errorf("%s %s failed: %w; stderr=%s", client.Participant, caseID, err, stderr)
	}
	if err := decodeOne(stdout, &observation); err != nil {
		return observation, fmt.Errorf("%s %s response: %w; stderr=%s", client.Participant, caseID, err, stderr)
	}
	return observation, nil
}

func (client Client) invoke(ctx context.Context, args ...string) (string, string, error) {
	if client.Path == "" {
		return "", "", fmt.Errorf("driver path is empty")
	}
	info, err := os.Stat(client.Path)
	if err != nil {
		return "", "", err
	}
	if info.IsDir() {
		return "", "", fmt.Errorf("driver path is a directory")
	}
	home, err := os.MkdirTemp("", "agent-framework-suite-home-")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(home)
	command := exec.CommandContext(ctx, client.Path, args...)
	command.Env = []string{
		"HOME=" + home,
		"TMPDIR=" + os.TempDir(),
		"PATH=" + os.Getenv("PATH"),
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
		"TZ=UTC",
		"NO_COLOR=1",
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &limitWriter{buffer: &stdout, remaining: streamLimit}
	command.Stderr = &limitWriter{buffer: &stderr, remaining: streamLimit}
	err = command.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func decodeOne(raw string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	switch err := decoder.Decode(&extra); {
	case err == io.EOF:
		return nil
	case err == nil:
		return fmt.Errorf("more than one JSON value")
	default:
		return fmt.Errorf("invalid trailing data: %w", err)
	}
}

type limitWriter struct {
	buffer    *bytes.Buffer
	remaining int
}

func (writer *limitWriter) Write(data []byte) (int, error) {
	original := len(data)
	if writer.remaining > 0 {
		if len(data) > writer.remaining {
			data = data[:writer.remaining]
		}
		_, _ = writer.buffer.Write(data)
		writer.remaining -= len(data)
	}
	return original, nil
}

func DefaultPaths(root string) (python string, rust string) {
	return filepath.Join(root, "drivers", "python", "agent-framework-py-driver"), filepath.Join(root, ".bin", "agent-framework-rs-driver")
}

func ProbeAll(clients []Client, timeout time.Duration) (map[string]suite.ParticipantVersion, error) {
	versions := make(map[string]suite.ParticipantVersion, len(clients))
	for _, client := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		version, err := client.Version(ctx)
		cancel()
		if err != nil {
			return nil, err
		}
		versions[client.Participant] = version
	}
	return versions, nil
}
