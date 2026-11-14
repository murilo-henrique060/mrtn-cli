package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
)

const (
	// DefaultTimeLimit is used when problem.yaml does not specify limits.time.
	// In the ICPC problem package standard (problem_format_version: 2025-09), the default time limit is 1s.
	DefaultTimeLimit = 1 * time.Second
	// DefaultCompileLimit is used when problem.yaml does not specify limits.compilation_time.
	DefaultCompileLimit = 60 * time.Second
	// DefaultJavaMemMB is used when problem.yaml does not specify limits.memory.
	DefaultJavaMemMB = 500
	// DefaultJavaStackKB mirrors the judge's explicit stack limit for JVM threads.
	DefaultJavaStackKB = 131072
)

// Language identifies a programming language supported by the runner.
type Language string

const (
	LangPython Language = "Python"
	LangC      Language = "C"
	LangCPP    Language = "C++"
	LangJava   Language = "Java"
	LangKotlin Language = "Kotlin"
)

var javaPublicTypeRe = regexp.MustCompile(`(?m)^[ \t]*public[ \t]+(?:[a-z]+[ \t]+)*(?:class|interface|enum|record)[ \t]+([A-Za-z_$][A-Za-z0-9_$]*)`)

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
	Language          Language
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
	RuntimeName string
	TimeLimit   time.Duration
	Results     []CaseResult
	TotalTime   time.Duration
	Passed      int
	Failed      int
	Counts      map[Verdict]int
}

// RunHooks provides optional progress callbacks during compilation and test execution.
type RunHooks struct {
	OnCompileStart func(filePath string, lang Language, compiler string)
	OnStart        func(detail oracle.QuestionDetail, runtimeName string, timeLimit time.Duration, total int)
	OnCaseDone     func(CaseResult)
}

type preparedProgram struct {
	runtimeName string
	buildCmd    func(ctx context.Context) *exec.Cmd
	cleanup     func()
}

// ResolvePythonInterpreter locates a working Python interpreter in PATH
// prioritizing pypy3, then python3, then python.
func ResolvePythonInterpreter(lookPath func(string) (string, error)) (string, error) {
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	candidates := []string{"pypy3", "python3", "python"}
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

	return "", errors.New("nenhum interpretador Python encontrado no PATH (tentado: pypy3, python3, python)")
}

// ExecuteTests validates oracle files, prepares/compiles the solution if needed, and runs the requested tests.
func ExecuteTests(ctx context.Context, cfg RunConfig, hooks RunHooks) (RunReport, error) {
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

	program, err := prepareProgram(ctx, cfg, detail.Meta.Limits, hooks)
	if err != nil {
		return RunReport{}, err
	}
	defer program.cleanup()

	timeLimit := DefaultTimeLimit
	if cfg.TimeLimitOverride > 0 {
		timeLimit = cfg.TimeLimitOverride
	} else if detail.Meta.Limits.Time > 0 {
		timeLimit = time.Duration(detail.Meta.Limits.Time * float64(time.Second))
	}

	if hooks.OnStart != nil {
		hooks.OnStart(detail, program.runtimeName, timeLimit, len(selected))
	}

	report := RunReport{
		Detail:      detail,
		RuntimeName: program.runtimeName,
		TimeLimit:   timeLimit,
		Results:     make([]CaseResult, 0, len(selected)),
		Counts:      make(map[Verdict]int),
	}

	for _, tc := range selected {
		res, runErr := runSingleCase(ctx, program.buildCmd, tc, timeLimit)
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

func prepareProgram(
	ctx context.Context,
	cfg RunConfig,
	limits oracle.ProblemLimits,
	hooks RunHooks,
) (preparedProgram, error) {
	lookPath := cfg.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	switch cfg.Language {
	case LangPython:
		return preparePythonProgram(ctx, cfg, limits.CompilationTime, lookPath, hooks)

	case LangC:
		return compileNativeProgram(ctx, cfg, limits.CompilationTime, lookPath, hooks, "gcc", []string{"-std=gnu11", "-O2"}, nativeTrailingFlags())

	case LangCPP:
		return compileNativeProgram(ctx, cfg, limits.CompilationTime, lookPath, hooks, "g++", []string{"-std=c++20", "-O2"}, nativeTrailingFlags())

	case LangJava:
		return prepareJavaProgram(ctx, cfg, limits, lookPath, hooks)

	case LangKotlin:
		return prepareKotlinProgram(ctx, cfg, limits, lookPath, hooks)

	default:
		return preparedProgram{}, fmt.Errorf("linguagem não suportada pelo runner: %s", cfg.Language)
	}
}

func nativeTrailingFlags() []string {
	if runtime.GOOS == "darwin" {
		return []string{"-lm"}
	}
	return []string{"-lm", "-static"}
}

func preparePythonProgram(
	ctx context.Context,
	cfg RunConfig,
	compileLimitSec int,
	lookPath func(string) (string, error),
	hooks RunHooks,
) (preparedProgram, error) {
	interpreter, err := ResolvePythonInterpreter(lookPath)
	if err != nil {
		return preparedProgram{}, err
	}
	runtimeName := filepath.Base(interpreter)

	// Copy the script into an isolated temporary directory so `py_compile` never leaves `__pycache__` in the workspace.
	buildDir, err := os.MkdirTemp("", "mrtn-py-*")
	if err != nil {
		return preparedProgram{}, fmt.Errorf("não foi possível criar diretório temporário de compilação: %w", err)
	}
	defer os.RemoveAll(buildDir)

	baseName := filepath.Base(cfg.FilePath)
	if _, err := copySourceToDir(cfg.FilePath, buildDir, baseName); err != nil {
		return preparedProgram{}, err
	}

	if hooks.OnCompileStart != nil {
		hooks.OnCompileStart(cfg.FilePath, LangPython, runtimeName)
	}

	if err := runCompileCommand(ctx, LangPython, compileLimitSec, buildDir, nil, interpreter, "-m", "py_compile", baseName); err != nil {
		return preparedProgram{}, err
	}

	return preparedProgram{
		runtimeName: runtimeName,
		buildCmd: func(caseCtx context.Context) *exec.Cmd {
			return exec.CommandContext(caseCtx, interpreter, "-B", cfg.FilePath)
		},
		cleanup: func() {},
	}, nil
}

func prepareJavaProgram(
	ctx context.Context,
	cfg RunConfig,
	limits oracle.ProblemLimits,
	lookPath func(string) (string, error),
	hooks RunHooks,
) (preparedProgram, error) {
	javacPath, err := lookPath("javac")
	if err != nil {
		return preparedProgram{}, errors.New("compilador Java (javac) não encontrado no PATH; instale um JDK para executar arquivos .java")
	}
	jarPath, err := lookPath("jar")
	if err != nil {
		return preparedProgram{}, errors.New("ferramenta Java (jar) não encontrada no PATH; instale um JDK para executar arquivos .java")
	}
	javaPath, err := lookPath("java")
	if err != nil {
		return preparedProgram{}, errors.New("runtime Java (java) não encontrado no PATH; instale um JDK/JRE para executar arquivos .java")
	}

	buildDir, err := os.MkdirTemp("", "mrtn-java-*")
	if err != nil {
		return preparedProgram{}, fmt.Errorf("não foi possível criar diretório temporário de compilação: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(buildDir)
	}

	srcBytes, err := os.ReadFile(cfg.FilePath)
	if err != nil {
		cleanup()
		return preparedProgram{}, fmt.Errorf("não foi possível ler o arquivo %q: %w", cfg.FilePath, err)
	}

	baseName := filepath.Base(cfg.FilePath)
	mainClass := detectJavaPublicType(string(srcBytes))
	targetFile := baseName
	if mainClass != "" {
		targetFile = mainClass + ".java"
	} else {
		mainClass = strings.TrimSuffix(baseName, filepath.Ext(baseName))
		if mainClass == "" {
			mainClass = "Main"
		}
	}

	if err := os.WriteFile(filepath.Join(buildDir, targetFile), srcBytes, 0o600); err != nil {
		cleanup()
		return preparedProgram{}, fmt.Errorf("não foi possível preparar cópia temporária de %q: %w", cfg.FilePath, err)
	}

	if hooks.OnCompileStart != nil {
		hooks.OnCompileStart(cfg.FilePath, LangJava, "javac")
	}

	javacEnv := []string{"_JAVA_OPTIONS=-Xmx700M -Xms64M"}
	if err := runCompileCommand(ctx, LangJava, limits.CompilationTime, buildDir, javacEnv, javacPath, targetFile); err != nil {
		cleanup()
		return preparedProgram{}, err
	}

	entries, err := os.ReadDir(buildDir)
	if err != nil {
		cleanup()
		return preparedProgram{}, fmt.Errorf("não foi possível listar os arquivos compilados em %q: %w", buildDir, err)
	}

	var classFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".class") {
			classFiles = append(classFiles, entry.Name())
		}
	}
	if len(classFiles) == 0 {
		cleanup()
		return preparedProgram{}, errors.New("erro de compilação (Java): nenhum arquivo .class foi gerado")
	}

	manifest := fmt.Sprintf("Main-Class: %s\n", mainClass)
	if err := os.WriteFile(filepath.Join(buildDir, "Manifest.txt"), []byte(manifest), 0o600); err != nil {
		cleanup()
		return preparedProgram{}, fmt.Errorf("não foi possível gerar Manifest.txt: %w", err)
	}

	jarArgs := make([]string, 0, 3+len(classFiles))
	jarArgs = append(jarArgs, "cfm", "prog.jar", "Manifest.txt")
	jarArgs = append(jarArgs, classFiles...)
	if err := runCompileCommand(ctx, LangJava, limits.CompilationTime, buildDir, nil, jarPath, jarArgs...); err != nil {
		cleanup()
		return preparedProgram{}, err
	}

	return buildJarProgram(javaPath, filepath.Join(buildDir, "prog.jar"), limits.Memory, cleanup), nil
}

func prepareKotlinProgram(
	ctx context.Context,
	cfg RunConfig,
	limits oracle.ProblemLimits,
	lookPath func(string) (string, error),
	hooks RunHooks,
) (preparedProgram, error) {
	kotlincPath, err := lookPath("kotlinc")
	if err != nil {
		return preparedProgram{}, errors.New("compilador Kotlin (kotlinc) não encontrado no PATH; instale o Kotlin para executar arquivos .kt")
	}
	javaPath, err := lookPath("java")
	if err != nil {
		return preparedProgram{}, errors.New("runtime Java (java) não encontrado no PATH; instale um JDK/JRE para executar arquivos .kt")
	}

	buildDir, err := os.MkdirTemp("", "mrtn-kt-*")
	if err != nil {
		return preparedProgram{}, fmt.Errorf("não foi possível criar diretório temporário de compilação: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(buildDir)
	}

	baseName := filepath.Base(cfg.FilePath)
	if _, err := copySourceToDir(cfg.FilePath, buildDir, baseName); err != nil {
		cleanup()
		return preparedProgram{}, err
	}

	if hooks.OnCompileStart != nil {
		hooks.OnCompileStart(cfg.FilePath, LangKotlin, "kotlinc")
	}

	kotlincEnv := []string{"JAVA_OPTS=-Xmx700M -Xms64M"}
	if err := runCompileCommand(ctx, LangKotlin, limits.CompilationTime, buildDir, kotlincEnv, kotlincPath, baseName, "-include-runtime", "-d", "prog.jar"); err != nil {
		cleanup()
		return preparedProgram{}, err
	}

	return buildJarProgram(javaPath, filepath.Join(buildDir, "prog.jar"), limits.Memory, cleanup), nil
}

func buildJarProgram(javaPath, jarFile string, memoryLimitMB int, cleanup func()) preparedProgram {
	memMB := DefaultJavaMemMB
	if memoryLimitMB > 0 {
		memMB = memoryLimitMB
	}

	return preparedProgram{
		runtimeName: "java",
		buildCmd: func(caseCtx context.Context) *exec.Cmd {
			return exec.CommandContext(
				caseCtx,
				javaPath,
				"-Xms10m",
				fmt.Sprintf("-Xmx%dm", memMB),
				fmt.Sprintf("-Xss%dk", DefaultJavaStackKB),
				"-jar",
				jarFile,
			)
		},
		cleanup: cleanup,
	}
}

func detectJavaPublicType(src string) string {
	matches := javaPublicTypeRe.FindStringSubmatch(src)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func copySourceToDir(srcPath, destDir, targetName string) ([]byte, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, fmt.Errorf("não foi possível ler o arquivo %q: %w", srcPath, err)
	}
	if err := os.WriteFile(filepath.Join(destDir, targetName), data, 0o600); err != nil {
		return nil, fmt.Errorf("não foi possível preparar cópia temporária de %q: %w", srcPath, err)
	}
	return data, nil
}

func compileNativeProgram(
	ctx context.Context,
	cfg RunConfig,
	compileLimitSec int,
	lookPath func(string) (string, error),
	hooks RunHooks,
	compilerName string,
	frontFlags []string,
	trailingFlags []string,
) (preparedProgram, error) {
	compilerPath, err := lookPath(compilerName)
	if err != nil {
		return preparedProgram{}, fmt.Errorf(
			"compilador %s (%s) não encontrado no PATH; instale o %s para executar arquivos %s",
			cfg.Language, compilerName, strings.ToUpper(compilerName), filepath.Ext(cfg.FilePath),
		)
	}

	buildDir, err := os.MkdirTemp("", "mrtn-build-*")
	if err != nil {
		return preparedProgram{}, fmt.Errorf("não foi possível criar diretório temporário de compilação: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(buildDir)
	}

	binName := "solution"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(buildDir, binName)

	if hooks.OnCompileStart != nil {
		hooks.OnCompileStart(cfg.FilePath, cfg.Language, compilerName)
	}

	args := make([]string, 0, len(frontFlags)+len(trailingFlags)+3)
	args = append(args, frontFlags...)
	args = append(args, cfg.FilePath, "-o", binPath)
	args = append(args, trailingFlags...)

	if err := runCompileCommand(ctx, cfg.Language, compileLimitSec, "", nil, compilerPath, args...); err != nil {
		cleanup()
		return preparedProgram{}, err
	}

	return preparedProgram{
		runtimeName: compilerName,
		buildCmd: func(caseCtx context.Context) *exec.Cmd {
			return exec.CommandContext(caseCtx, binPath)
		},
		cleanup: cleanup,
	}, nil
}

func runCompileCommand(
	ctx context.Context,
	lang Language,
	compileLimitSec int,
	dir string,
	extraEnv []string,
	bin string,
	args ...string,
) error {
	compileTimeout := DefaultCompileLimit
	if compileLimitSec > 0 {
		compileTimeout = time.Duration(compileLimitSec) * time.Second
	}
	compileCtx, cancel := context.WithTimeout(ctx, compileTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(compileCtx, bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = 500 * time.Millisecond

	if err := cmd.Run(); err != nil {
		if errors.Is(compileCtx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("tempo limite de compilação excedido (%s)", compileTimeout)
		}
		errMsg := formatStderrSnippet(stderr.String())
		if errMsg == "" {
			errMsg = formatStderrSnippet(stdout.String())
		}
		if errMsg != "" {
			return fmt.Errorf("erro de compilação (%s):\n%s", lang, errMsg)
		}
		return fmt.Errorf("erro de compilação (%s): %w", lang, err)
	}

	return nil
}

func runSingleCase(
	ctx context.Context,
	buildCmd func(context.Context) *exec.Cmd,
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
	cmd := buildCmd(caseCtx)
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
