package oracle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// SyncAction indicates whether Sync performed an initial clone or an update pull.
type SyncAction string

const (
	ActionCloned  SyncAction = "cloned"
	ActionUpdated SyncAction = "updated"
)

// SyncConfig configures how Sync clones or updates the mrtn-oracle repository.
type SyncConfig struct {
	RepoURL   string
	TargetDir string
	LookPath  func(file string) (string, error)
}

// SyncResult describes the outcome of a successful Sync operation.
type SyncResult struct {
	Action    SyncAction
	TargetDir string
}

// Sync verifies that git is installed and clones or updates the mrtn-oracle repository
// in the designated user data directory.
func Sync(ctx context.Context, cfg SyncConfig, stdout, stderr io.Writer) (SyncResult, error) {
	lookPath := cfg.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	// Verify git availability before touching any files or directories.
	gitBin, err := lookPath("git")
	if err != nil {
		return SyncResult{}, fmt.Errorf("git não encontrado no sistema: instale o Git (https://git-scm.com/) para usar o 'mrtn sync': %w", err)
	}

	repoURL := cfg.RepoURL
	if repoURL == "" {
		repoURL = DefaultRepoURL
	}

	targetDir := cfg.TargetDir
	if targetDir == "" {
		targetDir, err = OracleDir()
		if err != nil {
			return SyncResult{}, err
		}
	}

	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	action, err := determineSyncAction(targetDir)
	if err != nil {
		return SyncResult{}, err
	}

	if action == ActionCloned {
		if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
			return SyncResult{}, fmt.Errorf("não foi possível criar o diretório de dados %q: %w", filepath.Dir(targetDir), err)
		}

		fmt.Fprintf(stdout, "Clonando casos de teste de %s em %s...\n", repoURL, targetDir)
		if err := runGit(ctx, gitBin, stdout, stderr, "clone", repoURL, targetDir); err != nil {
			return SyncResult{}, fmt.Errorf("falha ao clonar o repositório %s: %w", repoURL, err)
		}

		return SyncResult{Action: ActionCloned, TargetDir: targetDir}, nil
	}

	fmt.Fprintf(stdout, "Atualizando casos de teste em %s...\n", targetDir)
	if err := runGit(ctx, gitBin, stdout, stderr, "-C", targetDir, "pull", "--ff-only"); err != nil {
		return SyncResult{}, fmt.Errorf("falha ao atualizar o repositório em %q: %w", targetDir, err)
	}

	return SyncResult{Action: ActionUpdated, TargetDir: targetDir}, nil
}

func determineSyncAction(targetDir string) (SyncAction, error) {
	info, err := os.Stat(targetDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ActionCloned, nil
		}
		return "", fmt.Errorf("não foi possível verificar o diretório de dados %q: %w", targetDir, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("o caminho de dados %q já existe, mas não é um diretório", targetDir)
	}

	if _, err := os.Stat(filepath.Join(targetDir, ".git")); err == nil {
		return ActionUpdated, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("não foi possível verificar o repositório Git em %q: %w", targetDir, err)
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return "", fmt.Errorf("não foi possível ler o diretório %q: %w", targetDir, err)
	}
	if len(entries) == 0 {
		return ActionCloned, nil
	}

	return "", fmt.Errorf(
		"o diretório %q já existe e não é um repositório Git válido; remova-o manualmente e execute 'mrtn sync' novamente",
		targetDir,
	)
}

func runGit(ctx context.Context, gitBin string, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, gitBin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	// Prevent git from hanging on interactive credential prompts in headless/CLI runs.
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	return cmd.Run()
}
