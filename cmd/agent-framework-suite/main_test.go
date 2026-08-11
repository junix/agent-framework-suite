package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/junix/agent-framework-suite/internal/suite"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("broken output")
}

// ---- run dispatch ----

func TestRunNoArgsReturnsUsageAndCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage") {
		t.Fatalf("expected usage on stderr, got %q", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("no-args should not write stdout, got %q", stdout.String())
	}
}

func TestRunUnknownCommandReturnsCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"frobnicate"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d want 2", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "frobnicate"`) {
		t.Fatalf("expected unknown command, got %q", stderr.String())
	}
}

// ---- version ----

func TestVersionJSONShape(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"version", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{suite.SuiteVersion, suite.ContractVersion, `"suite"`, `"version"`, `"contract"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("version json missing %q: %s", want, out)
		}
	}
}

func TestVersionTextContainsVersionAndContract(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	line := strings.TrimSpace(stdout.String())
	if !strings.Contains(line, suite.SuiteVersion) || !strings.Contains(line, suite.ContractVersion) {
		t.Fatalf("text version line=%q", line)
	}
}

func TestVersionRejectsExtraPositional(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"version", "extra"}, &stdout, &stderr); code != 2 {
		t.Fatalf("expected exit 2 for extra positional, got %d", code)
	}
}

func TestJSONWriteFailureProducesNonzeroExit(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"version", "--json"}, failingWriter{}, &stderr); code == 0 {
		t.Fatal("expected nonzero exit code")
	}
	if !strings.Contains(stderr.String(), "broken output") {
		t.Fatalf("stderr did not report write failure: %s", stderr.String())
	}
}

// ---- list ----

func TestListJSONOutputsAllCases(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"list", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	// Boundary cases of the catalog must both appear.
	for _, id := range []string{"WF-001", "WF-010"} {
		if !strings.Contains(out, id) {
			t.Fatalf("list output missing %s: %s", id, out)
		}
	}
}

func TestListTextIsTSV(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"list", "WF-001"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	line := strings.TrimSpace(stdout.String())
	fields := strings.Split(line, "\t")
	// id \t name \t tags \t timeout
	if len(fields) < 4 {
		t.Fatalf("expected TSV with >=4 columns, got %q", line)
	}
	if fields[0] != "WF-001" {
		t.Fatalf("first field=%q want WF-001", fields[0])
	}
	if fields[1] != "chain_happy" {
		t.Fatalf("name field=%q want chain_happy", fields[1])
	}
}

func TestListBadSelectorReturnsCode2(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"list", "--select", "missing"}, &stdout, &stderr); code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "matched no cases") {
		t.Fatalf("stderr=%s", stderr.String())
	}
}

func TestListWithTagFilters(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"list", "--tag", "checkpoint", "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	// WF-007 and WF-008 carry the checkpoint tag; WF-001 does not.
	if !strings.Contains(out, "WF-007") || !strings.Contains(out, "WF-008") {
		t.Fatalf("expected checkpoint cases: %s", out)
	}
	if strings.Contains(out, "WF-001") {
		t.Fatalf("WF-001 is not a checkpoint case: %s", out)
	}
}

func TestVersionAndListDoNotNeedParticipants(t *testing.T) {
	for _, args := range [][]string{{"version", "--json"}, {"list", "WF-001", "--json"}, {"list", "--select", "WF-001", "--json"}} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 0 {
			t.Fatalf("run(%v) code=%d stderr=%s", args, code, stderr.String())
		}
		if stdout.Len() == 0 {
			t.Fatalf("run(%v) produced no output", args)
		}
	}
}

// ---- run command (validation paths that don't need live drivers) ----

func TestRunCommandRejectsNonPositiveTimeout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"run", "--timeout", "0s"}, &stdout, &stderr); code != 2 {
		t.Fatalf("expected exit 2 for zero timeout, got %d", code)
	}
	if !strings.Contains(stderr.String(), "timeout must be positive") {
		t.Fatalf("stderr=%s", stderr.String())
	}
}

func TestRunCommandRejectsBadSelectorBeforeProbing(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"run", "--select", "missing"}, &stdout, &stderr); code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "matched no cases") {
		t.Fatalf("stderr=%s", stderr.String())
	}
}

// ---- appendSelector ----

func TestAppendSelectorWithFlagValue(t *testing.T) {
	got := appendSelector("WF-001", []string{"pos"})
	want := []string{"pos", "WF-001"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAppendSelectorWithoutFlagValue(t *testing.T) {
	got := appendSelector("", []string{"pos"})
	want := []string{"pos"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestAppendSelectorEmpty(t *testing.T) {
	got := appendSelector("", nil)
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

// ---- intersperseFlags ----

func TestIntersperseFlagsKeepsFlagValuesAndPositionals(t *testing.T) {
	got := intersperseFlags(
		[]string{"WF-001", "--report", "out.json", "--json"},
		map[string]bool{"--json": true},
		map[string]bool{"--report": true},
	)
	want := []string{"--report", "out.json", "--json", "WF-001"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestIntersperseFlagsHandlesEqualsForm(t *testing.T) {
	// "--report=out.json" is a valued flag in = form; it stays in options, no extra arg consumed.
	got := intersperseFlags(
		[]string{"WF-001", "--report=out.json"},
		map[string]bool{},
		map[string]bool{"--report": true},
	)
	want := []string{"--report=out.json", "WF-001"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestIntersperseFlagsValuedAsLastArg(t *testing.T) {
	// A valued flag as the last arg with no following value is kept as-is in options.
	got := intersperseFlags(
		[]string{"WF-001", "--report"},
		map[string]bool{},
		map[string]bool{"--report": true},
	)
	want := []string{"--report", "WF-001"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestIntersperseFlagsUnknownDashArgGoesToOptions(t *testing.T) {
	got := intersperseFlags(
		[]string{"WF-001", "--unknown", "val"},
		map[string]bool{},
		map[string]bool{},
	)
	// "--unknown" is a dash arg -> options; "val" and "WF-001" are positionals.
	want := []string{"--unknown", "WF-001", "val"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

// ---- writeReport ----

func TestWriteReportIsAtomicAndReadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "report.json")
	report := suite.Report{
		SuiteVersion: suite.SuiteVersion,
		Contract:     suite.ContractVersion,
		Summary:      suite.Summary{Total: 3, Passed: 2, Failed: 1},
	}
	if err := writeReport(path, report); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"total": 3`) || !strings.Contains(string(data), suite.SuiteVersion) {
		t.Fatalf("report content wrong: %s", data)
	}
	// Atomicity: no temp leftovers in the directory tree.
	matches, _ := filepath.Glob(filepath.Join(dir, "nested", ".agent-framework-report-*.tmp"))
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}

// ---- envOr ----

func TestEnvOrFallback(t *testing.T) {
	const key = "AGENT_FRAMEWORK_TEST_ENVOR"
	t.Setenv(key, "")
	if got := envOr(key, "fallback"); got != "fallback" {
		t.Fatalf("got %q want fallback", got)
	}
	t.Setenv(key, "real")
	if got := envOr(key, "fallback"); got != "real" {
		t.Fatalf("got %q want real", got)
	}
}

// ---- findRoot ----

func TestFindRootUsesCwdWithContract(t *testing.T) {
	// findRoot's first branch resolves to cwd when CONTRACT.md sits in it.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CONTRACT.md"), []byte("# contract\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	got := findRoot()
	if _, statErr := os.Stat(filepath.Join(got, "CONTRACT.md")); statErr != nil {
		t.Fatalf("findRoot=%q does not contain CONTRACT.md: %v", got, statErr)
	}
}

// ---- usage ----

func TestUsageListsSubcommands(t *testing.T) {
	var stderr bytes.Buffer
	usage(&stderr)
	out := stderr.String()
	for _, sub := range []string{"doctor", "list", "run", "version"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("usage missing %q: %s", sub, out)
		}
	}
}
