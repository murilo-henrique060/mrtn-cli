package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/murilo-henrique060/mrtn-cli/internal/oracle"
)

func runRemove(args []string, stdout, _ io.Writer) error {
	posArgs, helpRequested, err := parseSubcommandFlags("remove", args, printRemoveUsage)
	if err != nil {
		return err
	}
	if helpRequested {
		printRemoveUsage(stdout)
		return nil
	}

	if len(posArgs) > 0 {
		return &commandError{
			err:   fmt.Errorf("o comando 'remove' não aceita argumentos (recebido: %v)", posArgs),
			usage: printRemoveUsage,
		}
	}

	// 1. Remove synced oracle and application data directory
	res, err := oracle.Clear()
	if err != nil {
		return err
	}
	if res.Removed {
		fmt.Fprintf(stdout, "Dados de provas e questões removidos: %s\n", res.TargetDir)
	}

	if dataDir, dirErr := oracle.DataDir(); dirErr == nil {
		if _, statErr := os.Stat(dataDir); statErr == nil {
			if rmErr := os.RemoveAll(dataDir); rmErr == nil {
				fmt.Fprintf(stdout, "Diretório de dados removido: %s\n", dataDir)
			}
		}
	}

	// 2. Remove temporary HTML statements cache
	stmtTempDir := filepath.Join(os.TempDir(), "mrtn-statements")
	if _, statErr := os.Stat(stmtTempDir); statErr == nil {
		_ = os.RemoveAll(stmtTempDir)
		fmt.Fprintf(stdout, "Cache temporário de enunciados removido: %s\n", stmtTempDir)
	}

	// 3. Remove the mrtn executable binary itself
	exePath, err := resolveExecutablePath()
	if err != nil {
		return err
	}

	_ = os.Remove(exePath + ".old")
	_ = os.Remove(filepath.Join(filepath.Dir(exePath), fmt.Sprintf(".%s.new", filepath.Base(exePath))))

	if runtime.GOOS == "windows" {
		scheduleWindowsUninstall(exePath)
		fmt.Fprintf(stdout, "Executável agendado para remoção: %s\n", exePath)
		fmt.Fprintln(stdout, "mrtn foi desinstalado com sucesso.")
		return nil
	}

	if err := os.Remove(exePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		if os.IsPermission(err) {
			return fmt.Errorf("permissão negada ao remover o executável %q (tente executar 'sudo mrtn remove')", exePath)
		}
		return fmt.Errorf("não foi possível remover o executável %q: %w", exePath, err)
	}

	fmt.Fprintf(stdout, "Executável removido: %s\n", exePath)
	fmt.Fprintln(stdout, "mrtn foi desinstalado com sucesso.")
	return nil
}

func scheduleWindowsUninstall(exePath string) {
	dir := filepath.Dir(exePath)
	var cmdLine string
	if filepath.Base(dir) == "mrtn" {
		cmdLine = fmt.Sprintf(`ping 127.0.0.1 -n 2 > nul & del /f /q "%s" & rmdir "%s" 2>nul`, exePath, dir)
	} else {
		cmdLine = fmt.Sprintf(`ping 127.0.0.1 -n 2 > nul & del /f /q "%s"`, exePath)
	}
	cmd := exec.Command("cmd", "/c", cmdLine)
	_ = cmd.Start()
}

func printRemoveUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn remove

Descrição:
  Desinstala o mrtn completamente do sistema, removendo todos os casos de teste
  sincronizados, diretórios de dados, caches temporários e o próprio executável.
`)
}
