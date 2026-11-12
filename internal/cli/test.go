package cli

import (
	"fmt"
	"io"
)

// TestOptions holds validated parameters for the `mrtn test` command.
type TestOptions struct {
	Contest    string
	Question   string
	TestNumber int // 0 indicates all tests should run.
	FilePath   string
	Language   Language
}

// ParseTestArgs validates positional arguments for `mrtn test <prova> <questao> [numero-do-teste] <arquivo>`.
func ParseTestArgs(args []string) (TestOptions, error) {
	if len(args) < 3 || len(args) > 4 {
		return TestOptions{}, fmt.Errorf("número inválido de argumentos para 'test' (esperado 3 ou 4, recebido %d)", len(args))
	}

	contest, err := ParseContest(args[0])
	if err != nil {
		return TestOptions{}, err
	}

	question, err := ParseQuestion(args[1])
	if err != nil {
		return TestOptions{}, err
	}

	var testNumber int
	var filePath string
	if len(args) == 3 {
		filePath = args[2]
	} else {
		testNumber, err = ParseTestNumber(args[2])
		if err != nil {
			return TestOptions{}, err
		}
		filePath = args[3]
	}

	lang, err := ValidateSourceFile(filePath)
	if err != nil {
		return TestOptions{}, err
	}

	return TestOptions{
		Contest:    contest,
		Question:   question,
		TestNumber: testNumber,
		FilePath:   filePath,
		Language:   lang,
	}, nil
}

func runTest(args []string, stdout, _ io.Writer) error {
	posArgs, helpRequested, err := parseSubcommandFlags("test", args, printTestUsage)
	if err != nil {
		return err
	}
	if helpRequested {
		printTestUsage(stdout)
		return nil
	}

	opts, err := ParseTestArgs(posArgs)
	if err != nil {
		return &commandError{err: err, usage: printTestUsage}
	}

	scope := "todos os testes"
	if opts.TestNumber > 0 {
		scope = fmt.Sprintf("teste #%d", opts.TestNumber)
	}

	fmt.Fprintf(
		stdout,
		"[stub] mrtn test: prova=%s questao=%s alvo=%s arquivo=%s linguagem=%s (runner será implementado na próxima fase).\n",
		opts.Contest,
		opts.Question,
		scope,
		opts.FilePath,
		opts.Language,
	)

	return nil
}

func printTestUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn test <prova> <questao> [numero-do-teste] <arquivo>

Descrição:
  Executa sua solução contra os casos de teste da questão especificada:
    <prova>             Identificador da prova (ex: 2026-1)
    <questao>           Identificador da questão, case-insensitive (ex: a ou A)
    [numero-do-teste]   Índice numérico do teste a partir de 1 (opcional)
    <arquivo>           Caminho para o arquivo-fonte (atualmente: .py)
`)
}
