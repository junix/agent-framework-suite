package driver

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const validOutput = `{"driver":"fake","participant":"python","subject_version":"版本🙂","outcome":"ok"}`

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func outputFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "output")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return shellQuote(path)
}

func TestProtocolOutputBudget(t *testing.T) {
	padded := validOutput + strings.Repeat(" ", streamLimit-len(validOutput))
	for _, test := range []struct {
		name, output, wantError string
	}{
		{"below limit", padded[:streamLimit-1], ""},
		{"exact limit", padded, ""},
		{"one extra whitespace byte", padded + " ", "stdout exceeds"},
		{"hidden second document", padded + `{}`, "stdout exceeds"},
		{"oversized single document", `{"padding":"` + strings.Repeat("x", streamLimit) + `"}`, "stdout exceeds"},
		{"visible second document", validOutput + `{}`, "more than one JSON value"},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := Client{Path: writeScript(t, "cat "+outputFile(t, test.output)), Participant: "python"}
			for _, operation := range []string{"version", "run"} {
				t.Run(operation, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					var err error
					if operation == "version" {
						_, err = client.Version(ctx)
					} else {
						_, err = client.Run(ctx, "WF-001")
					}
					if test.wantError == "" {
						if err != nil {
							t.Fatal(err)
						}
					} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
						t.Fatalf("want error containing %q, got %v", test.wantError, err)
					}
				})
			}
		})
	}
}

func TestStderrBudgetIsIndependent(t *testing.T) {
	for _, size := range []int{streamLimit - 1, streamLimit, streamLimit + 1} {
		client := Client{Path: writeScript(t, "cat "+outputFile(t, validOutput)+"\ncat "+outputFile(t, strings.Repeat("e", size))+" >&2"), Participant: "python"}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		stdout, stderr, err := client.invoke(ctx, "run", "WF-001")
		cancel()
		if err != nil || stdout != validOutput {
			t.Fatalf("stderr size %d affected response: stdout=%q err=%v", size, stdout, err)
		}
		want := strings.Repeat("e", min(size, streamLimit))
		if size > streamLimit {
			want += "\n[stderr truncated after 65536 bytes]"
		}
		if stderr != want {
			t.Fatalf("stderr size %d: got %d bytes, want retained prefix and marker (%d bytes)", size, len(stderr), len(want))
		}
	}
}

func TestOverflowDrainsBothStreamsToCompletion(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "completed")
	output := outputFile(t, strings.Repeat("x", 16*streamLimit))
	client := Client{Path: writeScript(t, "cat "+output+" &\na=$!\ncat "+output+" >&2 &\nb=$!\nwait \"$a\" || exit 61\nwait \"$b\" || exit 62\nprintf done > "+shellQuote(marker)), Participant: "python"}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, stderr, err := client.invoke(ctx, "run")
	if err == nil || !strings.Contains(err.Error(), "stdout exceeds") {
		t.Fatalf("want stdout budget error, got %v", err)
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) || ctx.Err() != nil {
		t.Fatalf("driver did not finish draining normally: %v; context=%v", err, ctx.Err())
	}
	if got, readErr := os.ReadFile(marker); readErr != nil || string(got) != "done" {
		t.Fatalf("driver did not reach completion: marker=%q error=%v", got, readErr)
	}
	if len(stderr) > streamLimit+64 || !strings.HasSuffix(stderr, "[stderr truncated after 65536 bytes]") {
		t.Fatalf("stderr diagnostic is not bounded and marked: length=%d", len(stderr))
	}
}

func TestOverflowPreservesExitError(t *testing.T) {
	output := outputFile(t, strings.Repeat("x", streamLimit+1))
	client := Client{Path: writeScript(t, "cat "+output+"\ncat "+output+" >&2\nexit 7")}
	_, stderr, err := client.invoke(context.Background(), "run")
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 7 {
		t.Fatalf("want retained exit code 7, got %v", err)
	}
	if !strings.Contains(err.Error(), "stdout exceeds") || !strings.HasSuffix(stderr, "[stderr truncated after 65536 bytes]") {
		t.Fatalf("missing overflow diagnostics: error=%v stderr length=%d", err, len(stderr))
	}
}

func TestCancellationReturnsWithOverflowedStreams(t *testing.T) {
	ready := filepath.Join(t.TempDir(), "ready")
	// Only shell builtins after launch: cancelling the command cannot leave a
	// descendant holding either pipe open.
	client := Client{Path: writeScript(t, "printf '%65537s' x\nprintf '%65537s' e >&2\nprintf ready > "+shellQuote(ready)+"\nwhile :; do printf x; printf e >&2; done")}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type result struct {
		stderr string
		err    error
	}
	done := make(chan result, 1)
	go func() {
		_, stderr, err := client.invoke(ctx, "run")
		done <- result{stderr, err}
	}()
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if ctx.Err() != nil {
			t.Fatal("driver did not emit both overflowing streams before deadline")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case got := <-done:
		var exitError *exec.ExitError
		if !errors.As(got.err, &exitError) || !strings.Contains(got.err.Error(), "stdout exceeds") {
			t.Fatalf("cancellation lost command or overflow error: %v", got.err)
		}
		if len(got.stderr) > streamLimit+64 || !strings.HasSuffix(got.stderr, "[stderr truncated after 65536 bytes]") {
			t.Fatalf("missing bounded stderr diagnostic after cancellation: length=%d", len(got.stderr))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled driver/copy goroutines did not return")
	}
}
