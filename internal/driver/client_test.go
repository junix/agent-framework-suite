package driver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionProbeValidatesIdentity(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "fake-driver")
	script := "#!/bin/sh\nprintf '%s\\n' '{\"driver\":\"fake\",\"participant\":\"python\",\"subject_version\":\"1\"}'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	client := Client{Path: path, Participant: "python"}
	version, err := client.Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if version.Driver != "fake" || version.SubjectVersion != "1" {
		t.Fatalf("unexpected version: %+v", version)
	}
}
