package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/murilo-henrique060/mrtn-cli/internal/runner"
)

// Language identifies the programming language detected from a solution file.
type Language = runner.Language

const (
	LangPython = runner.LangPython
	LangC      = runner.LangC
	LangCPP    = runner.LangCPP
	LangJava   = runner.LangJava
	LangKotlin = runner.LangKotlin
)

const supportedExtensionsLabel = "Python (.py), C (.c), C++ (.cpp, .cc, .cxx), Java (.java), Kotlin (.kt)"

var supportedLanguages = map[string]Language{
	".py":   LangPython,
	".c":    LangC,
	".cpp":  LangCPP,
	".cc":   LangCPP,
	".cxx":  LangCPP,
	".java": LangJava,
	".kt":   LangKotlin,
}

// ParseContest normalizes and validates a contest identifier (e.g., "2026-1", "2026-regional").
// Existence of the directory on disk is verified in a later execution phase.
func ParseContest(raw string) (string, error) {
	contest := strings.ToLower(strings.TrimSpace(raw))
	if contest == "" {
		return "", errors.New("o identificador da prova não pode ser vazio")
	}

	for i, r := range contest {
		isAlnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		isSep := r == '-' || r == '_'
		if i == 0 && !isAlnum {
			return "", fmt.Errorf("prova inválida %q: deve começar com letra ou número", raw)
		}
		if !isAlnum && !isSep {
			return "", fmt.Errorf("prova inválida %q: use apenas letras, números, '-' ou '_'", raw)
		}
	}

	return contest, nil
}

// ParseQuestion normalizes and validates a question identifier containing only ASCII letters.
// Existence of the question directory on disk is verified in a later execution phase.
func ParseQuestion(raw string) (string, error) {
	question := strings.ToLower(strings.TrimSpace(raw))
	if question == "" {
		return "", errors.New("a questão não pode ser vazia")
	}

	for _, r := range question {
		if r < 'a' || r > 'z' {
			return "", fmt.Errorf("questão inválida %q: informe apenas letras (ex: a, A)", raw)
		}
	}

	return question, nil
}

// ParseTestNumber parses a 1-based test index matching the ordering shown by `mrtn info`.
func ParseTestNumber(raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	n, err := strconv.Atoi(trimmed)
	if err != nil || n < 1 {
		return 0, fmt.Errorf(
			"número do teste inválido %q: apenas o índice numérico (>= 1, na ordem exibida pelo 'mrtn info') é válido (ex: 1 ao invés de sample001)",
			raw,
		)
	}

	return n, nil
}

// ValidateSourceFile verifies that path points to a regular file (not a directory) with a supported extension.
func ValidateSourceFile(path string) (Language, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", errors.New("o caminho do arquivo não pode ser vazio")
	}

	info, err := os.Stat(trimmed)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("arquivo não encontrado: %s", trimmed)
		}
		return "", fmt.Errorf("não foi possível acessar o arquivo %q: %w", trimmed, err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("o caminho %q é um diretório; informe um arquivo de código-fonte", trimmed)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("o caminho %q não é um arquivo regular válido", trimmed)
	}

	ext := strings.ToLower(filepath.Ext(trimmed))
	if lang, ok := supportedLanguages[ext]; ok {
		return lang, nil
	}

	if ext == "" {
		return "", fmt.Errorf("arquivo %q não possui extensão; linguagens suportadas atualmente: %s", trimmed, supportedExtensionsLabel)
	}

	return "", fmt.Errorf("extensão de arquivo não suportada %q; linguagens suportadas atualmente: %s", ext, supportedExtensionsLabel)
}
