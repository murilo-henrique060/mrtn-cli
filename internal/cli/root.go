package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
)

// Version holds the CLI version, injected via -ldflags or resolved from Go build info.
var Version = "v1.0.1"

type commandError struct {
	err   error
	usage func(io.Writer)
}

func (e *commandError) Error() string {
	return e.err.Error()
}

func (e *commandError) Unwrap() error {
	return e.err
}

// Execute runs the mrtn CLI with the provided arguments and output streams, returning an exit code.
func Execute(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stdout)
		return 0
	}

	subcmd := args[0]
	rest := args[1:]

	if subcmd == "-h" || subcmd == "--help" {
		printRootUsage(stdout)
		return 0
	}

	if subcmd == "version" || subcmd == "-v" || subcmd == "--version" {
		fmt.Fprintf(stdout, "mrtn %s\n", resolveVersion())
		return 0
	}

	if subcmd == "help" {
		return runHelp(rest, stdout, stderr)
	}

	var err error
	switch subcmd {
	case "sync":
		err = runSync(rest, stdout, stderr)
	case "info":
		err = runInfo(rest, stdout, stderr)
	case "run":
		err = runRun(rest, stdout, stderr)
	case "clear":
		err = runClear(rest, stdout, stderr)
	case "update":
		err = runUpdate(rest, stdout, stderr)
	case "remove":
		err = runRemove(rest, stdout, stderr)
	default:
		if strings.HasPrefix(subcmd, "-") {
			fmt.Fprintf(stderr, "erro: flag desconhecida %q\n\n", subcmd)
		} else {
			fmt.Fprintf(stderr, "erro: comando desconhecido %q\n\n", subcmd)
		}
		printRootUsage(stderr)
		return 1
	}

	if err != nil {
		fmt.Fprintf(stderr, "erro: %s\n", err)
		var cmdErr *commandError
		if errors.As(err, &cmdErr) && cmdErr.usage != nil {
			fmt.Fprintln(stderr)
			cmdErr.usage(stderr)
		}
		return 1
	}

	return 0
}

func runHelp(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stdout)
		return 0
	}

	switch args[0] {
	case "sync":
		printSyncUsage(stdout)
	case "info":
		printInfoUsage(stdout)
	case "run":
		printRunUsage(stdout)
	case "clear":
		printClearUsage(stdout)
	case "update":
		printUpdateUsage(stdout)
	case "remove":
		printRemoveUsage(stdout)
	default:
		fmt.Fprintf(stderr, "erro: comando desconhecido %q\n\n", args[0])
		printRootUsage(stderr)
		return 1
	}

	return 0
}

func parseSubcommandFlags(name string, args []string, usage func(io.Writer)) (posArgs []string, helpRequested bool, err error) {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return nil, true, nil
		}
	}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	var errBuf bytes.Buffer
	fs.SetOutput(&errBuf)
	// Suppress FlagSet's built-in usage output so callers control stdout vs stderr.
	fs.Usage = func() {}

	if parseErr := fs.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			return nil, true, nil
		}
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = parseErr.Error()
		}
		return nil, false, &commandError{err: errors.New(msg), usage: usage}
	}

	return fs.Args(), false, nil
}

func resolveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}

func printRootUsage(w io.Writer) {
	fmt.Fprint(w, `mrtn - Minimal Runtime Test Node

Uso:
  mrtn <comando> [argumentos]

Comandos disponíveis:
  sync      Sincroniza o repositório de casos de teste (mrtn-oracle)
  info      Consulta provas, questões, enunciados e casos de teste disponíveis
  run       Executa e valida sua solução contra os casos de teste
  clear     Remove os dados locais das provas e questões sincronizadas
  update    Atualiza o mrtn para a versão mais recente
  remove    Desinstala o mrtn e remove todos os seus arquivos locais
  version   Exibe a versão da CLI
  help      Exibe ajuda sobre um comando específico

Use "mrtn <comando> --help" para mais informações sobre cada comando.
`)
}
