package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
)

// ClearOptions holds validated parameters for the `mrtn clear` command.
type ClearOptions struct{}

// ParseClearArgs validates positional arguments for `mrtn clear`.
func ParseClearArgs(args []string) (ClearOptions, error) {
	if len(args) > 0 {
		return ClearOptions{}, fmt.Errorf("o comando 'clear' não aceita argumentos (recebido: %s)", strings.Join(args, " "))
	}
	return ClearOptions{}, nil
}

func runClear(args []string, stdout, _ io.Writer) error {
	posArgs, helpRequested, err := parseSubcommandFlags("clear", args, printClearUsage)
	if err != nil {
		return err
	}
	if helpRequested {
		printClearUsage(stdout)
		return nil
	}

	if _, err := ParseClearArgs(posArgs); err != nil {
		return &commandError{err: err, usage: printClearUsage}
	}

	res, err := oracle.Clear()
	if err != nil {
		return err
	}

	if !res.Removed {
		fmt.Fprintf(stdout, "Nenhum dado encontrado para remover em: %s\n", res.TargetDir)
		return nil
	}

	fmt.Fprintf(stdout, "Dados das questões removidos com sucesso de: %s\n", res.TargetDir)
	return nil
}

func printClearUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn clear

Descrição:
  Remove os dados locais das provas e questões sincronizadas via 'mrtn sync'.
`)
}
