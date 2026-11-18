package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubRepo          = "murilo-henrique060/mrtn-cli"
	githubLatestAPIURL  = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	githubLatestWebURL  = "https://github.com/" + githubRepo + "/releases/latest"
	updateHTTPTimeout   = 60 * time.Second
)

// UpdateOptions holds options for `mrtn update`.
type UpdateOptions struct {
	TargetVersion string
	Force         bool
}

func extractUpdateFlags(args []string) (opts UpdateOptions, helpRequested bool, err error) {
	var posArgs []string
	for _, a := range args {
		switch a {
		case "-h", "--help":
			return UpdateOptions{}, true, nil
		case "-f", "--force":
			opts.Force = true
		default:
			if strings.HasPrefix(a, "-") {
				return UpdateOptions{}, false, &commandError{
					err:   fmt.Errorf("flag desconhecida: %s", a),
					usage: printUpdateUsage,
				}
			}
			posArgs = append(posArgs, a)
		}
	}
	if len(posArgs) > 1 {
		return UpdateOptions{}, false, &commandError{
			err:   fmt.Errorf("excesso de argumentos para 'update' (esperado no máximo 1, recebido %d)", len(posArgs)),
			usage: printUpdateUsage,
		}
	}
	if len(posArgs) == 1 {
		v := strings.TrimSpace(posArgs[0])
		if v != "" && !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		opts.TargetVersion = v
	}
	return opts, false, nil
}

func runUpdate(args []string, stdout, _ io.Writer) error {
	opts, helpRequested, err := extractUpdateFlags(args)
	if err != nil {
		return err
	}
	if helpRequested {
		printUpdateUsage(stdout)
		return nil
	}

	osName, archName, err := detectPlatform()
	if err != nil {
		return err
	}

	targetVersion := opts.TargetVersion
	if targetVersion == "" {
		fmt.Fprintln(stdout, "Verificando a versão mais recente do mrtn no GitHub...")
		targetVersion, err = fetchLatestVersion()
		if err != nil {
			return fmt.Errorf("não foi possível consultar a versão mais recente: %w", err)
		}
	}

	currentVersion := resolveVersion()
	if !opts.Force && normalizeVersion(currentVersion) == normalizeVersion(targetVersion) {
		fmt.Fprintf(stdout, "O mrtn já está na versão mais recente (%s).\n", currentVersion)
		return nil
	}

	exePath, err := resolveExecutablePath()
	if err != nil {
		return err
	}

	assetName := buildAssetName(targetVersion, osName, archName)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", githubRepo, targetVersion, assetName)

	fmt.Fprintf(stdout, "Baixando %s (%s -> %s)...\n", assetName, currentVersion, targetVersion)
	tmpFile, err := downloadReleaseBinary(downloadURL)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	if err := replaceExecutable(tmpFile, exePath); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "mrtn atualizado com sucesso para %s em %s!\n", targetVersion, exePath)
	return nil
}

func detectPlatform() (string, string, error) {
	var osName string
	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		osName = runtime.GOOS
	default:
		return "", "", fmt.Errorf("sistema operacional %q não suportado para atualização automática", runtime.GOOS)
	}

	var archName string
	switch runtime.GOARCH {
	case "amd64", "arm64":
		archName = runtime.GOARCH
	default:
		return "", "", fmt.Errorf("arquitetura %q não suportada para atualização automática", runtime.GOARCH)
	}

	return osName, archName, nil
}

func buildAssetName(version, osName, archName string) string {
	name := fmt.Sprintf("mrtn-%s-%s-%s", version, osName, archName)
	if osName == "windows" {
		name += ".exe"
	}
	return name
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, githubLatestAPIURL, nil)
	if err == nil {
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "mrtn-cli")
		resp, doErr := client.Do(req)
		if doErr == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var payload struct {
					TagName string `json:"tag_name"`
				}
				if decErr := json.NewDecoder(resp.Body).Decode(&payload); decErr == nil && payload.TagName != "" {
					return payload.TagName, nil
				}
			}
		}
	}

	// Fallback: inspect redirect location from https://github.com/<repo>/releases/latest
	noRedirectClient := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := noRedirectClient.Get(githubLatestWebURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if idx := strings.LastIndex(loc, "/tag/"); idx != -1 {
		tag := strings.TrimSpace(loc[idx+len("/tag/"):])
		if tag != "" {
			return tag, nil
		}
	}

	return "", errors.New("não foi possível identificar a tag da release mais recente no GitHub")
}

func downloadReleaseBinary(url string) (string, error) {
	client := &http.Client{Timeout: updateHTTPTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "mrtn-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("falha ao baixar %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("falha ao baixar %s (HTTP %d)", url, resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "mrtn-update-*")
	if err != nil {
		return "", fmt.Errorf("não foi possível criar arquivo temporário para download: %w", err)
	}
	tmpPath := tmpFile.Name()

	written, copyErr := io.Copy(tmpFile, resp.Body)
	closeErr := tmpFile.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("erro durante o download do binário: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", closeErr
	}
	if written == 0 {
		_ = os.Remove(tmpPath)
		return "", errors.New("o arquivo baixado está vazio")
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil && runtime.GOOS != "windows" {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("não foi possível definir permissão de execução no binário baixado: %w", err)
	}

	return tmpPath, nil
}

func resolveExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("não foi possível localizar o caminho do executável atual: %w", err)
	}
	if resolved, evalErr := filepath.EvalSymlinks(exePath); evalErr == nil {
		exePath = resolved
	}
	return exePath, nil
}

func replaceExecutable(newBinPath, targetExePath string) error {
	targetDir := filepath.Dir(targetExePath)
	stagedPath := filepath.Join(targetDir, fmt.Sprintf(".%s.new", filepath.Base(targetExePath)))

	if err := copyFileExecutable(newBinPath, stagedPath); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permissão negada ao atualizar %q (tente executar com sudo ou verifique as permissões)", targetExePath)
		}
		return fmt.Errorf("não foi possível preparar o novo binário em %q: %w", stagedPath, err)
	}

	if runtime.GOOS == "windows" {
		oldPath := targetExePath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(targetExePath, oldPath); err != nil {
			_ = os.Remove(stagedPath)
			return fmt.Errorf("não foi possível substituir o executável em uso %q: %w", targetExePath, err)
		}
		if err := os.Rename(stagedPath, targetExePath); err != nil {
			_ = os.Rename(oldPath, targetExePath)
			_ = os.Remove(stagedPath)
			return fmt.Errorf("não foi possível ativar o novo executável em %q: %w", targetExePath, err)
		}
		scheduleWindowsFileDelete(oldPath)
		return nil
	}

	if err := os.Rename(stagedPath, targetExePath); err != nil {
		_ = os.Remove(stagedPath)
		if os.IsPermission(err) {
			return fmt.Errorf("permissão negada ao substituir %q (tente executar com sudo)", targetExePath)
		}
		return fmt.Errorf("não foi possível substituir o executável %q: %w", targetExePath, err)
	}

	return nil
}

func copyFileExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return closeErr
	}
	return nil
}

func scheduleWindowsFileDelete(filePath string) {
	cmdLine := fmt.Sprintf(`ping 127.0.0.1 -n 2 > nul & del /f /q "%s"`, filePath)
	cmd := exec.Command("cmd", "/c", cmdLine)
	_ = cmd.Start()
}

func printUpdateUsage(w io.Writer) {
	fmt.Fprint(w, `Uso:
  mrtn update [versao] [--force]

Descrição:
  Baixa e instala a versão mais recente do mrtn diretamente das releases do GitHub.

Opções:
  [versao]        Versão específica para instalar (opcional, ex: v1.0.2)
  -f, --force     Força a reinstalação mesmo se já estiver na versão mais recente
`)
}
