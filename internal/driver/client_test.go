package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeScript creates an executable /bin/sh script in a temp dir and returns its path.
func writeScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-driver")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// ---- ParticipantName ----

func TestParticipantNameReturnsConfigured(t *testing.T) {
	client := Client{Participant: "python"}
	if client.ParticipantName() != "python" {
		t.Fatalf("name=%q", client.ParticipantName())
	}
}

// ---- Version ----

func TestVersionHappyPath(t *testing.T) {
	path := writeScript(t, `printf '%s\n' '{"driver":"fake","participant":"python","subject_version":"1.2.3"}'`)
	client := Client{Path: path, Participant: "python"}
	version, err := client.Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if version.Driver != "fake" || version.Participant != "python" || version.SubjectVersion != "1.2.3" {
		t.Fatalf("version=%+v", version)
	}
}

func TestVersionRejectsParticipantMismatch(t *testing.T) {
	path := writeScript(t, `printf '%s\n' '{"driver":"fake","participant":"rust","subject_version":"1"}'`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Version(context.Background())
	if err == nil || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("expected identity mismatch, got %v", err)
	}
}

func TestVersionRejectsEmptyDriver(t *testing.T) {
	path := writeScript(t, `printf '%s\n' '{"driver":"","participant":"python","subject_version":"1"}'`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Version(context.Background())
	if err == nil || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("expected identity mismatch for empty driver, got %v", err)
	}
}

func TestVersionRejectsEmptySubjectVersion(t *testing.T) {
	path := writeScript(t, `printf '%s\n' '{"driver":"fake","participant":"python","subject_version":""}'`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Version(context.Background())
	if err == nil || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("expected identity mismatch for empty subject_version, got %v", err)
	}
}

func TestVersionRejectsNonZeroExit(t *testing.T) {
	path := writeScript(t, `echo "boom" >&2; exit 4`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Version(context.Background())
	if err == nil || !strings.Contains(err.Error(), "version probe failed") {
		t.Fatalf("expected version probe failed, got %v", err)
	}
}

func TestVersionRejectsMalformedJSON(t *testing.T) {
	path := writeScript(t, `printf '%s\n' 'not-json'`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Version(context.Background())
	if err == nil || !strings.Contains(err.Error(), "version response") {
		t.Fatalf("expected version response error, got %v", err)
	}
}

// ---- Run ----

func TestRunHappyPathDecodesObservation(t *testing.T) {
	path := writeScript(t, `printf '%s\n' '{"contract":"workflow-core/v1","case_id":"WF-001","participant":"python","outcome":"ok","outputs":[],"requests":[],"checkpoints":[],"evidence":{}}'`)
	client := Client{Path: path, Participant: "python"}
	obs, err := client.Run(context.Background(), "WF-001")
	if err != nil {
		t.Fatal(err)
	}
	if obs.Contract != "workflow-core/v1" || obs.CaseID != "WF-001" || obs.Outcome != "ok" {
		t.Fatalf("observation=%+v", obs)
	}
	if obs.Participant != "python" {
		t.Fatalf("participant=%q", obs.Participant)
	}
}

func TestRunRejectsNonZeroExit(t *testing.T) {
	path := writeScript(t, `echo "crash" >&2; exit 7`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Run(context.Background(), "WF-001")
	if err == nil || !strings.Contains(err.Error(), "WF-001 failed") {
		t.Fatalf("expected run failure, got %v", err)
	}
}

func TestRunRejectsMalformedJSON(t *testing.T) {
	path := writeScript(t, `printf '%s\n' 'broken'`)
	client := Client{Path: path, Participant: "python"}
	_, err := client.Run(context.Background(), "WF-001")
	if err == nil || !strings.Contains(err.Error(), "response") {
		t.Fatalf("expected response error, got %v", err)
	}
}

// ---- invoke ----

func TestInvokeRejectsEmptyPath(t *testing.T) {
	client := Client{Path: "", Participant: "python"}
	_, _, err := client.invoke(context.Background(), "version")
	if err == nil || !strings.Contains(err.Error(), "driver path is empty") {
		t.Fatalf("expected empty path error, got %v", err)
	}
}

func TestInvokeRejectsDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	client := Client{Path: dir, Participant: "python"}
	_, _, err := client.invoke(context.Background(), "version")
	if err == nil || !strings.Contains(err.Error(), "driver path is a directory") {
		t.Fatalf("expected directory error, got %v", err)
	}
}

func TestInvokeRejectsNonexistentPath(t *testing.T) {
	client := Client{Path: filepath.Join(t.TempDir(), "nope"), Participant: "python"}
	_, _, err := client.invoke(context.Background(), "version")
	if err == nil {
		t.Fatal("expected stat error for nonexistent path")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("expected IsNotExist error, got %v", err)
	}
}

func TestInvokeSetsSandboxedEnvironment(t *testing.T) {
	// The driver must run hermetically: a fresh HOME, TZ=UTC, NO_COLOR=1.
	path := writeScript(t, `printf '{"home":"%s","tz":"%s","no_color":"%s"}\n' "$HOME" "$TZ" "$NO_COLOR"`)
	client := Client{Path: path, Participant: "python"}
	stdout, _, err := client.invoke(context.Background(), "run")
	if err != nil {
		t.Fatal(err)
	}
	var env map[string]string
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("unmarshal %q: %v", stdout, err)
	}
	if env["tz"] != "UTC" {
		t.Fatalf("TZ=%q want UTC", env["tz"])
	}
	if env["no_color"] != "1" {
		t.Fatalf("NO_COLOR=%q want 1", env["no_color"])
	}
	if env["home"] == "" {
		t.Fatal("HOME must be set to a fresh temp dir, got empty")
	}
	if env["home"] == os.Getenv("HOME") {
		t.Fatalf("HOME must not leak the host home, got %q", env["home"])
	}
}

func TestInvokeTrimsOutputWhitespace(t *testing.T) {
	path := writeScript(t, `printf '  value  \n\n'`)
	client := Client{Path: path, Participant: "python"}
	stdout, stderr, err := client.invoke(context.Background(), "run")
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "value" {
		t.Fatalf("stdout=%q want %q", stdout, "value")
	}
	if stderr != "" {
		t.Fatalf("stderr=%q want empty", stderr)
	}
}

// ---- decodeOne ----

func TestDecodeOneAcceptsSingleValue(t *testing.T) {
	var target map[string]any
	if err := decodeOne(`{"a":1}`, &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target["a"] != float64(1) {
		t.Fatalf("target=%+v", target)
	}
}

func TestDecodeOneAcceptsSurroundingWhitespace(t *testing.T) {
	var target map[string]any
	if err := decodeOne("  \n {\"a\":1} \n  ", &target); err != nil {
		t.Fatal(err)
	}
	if target["a"] != float64(1) {
		t.Fatalf("target=%+v", target)
	}
}

func TestDecodeOneRejectsEmptyInput(t *testing.T) {
	var target map[string]any
	if err := decodeOne("", &target); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestDecodeOneRejectsMalformedTrailingData(t *testing.T) {
	var target map[string]any
	err := decodeOne(`{"outcome":"ok"} garbage`, &target)
	if err == nil || !strings.Contains(err.Error(), "invalid trailing data") {
		t.Fatalf("expected invalid trailing data, got %v", err)
	}
}

func TestDecodeOneRejectsSecondJSONValue(t *testing.T) {
	var target map[string]any
	err := decodeOne(`{"outcome":"ok"} {"second":true}`, &target)
	if err == nil || !strings.Contains(err.Error(), "more than one JSON value") {
		t.Fatalf("expected second value error, got %v", err)
	}
}

// ---- limitWriter ----

func TestLimitWriterWritesWithinBudget(t *testing.T) {
	var buf bytes.Buffer
	w := &limitWriter{buffer: &buf, remaining: 10}
	n, err := w.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if buf.String() != "hello" {
		t.Fatalf("buffer=%q", buf.String())
	}
	if w.remaining != 5 {
		t.Fatalf("remaining=%d want 5", w.remaining)
	}
}

func TestLimitWriterTruncatesAndReportsOriginalLength(t *testing.T) {
	var buf bytes.Buffer
	w := &limitWriter{buffer: &buf, remaining: 3}
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The writer must report the full original length so exec never sees a short write,
	// while only the budgeted prefix is actually buffered.
	if n != 5 {
		t.Fatalf("limitWriter must report original length 5, got %d", n)
	}
	if buf.String() != "hel" {
		t.Fatalf("truncated buffer=%q want hel", buf.String())
	}
	if w.remaining != 0 {
		t.Fatalf("remaining=%d want 0", w.remaining)
	}
}

func TestLimitWriterDropsDataAfterExhausted(t *testing.T) {
	var buf bytes.Buffer
	w := &limitWriter{buffer: &buf, remaining: 0}
	n, err := w.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("must still report original length when exhausted: n=%d err=%v", n, err)
	}
	if buf.Len() != 0 {
		t.Fatalf("buffer should stay empty, got %q", buf.String())
	}
}

func TestLimitWriterExactBoundary(t *testing.T) {
	var buf bytes.Buffer
	w := &limitWriter{buffer: &buf, remaining: 5}
	n, _ := w.Write([]byte("hello"))
	if n != 5 || buf.String() != "hello" || w.remaining != 0 {
		t.Fatalf("boundary write n=%d buf=%q remaining=%d", n, buf.String(), w.remaining)
	}
	// A follow-up write after exhaustion drops everything but still reports full length.
	n, _ = w.Write([]byte("world"))
	if n != 5 || buf.String() != "hello" {
		t.Fatalf("post-exhaustion write n=%d buf=%q", n, buf.String())
	}
}

// ---- DefaultPaths ----

func TestDefaultPathsJoinsExpectedLocations(t *testing.T) {
	python, rust := DefaultPaths("/opt/root")
	wantPython := filepath.ToSlash(filepath.Join("/opt/root", "drivers", "python", "agent-framework-py-driver"))
	wantRust := filepath.ToSlash(filepath.Join("/opt/root", ".bin", "agent-framework-rs-driver"))
	if filepath.ToSlash(python) != wantPython {
		t.Fatalf("python=%q want %q", python, wantPython)
	}
	if filepath.ToSlash(rust) != wantRust {
		t.Fatalf("rust=%q want %q", rust, wantRust)
	}
}

// ---- ProbeAll ----

func TestProbeAllCollectsVersions(t *testing.T) {
	pyPath := writeScript(t, `printf '%s\n' '{"driver":"py","participant":"python","subject_version":"1"}'`)
	rsPath := writeScript(t, `printf '%s\n' '{"driver":"rs","participant":"rust","subject_version":"2"}'`)
	clients := []Client{
		{Path: pyPath, Participant: "python"},
		{Path: rsPath, Participant: "rust"},
	}
	versions, err := ProbeAll(clients, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if versions["python"].SubjectVersion != "1" || versions["rust"].SubjectVersion != "2" {
		t.Fatalf("versions=%+v", versions)
	}
}

func TestProbeAllFailsOnFirstError(t *testing.T) {
	bad := writeScript(t, `exit 1`)
	good := writeScript(t, `printf '%s\n' '{"driver":"rs","participant":"rust","subject_version":"2"}'`)
	clients := []Client{
		{Path: bad, Participant: "python"},
		{Path: good, Participant: "rust"},
	}
	versions, err := ProbeAll(clients, 5*time.Second)
	if err == nil {
		t.Fatal("expected error from failing driver")
	}
	if versions != nil {
		t.Fatalf("versions should be nil on error, got %+v", versions)
	}
}

func TestProbeAllEmptyClients(t *testing.T) {
	versions, err := ProbeAll(nil, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 0 {
		t.Fatalf("expected empty map, got %+v", versions)
	}
}
