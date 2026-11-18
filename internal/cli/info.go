package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
	"github.com/murilo-henrique060/mrtn-cli/internal/statement"
)

const (
	maxPreviewBytes = 4096
	maxPreviewLines = 60
)

// InfoMode identifies the inspection granularity requested by `mrtn info`.
type InfoMode int

const (
	InfoListContests InfoMode = iota
	InfoListQuestions
	InfoListTests
	InfoShowTest
)

// InfoOptions holds validated parameters for the `mrtn info` command.
type InfoOptions struct {
	Mode       InfoMode
	Contest    string
	Question   string
	TestNumber int
	Web        bool
}

// ParseInfoArgs validates positional arguments for `mrtn info [prova] [questao] [numero-do-teste]`.
func ParseInfoArgs(args []string) (InfoOptions, error) {
	if len(args) > 3 {
		return InfoOptions{}, fmt.Errorf("excesso de argumentos para 'info' (esperado de 0 a 3, recebido %d)", len(args))
	}

	if len(args) == 0 {
		return InfoOptions{Mode: InfoListContests}, nil
	}

	contest, err := ParseContest(args[0])
	if err != nil {
		return InfoOptions{}, err
	}
	if len(args) == 1 {
		return InfoOptions{
			Mode:    InfoListQuestions,
			Contest: contest,
		}, nil
	}

	question, err := ParseQuestion(args[1])
	if err != nil {
		return InfoOptions{}, err
	}
	if len(args) == 2 {
		return InfoOptions{
			Mode:     InfoListTests,
			Contest:  contest,
			Question: question,
		}, nil
	}

	testNum, err := ParseTestNumber(args[2])
	if err != nil {
		return InfoOptions{}, err
	}

	return InfoOptions{
		Mode:       InfoShowTest,
		Contest:    contest,
		Question:   question,
		TestNumber: testNum,
	}, nil
}

func extractInfoFlags(args []string) (posArgs []string, web bool, helpRequested bool, err error) {
	for _, a := range args {
		switch a {
		case "-h", "--help":
			return nil, false, true, nil
		case "-w", "--web":
			web = true
		default:
			if strings.HasPrefix(a, "-") {
				return nil, false, false, &commandError{
					err:   fmt.Errorf("flag desconhecida: %s", a),
					usage: printInfoUsage,
				}
			}
			posArgs = append(posArgs, a)
		}
	}
	return posArgs, web, false, nil
}

func runInfo(args []string, stdout, _ io.Writer) error {
	posArgs, web, helpRequested, err := extractInfoFlags(args)
	if err != nil {
		return err
	}
	if helpRequested {
		printInfoUsage(stdout)
		return nil
	}

	opts, err := ParseInfoArgs(posArgs)
	if err != nil {
		return &commandError{err: err, usage: printInfoUsage}
	}
	opts.Web = web

	switch opts.Mode {
	case InfoListContests:
		return showContests(stdout)
	case InfoListQuestions:
		return showQuestions(stdout, opts.Contest)
	case InfoListTests:
		return showTests(stdout, opts.Contest, opts.Question, opts.Web)
	case InfoShowTest:
		return showTestDetail(stdout, opts.Contest, opts.Question, opts.TestNumber)
	}

	return nil
}

func showContests(stdout io.Writer) error {
	contests, err := oracle.ListContests()
	if err != nil {
		return err
	}
	if len(contests) == 0 {
		fmt.Fprintln(stdout, "Nenhuma prova encontrada. Execute 'mrtn sync' para atualizar os casos de teste.")
		return nil
	}

	fmt.Fprintf(stdout, "Provas disponíveis (%d):\n", len(contests))
	for _, c := range contests {
		fmt.Fprintf(stdout, "  %-16s (%s)\n", c.ID, pluralize(c.QuestionCount, "questão", "questões"))
	}
	return nil
}

func showQuestions(stdout io.Writer, contest string) error {
	questions, err := oracle.ListQuestions(contest)
	if err != nil {
		return err
	}
	if len(questions) == 0 {
		fmt.Fprintf(stdout, "Nenhuma questão encontrada na prova %q.\n", contest)
		return nil
	}

	fmt.Fprintf(stdout, "Prova: %s (%s)\n", contest, pluralize(len(questions), "questão", "questões"))
	for _, q := range questions {
		title := q.Meta.Title
		if title == "" {
			title = "(sem título)"
		}
		fmt.Fprintf(stdout, "  %-4s %-32s (%s)\n", strings.ToUpper(q.ID), title, pluralize(q.TestCount, "teste", "testes"))
	}
	return nil
}

func showTests(stdout io.Writer, contest, question string, openWeb bool) error {
	detail, err := oracle.ListTests(contest, question)
	if err != nil {
		return err
	}

	printQuestionHeader(stdout, detail)

	if strings.TrimSpace(detail.Statement) != "" {
		fmt.Fprintln(stdout, "\n--- Enunciado ---")
		fmt.Fprintln(stdout, statement.FormatForTerminal(detail.Statement))
		fmt.Fprintln(stdout)

		if openWeb {
			htmlPath, webErr := statement.OpenInBrowser(detail.Contest, detail.Question, detail.Meta.Title, detail.Statement)
			if webErr != nil {
				fmt.Fprintf(stdout, "Aviso: HTML gerado em %s, mas não foi possível abrir o navegador automaticamente (%v)\n\n", htmlPath, webErr)
			} else {
				fmt.Fprintf(stdout, "Enunciado aberto no navegador com KaTeX (%s)\n\n", htmlPath)
			}
		}
	} else if openWeb {
		fmt.Fprintln(stdout, "Aviso: arquivo 'statement.md' não encontrado para esta questão.")
	}

	if len(detail.Tests) == 0 {
		fmt.Fprintln(stdout, "Nenhum caso de teste encontrado.")
		return nil
	}

	fmt.Fprintf(stdout, "Casos de teste (%d):\n", len(detail.Tests))
	for _, tc := range detail.Tests {
		fmt.Fprintf(
			stdout,
			"  %3d  %-22s (entrada: %s | saída: %s)\n",
			tc.Index,
			tc.Name,
			formatSize(tc.InputSize),
			formatSize(tc.OutputSize),
		)
	}
	return nil
}

func showTestDetail(stdout io.Writer, contest, question string, testNumber int) error {
	detail, tc, err := oracle.GetTest(contest, question, testNumber)
	if err != nil {
		return err
	}

	printQuestionHeader(stdout, detail)
	fmt.Fprintf(stdout, "Teste #%d (%s)\n", tc.Index, tc.Name)
	fmt.Fprintf(stdout, "Tamanho: entrada %s | saída %s\n", formatSize(tc.InputSize), formatSize(tc.OutputSize))

	fmt.Fprintln(stdout, "\n--- Entrada (input) ---")
	writePreview(stdout, tc.Input, tc.InputPath)

	fmt.Fprintln(stdout, "\n--- Saída esperada (output) ---")
	writePreview(stdout, tc.Output, tc.OutputPath)

	return nil
}

func printQuestionHeader(w io.Writer, detail oracle.QuestionDetail) {
	label := strings.ToUpper(detail.Question)
	if detail.Meta.Title != "" {
		fmt.Fprintf(w, "Prova: %s | Questão: %s - %s\n", detail.Contest, label, detail.Meta.Title)
	} else {
		fmt.Fprintf(w, "Prova: %s | Questão: %s\n", detail.Contest, label)
	}

	var metaParts []string
	if detail.Meta.Limits.Time > 0 {
		metaParts = append(metaParts, fmt.Sprintf("tempo=%gs", detail.Meta.Limits.Time))
	}
	if detail.Meta.Limits.Memory > 0 {
		metaParts = append(metaParts, fmt.Sprintf("memória=%d MB", detail.Meta.Limits.Memory))
	}
	if detail.Meta.Limits.Output > 0 {
		metaParts = append(metaParts, fmt.Sprintf("saída=%d MB", detail.Meta.Limits.Output))
	}
	if len(detail.Meta.Authors) > 0 {
		metaParts = append(metaParts, fmt.Sprintf("autores=%s", strings.Join(detail.Meta.Authors, ", ")))
	}
	if len(metaParts) > 0 {
		fmt.Fprintf(w, "Metadados: %s\n", strings.Join(metaParts, " | "))
	}
}

func writePreview(w io.Writer, content, filePath string) {
	if content == "" {
		fmt.Fprintln(w, "(vazio)")
		return
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	truncated := false
	preview := normalized

	if len(lines) > maxPreviewLines {
		preview = strings.Join(lines[:maxPreviewLines], "\n")
		truncated = true
	}
	if len(preview) > maxPreviewBytes {
		preview = preview[:maxPreviewBytes]
		truncated = true
	}

	fmt.Fprint(w, preview)
	if !strings.HasSuffix(preview, "\n") {
		fmt.Fprintln(w)
	}
	if truncated {
		fmt.Fprintf(w, "... [conteúdo truncado; total: %s, %d linhas | arquivo: %s]\n", formatSize(int64(len(content))), len(lines), filePath)
	}
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

func printInfoUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn info [prova] [questao] [numero-do-teste] [--web]

Descrição:
  Inspeciona provas, questões, enunciados e casos de teste sincronizados localmente:
    mrtn info                           Lista todas as provas disponíveis
    mrtn info <prova>                   Lista as questões da prova
    mrtn info <prova> <questao>         Exibe metadados, enunciado (statement.md) e lista os testes da questão
    mrtn info <prova> <questao> --web   Além de exibir no terminal, abre o enunciado com KaTeX no navegador
    mrtn info <prova> <questao> <num>   Exibe input, output e metadados do teste <num>
`)
}
