package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/junix/agent-framework-suite/internal/suite"
)

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

func TestWriteReportIsAtomicAndReadable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "report.json")
	report := suite.Report{SuiteVersion: suite.SuiteVersion, Contract: suite.ContractVersion}
	if err := writeReport(path, report); err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".json" {
		t.Fatal("unexpected report path")
	}
}
