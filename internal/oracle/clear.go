package oracle

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ClearResult describes the outcome of removing locally installed oracle data.
type ClearResult struct {
	TargetDir string
	Removed   bool
}

// Clear removes the locally synced mrtn-oracle directory and, if empty, its parent data directory.
func Clear() (ClearResult, error) {
	oracleDir, err := OracleDir()
	if err != nil {
		return ClearResult{}, err
	}

	if _, err := os.Stat(oracleDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ClearResult{TargetDir: oracleDir, Removed: false}, nil
		}
		return ClearResult{}, fmt.Errorf("não foi possível acessar o diretório de dados %q: %w", oracleDir, err)
	}

	if err := removeAllWritable(oracleDir); err != nil {
		return ClearResult{}, fmt.Errorf("não foi possível remover os dados em %q: %w", oracleDir, err)
	}

	// Clean up parent data directory only if it has no remaining entries.
	_ = os.Remove(filepath.Dir(oracleDir))

	return ClearResult{TargetDir: oracleDir, Removed: true}, nil
}

func removeAllWritable(path string) error {
	if err := os.RemoveAll(path); err == nil {
		return nil
	}

	// Git pack files may be marked read-only (especially on Windows); clear read-only bits before retrying.
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			_ = os.Chmod(p, 0o700)
		} else {
			_ = os.Chmod(p, 0o600)
		}
		return nil
	})

	return os.RemoveAll(path)
}
