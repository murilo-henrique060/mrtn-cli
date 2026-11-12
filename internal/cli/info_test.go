package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCLIOracleFixture(t *testing.T) string {
	t.Helper()
	dataDir := t.TempDir()
	t.Setenv("MRTN_DATA_DIR", dataDir)
	oracleDir := filepath.Join(dataDir, "oracle")

	qADir := filepath.Join(oracleDir, "2026-1", "a")
	if err := os.MkdirAll(filepath.Join(qADir, "tests", "input"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(qADir, "tests", "output"), 0o755); err != nil {
		t.Fatal(err)
	}

	yamlA := `name:
  pt-BR: Alta Frequência
limits:
  memory: 2048
  output: 100
credits:
  authors:
  - Leonardo Conegundes
`
	if err := os.WriteFile(filepath.Join(qADir, "problem.yaml"), []byte(yamlA), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(qADir, "tests", "input", "sample001"), []byte("1 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(qADir, "tests", "output", "sample001"), []byte("3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	largeInput := strings.Repeat("1234567890\n", 100)
	if err := os.WriteFile(filepath.Join(qADir, "tests", "input", "t01_large"), []byte(largeInput), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(qADir, "tests", "output", "t01_large"), []byte("OK\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	return oracleDir
}

func TestInfoCommandExecution(t *testing.T) {
	setupCLIOracleFixture(t)

	t.Run("list contests", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info"}, &stdout, &stderr); code != 0 {
			t.Fatalf("exit=%d stderr=%q", code, stderr.String())
		}
		out := stdout.String()
		if !strings.Contains(out, "2026-1") || !strings.Contains(out, "1 questão") {
			t.Errorf("unexpected stdout:\n%s", out)
		}
	})

	t.Run("list questions in contest", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1"}, &stdout, &stderr); code != 0 {
			t.Fatalf("exit=%d stderr=%q", code, stderr.String())
		}
		out := stdout.String()
		if !strings.Contains(out, "A") || !strings.Contains(out, "Alta Frequência") || !strings.Contains(out, "2 testes") {
			t.Errorf("unexpected stdout:\n%s", out)
		}
	})

	t.Run("list tests in question", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "A"}, &stdout, &stderr); code != 0 {
			t.Fatalf("exit=%d stderr=%q", code, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{"Alta Frequência", "Leonardo Conegundes", "1  sample001", "2  t01_large"} {
			if !strings.Contains(out, want) {
				t.Errorf("stdout missing %q:\n%s", want, out)
			}
		}
	})

	t.Run("show specific test details", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "a", "1"}, &stdout, &stderr); code != 0 {
			t.Fatalf("exit=%d stderr=%q", code, stderr.String())
		}
		out := stdout.String()
		for _, want := range []string{"Teste #1 (sample001)", "1 2", "3"} {
			if !strings.Contains(out, want) {
				t.Errorf("stdout missing %q:\n%s", want, out)
			}
		}
	})

	t.Run("truncates large test preview cleanly", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "a", "2"}, &stdout, &stderr); code != 0 {
			t.Fatalf("exit=%d stderr=%q", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "conteúdo truncado") {
			t.Errorf("expected truncation notice, got:\n%s", stdout.String())
		}
	})
}

func TestInfoCommandExistenceErrors(t *testing.T) {
	t.Run("oracle not synced yet", func(t *testing.T) {
		t.Setenv("MRTN_DATA_DIR", filepath.Join(t.TempDir(), "empty"))
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info"}, &stdout, &stderr); code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		if !strings.Contains(stderr.String(), "mrtn sync") {
			t.Errorf("stderr=%q, want mrtn sync hint", stderr.String())
		}
	})

	t.Run("missing contest", func(t *testing.T) {
		setupCLIOracleFixture(t)
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2025-1"}, &stdout, &stderr); code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		if !strings.Contains(stderr.String(), `prova "2025-1" não encontrada`) {
			t.Errorf("stderr=%q", stderr.String())
		}
	})

	t.Run("missing question", func(t *testing.T) {
		setupCLIOracleFixture(t)
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "z"}, &stdout, &stderr); code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		if !strings.Contains(stderr.String(), `questão "z" não encontrada`) {
			t.Errorf("stderr=%q", stderr.String())
		}
	})

	t.Run("missing problem.yaml", func(t *testing.T) {
		oracleDir := setupCLIOracleFixture(t)
		if err := os.Remove(filepath.Join(oracleDir, "2026-1", "a", "problem.yaml")); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "a"}, &stdout, &stderr); code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		if !strings.Contains(stderr.String(), "problem.yaml") {
			t.Errorf("stderr=%q", stderr.String())
		}
	})

	t.Run("test number out of range", func(t *testing.T) {
		setupCLIOracleFixture(t)
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "a", "99"}, &stdout, &stderr); code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		if !strings.Contains(stderr.String(), "disponíveis: 1 a 2") {
			t.Errorf("stderr=%q", stderr.String())
		}
	})

	t.Run("rejects test name sample001", func(t *testing.T) {
		setupCLIOracleFixture(t)
		var stdout, stderr bytes.Buffer
		if code := Execute([]string{"info", "2026-1", "a", "sample001"}, &stdout, &stderr); code != 1 {
			t.Fatalf("exit=%d, want 1", code)
		}
		if !strings.Contains(stderr.String(), "apenas o índice numérico") {
			t.Errorf("stderr=%q", stderr.String())
		}
	})
}
