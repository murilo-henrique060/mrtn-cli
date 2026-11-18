package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
	"github.com/murilo-henrique060/mrtn-cli/internal/runner"
)

// RunOptions holds validated parameters for the `mrtn run` command.
type RunOptions struct {
	Contest    string
	Question   string
	TestNumber int // 0 indicates all tests should run.
	FilePath   string
	Language   Language
}

// ParseRunArgs validates positional arguments for `mrtn run <prova> <questao> [numero-do-teste] <arquivo>`.
func ParseRunArgs(args []string) (RunOptions, error) {
	if len(args) < 3 || len(args) > 4 {
		return RunOptions{}, fmt.Errorf("número inválido de argumentos para 'run' (esperado 3 ou 4, recebido %d)", len(args))
	}

	contest, err := ParseContest(args[0])
	if err != nil {
		return RunOptions{}, err
	}

	question, err := ParseQuestion(args[1])
	if err != nil {
		return RunOptions{}, err
	}

	var testNumber int
	var filePath string
	if len(args) == 3 {
		filePath = args[2]
	} else {
		testNumber, err = ParseTestNumber(args[2])
		if err != nil {
			return RunOptions{}, err
		}
		filePath = args[3]
	}

	lang, err := ValidateSourceFile(filePath)
	if err != nil {
		return RunOptions{}, err
	}

	return RunOptions{
		Contest:    contest,
		Question:   question,
		TestNumber: testNumber,
		FilePath:   filePath,
		Language:   lang,
	}, nil
}

func runRun(args []string, stdout, _ io.Writer) error {
	posArgs, helpRequested, err := parseSubcommandFlags("run", args, printRunUsage)
	if err != nil {
		return err
	}
	if helpRequested {
		printRunUsage(stdout)
		return nil
	}

	opts, err := ParseRunArgs(posArgs)
	if err != nil {
		return &commandError{err: err, usage: printRunUsage}
	}

	cfg := runner.RunConfig{
		Contest:    opts.Contest,
		Question:   opts.Question,
		TestNumber: opts.TestNumber,
		FilePath:   opts.FilePath,
		Language:   opts.Language,
	}

	detailed := opts.TestNumber > 0
	hooks := runner.RunHooks{
		OnCompileStart: func(filePath string, lang runner.Language, compiler string) {
			fmt.Fprintf(stdout, "Compilando %s (%s via %s)...\n", filePath, lang, compiler)
		},
		OnStart: func(detail oracle.QuestionDetail, runtimeName string, timeLimit time.Duration, total int) {
			printRunHeader(stdout, detail, opts, runtimeName, timeLimit, total)
		},
		OnCaseDone: func(res runner.CaseResult) {
			printCaseResult(stdout, res, opts, detailed)
		},
	}

	report, err := runner.ExecuteTests(context.Background(), cfg, hooks)
	if err != nil {
		return err
	}

	printRunSummary(stdout, opts, report)
	return nil
}

func printRunHeader(
	w io.Writer,
	detail oracle.QuestionDetail,
	opts RunOptions,
	runtimeName string,
	timeLimit time.Duration,
	total int,
) {
	label := strings.ToUpper(detail.Question)
	if detail.Meta.Title != "" {
		fmt.Fprintf(w, "Executando: %s %s - %s\n", detail.Contest, label, detail.Meta.Title)
	} else {
		fmt.Fprintf(w, "Executando: %s %s\n", detail.Contest, label)
	}

	fmt.Fprintf(
		w,
		"Arquivo: %s (%s via %s) | Limite de tempo: %s | Casos: %d\n",
		opts.FilePath,
		opts.Language,
		runtimeName,
		timeLimit,
		total,
	)
	fmt.Fprintln(w, strings.Repeat("-", 60))
}

func printCaseResult(w io.Writer, res runner.CaseResult, opts RunOptions, detailed bool) {
	badge := fmt.Sprintf("[%s]", res.Verdict)
	dur := formatDuration(res.Duration)

	switch res.Verdict {
	case runner.VerdictAC:
		fmt.Fprintf(w, "  %-6s #%-3d %-22s (%s - Aceito)\n", badge, res.Test.Index, res.Test.Name, dur)

	case runner.VerdictWA:
		fmt.Fprintf(w, "  %-6s #%-3d %-22s (%s - Resposta incorreta)\n", badge, res.Test.Index, res.Test.Name, dur)
		if !detailed {
			return
		}
		if res.Input != "" {
			fmt.Fprintln(w, "  Entrada:")
			printIndentedBlock(w, res.Input)
		} else if res.InputSize > 0 {
			fmt.Fprintf(
				w,
				"  Entrada: (grande: %s - use 'mrtn info %s %s %d' para visualizar)\n",
				formatSize(res.InputSize),
				opts.Contest,
				opts.Question,
				res.Test.Index,
			)
		}
		if res.FirstDiff != nil {
			fmt.Fprintf(w, "  Primeira diferença (linha %d):\n", res.FirstDiff.Line)
			fmt.Fprintln(w, res.FirstDiff.Format())
		}
		if res.Stderr != "" {
			fmt.Fprintln(w, "  Prints de debug (stderr):")
			printIndentedBlock(w, res.Stderr)
		}
		if tip := debugTipForLanguage(opts.Language); tip != "" {
			fmt.Fprintf(w, "  %s\n", tip)
		}

	case runner.VerdictTLE:
		fmt.Fprintf(w, "  %-6s #%-3d %-22s (%s - Tempo limite excedido)\n", badge, res.Test.Index, res.Test.Name, dur)
		if !detailed {
			return
		}
		if res.Stderr != "" {
			fmt.Fprintln(w, "  Prints de debug (stderr):")
			printIndentedBlock(w, res.Stderr)
		}
		if res.Stdout != "" {
			fmt.Fprintln(w, "  Prints de debug (stdout):")
			printIndentedBlock(w, res.Stdout)
		}

	case runner.VerdictRTE:
		fmt.Fprintf(w, "  %-6s #%-3d %-22s (%s - Erro de execução, código=%d)\n", badge, res.Test.Index, res.Test.Name, dur, res.ExitCode)
		if !detailed {
			return
		}
		if res.Stderr != "" {
			fmt.Fprintln(w, "  Mensagem de erro (stderr):")
			printIndentedBlock(w, res.Stderr)
		}
		if res.Stdout != "" {
			fmt.Fprintln(w, "  Prints de debug (stdout):")
			printIndentedBlock(w, res.Stdout)
		}
	}
}

func printIndentedBlock(w io.Writer, text string) {
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintf(w, "    | %s\n", line)
	}
}

func debugTipForLanguage(lang Language) string {
	switch lang {
	case runner.LangPython:
		return "Dica: em Python, use print(..., file=sys.stderr) para debugar sem alterar a saída padrão (stdout)."
	case runner.LangC:
		return "Dica: em C, use fprintf(stderr, ...) para debugar sem alterar a saída padrão (stdout)."
	case runner.LangCPP:
		return "Dica: em C++, use std::cerr << ... para debugar sem alterar a saída padrão (stdout)."
	case runner.LangJava:
		return "Dica: em Java, use System.err.println(...) para debugar sem alterar a saída padrão (stdout)."
	case runner.LangKotlin:
		return "Dica: em Kotlin, use System.err.println(...) para debugar sem alterar a saída padrão (stdout)."
	default:
		return ""
	}
}

func printRunSummary(w io.Writer, opts RunOptions, report runner.RunReport) {
	fmt.Fprintln(w, strings.Repeat("-", 60))
	fmt.Fprintf(
		w,
		"Resumo: %d/%d aceitos | AC: %d | WA: %d | TLE: %d | RTE: %d | Tempo total: %s\n",
		report.Passed,
		len(report.Results),
		report.Counts[runner.VerdictAC],
		report.Counts[runner.VerdictWA],
		report.Counts[runner.VerdictTLE],
		report.Counts[runner.VerdictRTE],
		formatDuration(report.TotalTime),
	)
	if opts.TestNumber == 0 && report.Failed > 0 {
		fmt.Fprintf(w, "Dica: execute 'mrtn run %s %s <numero> %s' para ver detalhes da falha de um teste específico.\n", opts.Contest, opts.Question, opts.FilePath)
	}
}

func formatDuration(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func printRunUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn run <prova> <questao> [numero-do-teste] <arquivo>

Descrição:
  Executa sua solução contra os casos de teste da questão especificada:
    <prova>             Identificador da prova (ex: 2026-1)
    <questao>           Identificador da questão, case-insensitive (ex: a ou A)
    [numero-do-teste]   Índice numérico do teste a partir de 1 (opcional)
    <arquivo>           Caminho para o arquivo-fonte (.py, .c, .cpp, .cc, .cxx, .java, .kt)
`)
}
