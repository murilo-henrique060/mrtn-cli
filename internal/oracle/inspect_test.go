package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupOracleFixture(t *testing.T) string {
	t.Helper()
	dataDir := t.TempDir()
	t.Setenv("MRTN_DATA_DIR", dataDir)
	oracleDir := filepath.Join(dataDir, "oracle")

	// Add non-contest entries that must be ignored (.git, LICENSE).
	if err := os.MkdirAll(filepath.Join(oracleDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oracleDir, "LICENSE"), []byte("MIT\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create 2026-1/a with 3 tests (including hidden ._ file that must be ignored).
	qADir := filepath.Join(oracleDir, "2026-1", "a")
	if err := os.MkdirAll(filepath.Join(qADir, "tests", "input"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(qADir, "tests", "output"), 0o755); err != nil {
		t.Fatal(err)
	}
	yamlA := `problem_format_version: 2025-09
name:
  pt-BR: Alta Frequência
uuid: c801963f-ba37-5bcf-83b4-88cd0e12b6c7
type: pass-fail
limits:
  memory: 2048
  output: 100
credits:
  authors:
  - Leonardo Conegundes
source:
  name: mdp-br-fp-2026
license: unknown
`
	if err := os.WriteFile(filepath.Join(qADir, "problem.yaml"), []byte(yamlA), 0o644); err != nil {
		t.Fatal(err)
	}
	testPairs := map[string][2]string{
		"t01_manual_001": {"10\n", "100\n"},
		"sample001":      {"1 2\n", "3\n"},
		"sample002":      {"4 5\n", "9\n"},
	}
	for name, pair := range testPairs {
		if err := os.WriteFile(filepath.Join(qADir, "tests", "input", name), []byte(pair[0]), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(qADir, "tests", "output", name), []byte(pair[1]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(qADir, "tests", "input", "._sample001"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create 2026-1/b with 1 test.
	qBDir := filepath.Join(oracleDir, "2026-1", "b")
	if err := os.MkdirAll(filepath.Join(qBDir, "tests", "input"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(qBDir, "tests", "output"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(qBDir, "problem.yaml"), []byte("name:\n  pt-BR: Sequências\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(qBDir, "tests", "input", "sample001"), []byte("5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(qBDir, "tests", "output", "sample001"), []byte("25\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	return oracleDir
}

func TestParseProblemYAML(t *testing.T) {
	t.Run("parses localized name, limits and authors", func(t *testing.T) {
		raw := "problem_format_version: 2025-09\r\n" +
			"name:\r\n" +
			"  en: High Frequency\r\n" +
			"  pt-BR: \"Alta Frequência\"\r\n" +
			"limits:\r\n" +
			"  time: 2\r\n" +
			"  memory: 2048\r\n" +
			"  output: 100\r\n" +
			"credits:\r\n" +
			"  authors:\r\n" +
			"  - Leonardo Conegundes\r\n" +
			"  - Outro Autor\r\n" +
			"source:\r\n" +
			"  name: mdp-br-fp-2026\r\n"

		meta, err := ParseProblemYAML([]byte(raw))
		if err != nil {
			t.Fatalf("ParseProblemYAML error: %v", err)
		}
		if meta.Title != "Alta Frequência" {
			t.Errorf("Title = %q, want Alta Frequência", meta.Title)
		}
		if meta.Limits.Time != 2 || meta.Limits.Memory != 2048 || meta.Limits.Output != 100 {
			t.Errorf("Limits = %+v, unexpected", meta.Limits)
		}
		if len(meta.Authors) != 2 || meta.Authors[0] != "Leonardo Conegundes" {
			t.Errorf("Authors = %v, unexpected", meta.Authors)
		}
		if meta.Source != "mdp-br-fp-2026" {
			t.Errorf("Source = %q, want mdp-br-fp-2026", meta.Source)
		}
	})

	t.Run("parses scalar name", func(t *testing.T) {
		meta, err := ParseProblemYAML([]byte("name: 'Problema Simples'\n"))
		if err != nil {
			t.Fatalf("ParseProblemYAML error: %v", err)
		}
		if meta.Title != "Problema Simples" {
			t.Errorf("Title = %q, want Problema Simples", meta.Title)
		}
	})
}

func TestInspectHappyPath(t *testing.T) {
	setupOracleFixture(t)

	contests, err := ListContests()
	if err != nil {
		t.Fatalf("ListContests() error: %v", err)
	}
	if len(contests) != 1 || contests[0].ID != "2026-1" || contests[0].QuestionCount != 2 {
		t.Fatalf("ListContests() = %+v, unexpected", contests)
	}

	questions, err := ListQuestions("2026-1")
	if err != nil {
		t.Fatalf("ListQuestions() error: %v", err)
	}
	if len(questions) != 2 {
		t.Fatalf("len(questions) = %d, want 2", len(questions))
	}
	if questions[0].ID != "a" || questions[0].Meta.Title != "Alta Frequência" || questions[0].TestCount != 3 {
		t.Errorf("question[0] = %+v, unexpected", questions[0])
	}
	if questions[1].ID != "b" || questions[1].Meta.Title != "Sequências" || questions[1].TestCount != 1 {
		t.Errorf("question[1] = %+v, unexpected", questions[1])
	}

	detail, err := ListTests("2026-1", "a")
	if err != nil {
		t.Fatalf("ListTests() error: %v", err)
	}
	if len(detail.Tests) != 3 {
		t.Fatalf("len(detail.Tests) = %d, want 3", len(detail.Tests))
	}
	if detail.Tests[0].Index != 1 || detail.Tests[0].Name != "sample001" {
		t.Errorf("first test = %+v, want #1 sample001", detail.Tests[0])
	}
	if detail.Tests[2].Index != 3 || detail.Tests[2].Name != "t01_manual_001" {
		t.Errorf("third test = %+v, want #3 t01_manual_001", detail.Tests[2])
	}

	_, tc, err := GetTest("2026-1", "a", 1)
	if err != nil {
		t.Fatalf("GetTest(1) error: %v", err)
	}
	if tc.Input != "1 2\n" || tc.Output != "3\n" {
		t.Errorf("GetTest(1) = (%q, %q), unexpected", tc.Input, tc.Output)
	}
}

func TestInspectExistenceValidation(t *testing.T) {
	t.Run("missing oracle dir suggests mrtn sync", func(t *testing.T) {
		t.Setenv("MRTN_DATA_DIR", filepath.Join(t.TempDir(), "not-synced"))
		_, err := ListContests()
		if err == nil || !strings.Contains(err.Error(), "mrtn sync") {
			t.Fatalf("expected mrtn sync hint, got: %v", err)
		}
	})

	t.Run("missing contest dir", func(t *testing.T) {
		setupOracleFixture(t)
		_, err := ListQuestions("2025-1")
		if err == nil || !strings.Contains(err.Error(), `prova "2025-1" não encontrada`) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing question dir", func(t *testing.T) {
		setupOracleFixture(t)
		_, err := ListTests("2026-1", "z")
		if err == nil || !strings.Contains(err.Error(), `questão "z" não encontrada`) {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing problem.yaml", func(t *testing.T) {
		oracleDir := setupOracleFixture(t)
		if err := os.Remove(filepath.Join(oracleDir, "2026-1", "a", "problem.yaml")); err != nil {
			t.Fatal(err)
		}
		_, err := ListTests("2026-1", "a")
		if err == nil || !strings.Contains(err.Error(), "problem.yaml") {
			t.Fatalf("expected problem.yaml error, got: %v", err)
		}
	})

	t.Run("missing tests output dir", func(t *testing.T) {
		oracleDir := setupOracleFixture(t)
		if err := os.RemoveAll(filepath.Join(oracleDir, "2026-1", "a", "tests", "output")); err != nil {
			t.Fatal(err)
		}
		_, err := ListTests("2026-1", "a")
		if err == nil || !strings.Contains(err.Error(), "tests/output") {
			t.Fatalf("expected tests/output error, got: %v", err)
		}
	})

	t.Run("missing matching output file", func(t *testing.T) {
		oracleDir := setupOracleFixture(t)
		if err := os.Remove(filepath.Join(oracleDir, "2026-1", "a", "tests", "output", "sample002")); err != nil {
			t.Fatal(err)
		}
		_, err := ListTests("2026-1", "a")
		if err == nil || !strings.Contains(err.Error(), "saída esperada não encontrada") {
			t.Fatalf("expected missing output error, got: %v", err)
		}
	})

	t.Run("test number out of range", func(t *testing.T) {
		setupOracleFixture(t)
		_, _, err := GetTest("2026-1", "a", 99)
		if err == nil || !strings.Contains(err.Error(), "disponíveis: 1 a 3") {
			t.Fatalf("expected out of range error, got: %v", err)
		}
	})
}
