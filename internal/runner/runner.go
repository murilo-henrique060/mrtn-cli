package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
)

// DefaultTimeLimit is used when problem.yaml does not specify limits.time.
const DefaultTimeLimit = 5 * time.Second

// Verdict represents the evaluation outcome of a single test case.
type Verdict string

const (
	VerdictAC  Verdict = "AC"
	VerdictWA  Verdict = "WA"
	VerdictTLE Verdict = "TLE"
	VerdictRTE Verdict = "RTE"
)

// RunConfig configures test execution against a local oracle question.
type RunConfig struct {
	Contest           string
	Question          string
	TestNumber        int // 0 runs all tests; >= 1 runs only that 1-based index.
	FilePath          string
	TimeLimitOverride time.Duration
	LookPath          func(string) (string, error)
}

// CaseResult holds the execution result of a single test case.
type CaseResult struct {
	Test     oracle.TestCaseSummary
	Verdict  Verdict
	Duration time.Duration
	ExitCode int
	Stderr   string
	Diff     string
}

// RunReport summarizes all executed test cases.
type RunReport struct {
	Detail      oracle.QuestionDetail
	Interpreter string
	TimeLimit   time.Duration
	Results     []CaseResult
	TotalTime   time.Duration
	Passed      int
	Failed      int
	Counts      map[Verdict]int
}

// ResolvePythonInterpreter locates a working Python interpreter in PATH
// following the README priority: python3, pypy3, python.
func ResolvePythonInterpreter(lookPath func(string) (string, error)) (string, error) {
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	candidates := []string{"python3", "pypy3", "python"}
	for _, candidate := range candidates {
		path, err := lookPath(candidate)
		if err != nil {
			continue
		}
		// Verify the binary actually runs (guards against Windows Store python3 stubs).
		checkCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = exec.CommandContext(checkCtx, path, "--version").Run()
		cancel()
		if err == nil {
			return path, nil
		}
	}

	return "", errors.New("nenhum interpretador Python encontrado no PATH (tentado: python3, pypy3, python)")
}

// RunHooks provides optional progress callbacks during test execution.
type RunHooks struct {
	OnStart    func(detail oracle.QuestionDetail, interpreter string, timeLimit time.Duration, total int)
	OnCaseDone func(CaseResult)
}

// ExecutePythonTests validates oracle files, resolves the Python interpreter, and runs the requested tests.
func ExecutePythonTests(ctx context.Context, cfg RunConfig, hooks RunHooks) (RunReport, error) {
	detail, err := oracle.ListTests(cfg.Contest, cfg.Question)
	if err != nil {
		return RunReport{}, err
	}
	if len(detail.Tests) == 0 {
		return RunReport{}, fmt.Errorf("nenhum caso de teste encontrado para %s %s", cfg.Contest, cfg.Question)
	}

	selected := detail.Tests
	if cfg.TestNumber > 0 {
		if cfg.TestNumber > len(detail.Tests) {
			return RunReport{}, fmt.Errorf(
				"teste #%d não encontrado para %s %s (disponíveis: 1 a %d)",
				cfg.TestNumber, cfg.Contest, cfg.Question, len(detail.Tests),
			)
		}
		selected = detail.Tests[cfg.TestNumber-1 : cfg.TestNumber]
	}

	interpreter, err := ResolvePythonInterpreter(cfg.LookPath)
	if err != nil {
		return RunReport{}, err
	}

	timeLimit := DefaultTimeLimit
	if cfg.TimeLimitOverride > 0 {
		timeLimit = cfg.TimeLimitOverride
	} else if detail.Meta.Limits.Time > 0 {
		timeLimit = time.Duration(detail.Meta.Limits.Time) * time.Second
	}

	if hooks.OnStart != nil {
		hooks.OnStart(detail, interpreter, timeLimit, len(selected))
	}

	report := RunReport{
		Detail:      detail,
		Interpreter: interpreter,
		TimeLimit:   timeLimit,
		Results:     make([]CaseResult, 0, len(selected)),
		Counts:      make(map[Verdict]int),
	}

	for _, tc := range selected {
		res, runErr := runSinglePythonCase(ctx, interpreter, cfg.FilePath, tc, timeLimit)
		if runErr != nil {
			return report, runErr
		}

		report.Results = append(report.Results, res)
		report.TotalTime += res.Duration
		report.Counts[res.Verdict]++
		if res.Verdict == VerdictAC {
			report.Passed++
		} else {
			report.Failed++
		}

		if hooks.OnCaseDone != nil {
			hooks.OnCaseDone(res)
		}
	}

	return report, nil
}

func runSinglePythonCase(
	ctx context.Context,
	interpreter string,
	scriptPath string,
	tc oracle.TestCaseSummary,
	timeLimit time.Duration,
) (CaseResult, error) {
	inFile, err := os.Open(tc.InputPath)
	if err != nil {
		return CaseResult{}, fmt.Errorf("não foi possível abrir a entrada do teste #%d (%s): %w", tc.Index, tc.InputPath, err)
	}
	defer inFile.Close()

	expectedBytes, err := os.ReadFile(tc.OutputPath)
	if err != nil {
		return CaseResult{}, fmt.Errorf("não foi possível ler a saída esperada do teste #%d (%s): %w", tc.Index, tc.OutputPath, err)
	}

	caseCtx, cancel := context.WithTimeout(ctx, timeLimit)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(caseCtx, interpreter, scriptPath)
	cmd.Stdin = inFile
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = 500 * time.Millisecond

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start)

	if errors.Is(caseCtx.Err(), context.DeadlineExceeded) {
		return CaseResult{
			Test:     tc,
			Verdict:  VerdictTLE,
			Duration: elapsed,
		}, nil
	}

	if errors.Is(caseCtx.Err(), context.Canceled) {
		return CaseResult{}, caseCtx.Err()
	}

	if runErr != nil {
		exitCode := -1
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return CaseResult{
			Test:     tc,
			Verdict:  VerdictRTE,
			Duration: elapsed,
			ExitCode: exitCode,
			Stderr:   formatStderrSnippet(stderr.String()),
		}, nil
	}

	equal, diff := CompareOutputs(string(expectedBytes), stdout.String())
	if !equal {
		return CaseResult{
			Test:     tc,
			Verdict:  VerdictWA,
			Duration: elapsed,
			Diff:     diff,
		}, nil
	}

	return CaseResult{
		Test:     tc,
		Verdict:  VerdictAC,
		Duration: elapsed,
	}, nil
}

func formatStderrSnippet(s string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	if trimmed == "" {
		return ""
	}
	const maxLines = 15
	lines := strings.Split(trimmed, "\n")
	if len(lines) > maxLines {
		lines = append(lines[:maxLines], fmt.Sprintf("... (+%d linha(s))", len(lines)-maxLines))
	}
	return strings.Join(lines, "\n")
}
