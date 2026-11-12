package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
)

// SyncOptions holds validated parameters for the `mrtn sync` command.
type SyncOptions struct{}

// ParseSyncArgs validates positional arguments for `mrtn sync`.
func ParseSyncArgs(args []string) (SyncOptions, error) {
	if len(args) > 0 {
		return SyncOptions{}, fmt.Errorf("o comando 'sync' não aceita argumentos (recebido: %s)", strings.Join(args, " "))
	}
	return SyncOptions{}, nil
}

func runSync(args []string, stdout, stderr io.Writer) error {
	posArgs, helpRequested, err := parseSubcommandFlags("sync", args, printSyncUsage)
	if err != nil {
		return err
	}
	if helpRequested {
		printSyncUsage(stdout)
		return nil
	}

	if _, err := ParseSyncArgs(posArgs); err != nil {
		return &commandError{err: err, usage: printSyncUsage}
	}

	res, err := oracle.Sync(context.Background(), oracle.SyncConfig{}, stdout, stderr)
	if err != nil {
		return err
	}

	status := "repositório atualizado"
	if res.Action == oracle.ActionCloned {
		status = "repositório clonado"
	}
	fmt.Fprintf(stdout, "Sincronização concluída (%s). Casos de teste disponíveis em: %s\n", status, res.TargetDir)
	return nil
}

func printSyncUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn sync

Descrição:
  Baixa ou atualiza localmente os casos de teste e respostas oficiais
  do repositório mrtn-oracle no diretório de dados do usuário
  (pode ser sobrescrito via variável MRTN_DATA_DIR).
`)
}
