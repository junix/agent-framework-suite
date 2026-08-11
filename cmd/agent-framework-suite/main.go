package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/junix/agent-framework-suite/internal/driver"
	"github.com/junix/agent-framework-suite/internal/suite"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	root := findRoot()
	defaultPython, defaultRust := driver.DefaultPaths(root)
	global := flag.NewFlagSet("agent-framework-suite", flag.ContinueOnError)
	global.SetOutput(stderr)
	pythonPath := global.String("python-driver", envOr("AGENT_FRAMEWORK_PY_DRIVER", defaultPython), "Python participant driver")
	rustPath := global.String("rust-driver", envOr("AGENT_FRAMEWORK_RS_DRIVER", defaultRust), "Rust participant driver")
	if err := global.Parse(args); err != nil {
		return 2
	}
	remaining := global.Args()
	if len(remaining) == 0 {
		usage(stderr)
		return 2
	}
	clients := []driver.Client{
		{Path: *pythonPath, Participant: "python"},
		{Path: *rustPath, Participant: "rust"},
	}
	switch remaining[0] {
	case "version":
		return versionCommand(remaining[1:], stdout, stderr)
	case "doctor":
		return doctorCommand(remaining[1:], clients, stdout, stderr)
	case "list":
		return listCommand(remaining[1:], stdout, stderr)
	case "run":
		return runCommand(remaining[1:], clients, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", remaining[0])
		usage(stderr)
		return 2
	}
}

func versionCommand(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("version", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write JSON")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return 2
	}
	if *jsonOutput {
		if err := writeJSON(stdout, map[string]string{"suite": "agent-framework-suite", "version": suite.SuiteVersion, "contract": suite.ContractVersion}); err != nil {
			fmt.Fprintf(stderr, "write JSON: %v\n", err)
			return 2
		}
	} else {
		fmt.Fprintf(stdout, "agent-framework-suite %s (%s)\n", suite.SuiteVersion, suite.ContractVersion)
	}
	return 0
}

func doctorCommand(args []string, clients []driver.Client, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write JSON")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return 2
	}
	versions, err := driver.ProbeAll(clients, 30*time.Second)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	if *jsonOutput {
		if err := writeJSON(stdout, map[string]any{"status": "ready", "contract": suite.ContractVersion, "participants": versions, "cases": len(suite.Catalog())}); err != nil {
			fmt.Fprintf(stderr, "write JSON: %v\n", err)
			return 2
		}
	} else {
		fmt.Fprintf(stdout, "ready: contract=%s cases=%d\n", suite.ContractVersion, len(suite.Catalog()))
		for _, participant := range []string{"python", "rust"} {
			version := versions[participant]
			fmt.Fprintf(stdout, "  %s: %s %s\n", participant, version.Driver, version.SubjectVersion)
		}
	}
	return 0
}

type listItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	Participants []string `json:"participants"`
	Timeout      string   `json:"timeout"`
}

func listCommand(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write JSON")
	selectFlag := flags.String("select", "", "case selector")
	tag := flags.String("tag", "", "required tag")
	args = intersperseFlags(args, map[string]bool{"--json": true}, map[string]bool{"--select": true, "--tag": true})
	if err := flags.Parse(args); err != nil {
		return 2
	}
	specs := appendSelector(*selectFlag, flags.Args())
	cases, err := suite.Select(suite.Catalog(), specs, *tag)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	items := make([]listItem, 0, len(cases))
	for _, item := range cases {
		items = append(items, listItem{ID: item.ID, Name: item.Name, Tags: item.Tags, Participants: []string{"python", "rust"}, Timeout: item.Timeout.String()})
	}
	if *jsonOutput {
		if err := writeJSON(stdout, items); err != nil {
			fmt.Fprintf(stderr, "write JSON: %v\n", err)
			return 2
		}
	} else {
		for _, item := range items {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", item.ID, item.Name, strings.Join(item.Tags, ","), item.Timeout)
		}
	}
	return 0
}

func runCommand(args []string, clients []driver.Client, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write report JSON to stdout")
	selectFlag := flags.String("select", "", "case selector")
	tag := flags.String("tag", "", "required tag")
	timeout := flags.Duration("timeout", 30*time.Second, "maximum per-participant case duration")
	stream := flags.Bool("stream", false, "write completed case NDJSON to stderr")
	reportPath := flags.String("report", "", "atomically write JSON report")
	args = intersperseFlags(
		args,
		map[string]bool{"--json": true, "--stream": true},
		map[string]bool{"--select": true, "--tag": true, "--timeout": true, "--report": true},
	)
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *timeout <= 0 {
		fmt.Fprintln(stderr, "timeout must be positive")
		return 2
	}
	specs := appendSelector(*selectFlag, flags.Args())
	cases, err := suite.Select(suite.Catalog(), specs, *tag)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	versions, err := driver.ProbeAll(clients, *timeout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 3
	}
	participantClients := make([]suite.ParticipantClient, len(clients))
	for index := range clients {
		participantClients[index] = clients[index]
	}
	options := suite.RunOptions{Timeout: *timeout}
	if *stream {
		options.Stream = func(result suite.CaseResult) {
			encoded, _ := json.Marshal(result)
			fmt.Fprintln(stderr, string(encoded))
		}
	}
	report, infrastructureOK := suite.Run(cases, participantClients, versions, options)
	if *reportPath != "" {
		if err := writeReport(*reportPath, report); err != nil {
			fmt.Fprintf(stderr, "write report: %v\n", err)
			return 2
		}
	}
	if *jsonOutput {
		if err := writeJSON(stdout, report); err != nil {
			fmt.Fprintf(stderr, "write JSON: %v\n", err)
			return 2
		}
	} else {
		for _, item := range report.Cases {
			fmt.Fprintf(stdout, "%s %s parity=%s\n", strings.ToUpper(item.Status), item.ID, item.Parity)
			for _, proof := range item.Proof {
				fmt.Fprintf(stdout, "  %s\n", proof)
			}
		}
		fmt.Fprintf(stdout, "summary: passed=%d failed=%d total=%d\n", report.Summary.Passed, report.Summary.Failed, report.Summary.Total)
	}
	if !infrastructureOK {
		return 3
	}
	if report.Summary.Failed > 0 {
		return 1
	}
	return 0
}

func appendSelector(flagValue string, positional []string) []string {
	result := append([]string(nil), positional...)
	if flagValue != "" {
		result = append(result, flagValue)
	}
	return result
}

func intersperseFlags(args []string, booleans, valued map[string]bool) []string {
	options := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		name := argument
		if equals := strings.IndexByte(argument, '='); equals >= 0 {
			name = argument[:equals]
		}
		switch {
		case booleans[name]:
			options = append(options, argument)
		case valued[name]:
			options = append(options, argument)
			if name == argument && index+1 < len(args) {
				index++
				options = append(options, args[index])
			}
		case strings.HasPrefix(argument, "-"):
			options = append(options, argument)
		default:
			positionals = append(positionals, argument)
		}
	}
	return append(options, positionals...)
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeReport(path string, report suite.Report) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".agent-framework-report-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func findRoot() string {
	if cwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "CONTRACT.md")); err == nil {
			return cwd
		}
	}
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Dir(filepath.Dir(executable))
		if _, err := os.Stat(filepath.Join(candidate, "CONTRACT.md")); err == nil {
			return candidate
		}
	}
	return "."
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func usage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: agent-framework-suite [--python-driver PATH] [--rust-driver PATH] <doctor|list|run|version> [options]")
}
