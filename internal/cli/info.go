package cli

import (
	"fmt"
	"io"
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

func runInfo(args []string, stdout, _ io.Writer) error {
	posArgs, helpRequested, err := parseSubcommandFlags("info", args, printInfoUsage)
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

	switch opts.Mode {
	case InfoListContests:
		fmt.Fprintln(stdout, "[stub] mrtn info: listar provas disponíveis (implementação na próxima fase).")
	case InfoListQuestions:
		fmt.Fprintf(stdout, "[stub] mrtn info: listar questões da prova %q (implementação na próxima fase).\n", opts.Contest)
	case InfoListTests:
		fmt.Fprintf(stdout, "[stub] mrtn info: listar testes de %s-%s (implementação na próxima fase).\n", opts.Contest, opts.Question)
	case InfoShowTest:
		fmt.Fprintf(stdout, "[stub] mrtn info: exibir detalhes do teste #%d de %s-%s (implementação na próxima fase).\n", opts.TestNumber, opts.Contest, opts.Question)
	}

	return nil
}

func printInfoUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn info [prova] [questao] [numero-do-teste]

Descrição:
  Inspeciona provas, questões e casos de teste sincronizados localmente:
    mrtn info                           Lista todas as provas disponíveis
    mrtn info <prova>                   Lista as questões da prova
    mrtn info <prova> <questao>         Lista os testes da questão
    mrtn info <prova> <questao> <num>   Exibe input, output e metadados do teste <num>
`)
}
