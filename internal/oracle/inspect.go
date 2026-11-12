package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ContestSummary describes an available contest/edition in the local oracle repository.
type ContestSummary struct {
	ID            string
	QuestionCount int
}

// QuestionSummary describes a question inside a contest.
type QuestionSummary struct {
	ID        string
	Meta      ProblemMeta
	TestCount int
}

// TestCaseSummary describes metadata and file paths for a single test case.
type TestCaseSummary struct {
	Index      int
	Name       string
	InputPath  string
	OutputPath string
	InputSize  int64
	OutputSize int64
}

// QuestionDetail contains full metadata and ordered test cases for a question.
type QuestionDetail struct {
	Contest  string
	Question string
	Meta     ProblemMeta
	Tests    []TestCaseSummary
}

// TestCaseDetail holds a test case summary along with its input and expected output contents.
type TestCaseDetail struct {
	TestCaseSummary
	Input  string
	Output string
}

func ensureOracleDir() (string, error) {
	oracleDir, err := OracleDir()
	if err != nil {
		return "", err
	}

	info, err := os.Stat(oracleDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("repositório de casos de teste não encontrado em %q; execute 'mrtn sync' primeiro", oracleDir)
		}
		return "", fmt.Errorf("não foi possível acessar o diretório %q: %w", oracleDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("o caminho de dados %q não é um diretório válido", oracleDir)
	}

	return oracleDir, nil
}

func ensureContestDir(contest string) (string, error) {
	oracleDir, err := ensureOracleDir()
	if err != nil {
		return "", err
	}

	contestDir := filepath.Join(oracleDir, contest)
	info, err := os.Stat(contestDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("prova %q não encontrada; use 'mrtn info' para listar as provas disponíveis", contest)
		}
		return "", fmt.Errorf("não foi possível acessar a prova %q: %w", contest, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("o caminho da prova %q (%s) não é um diretório", contest, contestDir)
	}

	return contestDir, nil
}

func ensureQuestionDir(contest, question string) (string, error) {
	contestDir, err := ensureContestDir(contest)
	if err != nil {
		return "", err
	}

	questionDir := filepath.Join(contestDir, question)
	info, err := os.Stat(questionDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("questão %q não encontrada na prova %q; use 'mrtn info %s' para listar as questões", question, contest, contest)
		}
		return "", fmt.Errorf("não foi possível acessar a questão %q da prova %q: %w", question, contest, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("o caminho da questão %q (%s) não é um diretório", question, questionDir)
	}

	return questionDir, nil
}

func ensureTestsDirs(questionDir, contest, question string) (inputDir, outputDir string, err error) {
	testsDir := filepath.Join(questionDir, "tests")
	if info, statErr := os.Stat(testsDir); statErr != nil || !info.IsDir() {
		return "", "", fmt.Errorf("diretório de testes não encontrado para %s %s (%s)", contest, question, testsDir)
	}

	inputDir = filepath.Join(testsDir, "input")
	if info, statErr := os.Stat(inputDir); statErr != nil || !info.IsDir() {
		return "", "", fmt.Errorf("diretório de entradas ('tests/input') não encontrado para %s %s (%s)", contest, question, inputDir)
	}

	outputDir = filepath.Join(testsDir, "output")
	if info, statErr := os.Stat(outputDir); statErr != nil || !info.IsDir() {
		return "", "", fmt.Errorf("diretório de saídas ('tests/output') não encontrado para %s %s (%s)", contest, question, outputDir)
	}

	return inputDir, outputDir, nil
}

// ListContests returns all contests stored in the local mrtn-oracle directory.
func ListContests() ([]ContestSummary, error) {
	oracleDir, err := ensureOracleDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(oracleDir)
	if err != nil {
		return nil, fmt.Errorf("não foi possível listar as provas em %q: %w", oracleDir, err)
	}

	var contests []ContestSummary
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !entry.IsDir() {
			continue
		}

		qEntries, err := os.ReadDir(filepath.Join(oracleDir, name))
		if err != nil {
			return nil, fmt.Errorf("não foi possível ler o diretório da prova %q: %w", name, err)
		}

		qCount := 0
		for _, qEntry := range qEntries {
			if !strings.HasPrefix(qEntry.Name(), ".") && qEntry.IsDir() {
				qCount++
			}
		}

		contests = append(contests, ContestSummary{
			ID:            name,
			QuestionCount: qCount,
		})
	}

	sort.Slice(contests, func(i, j int) bool {
		return contests[i].ID < contests[j].ID
	})

	return contests, nil
}

// ListQuestions returns all questions and their metadata for a given contest.
func ListQuestions(contest string) ([]QuestionSummary, error) {
	contestDir, err := ensureContestDir(contest)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(contestDir)
	if err != nil {
		return nil, fmt.Errorf("não foi possível listar as questões da prova %q: %w", contest, err)
	}

	var questions []QuestionSummary
	for _, entry := range entries {
		qID := entry.Name()
		if strings.HasPrefix(qID, ".") || !entry.IsDir() {
			continue
		}

		detail, err := ListTests(contest, qID)
		if err != nil {
			return nil, err
		}

		questions = append(questions, QuestionSummary{
			ID:        qID,
			Meta:      detail.Meta,
			TestCount: len(detail.Tests),
		})
	}

	sort.Slice(questions, func(i, j int) bool {
		return questions[i].ID < questions[j].ID
	})

	return questions, nil
}

// ListTests verifies and returns metadata plus ordered test cases for a contest question.
func ListTests(contest, question string) (QuestionDetail, error) {
	questionDir, err := ensureQuestionDir(contest, question)
	if err != nil {
		return QuestionDetail{}, err
	}

	meta, err := LoadProblemYAML(filepath.Join(questionDir, "problem.yaml"))
	if err != nil {
		return QuestionDetail{}, err
	}

	inputDir, outputDir, err := ensureTestsDirs(questionDir, contest, question)
	if err != nil {
		return QuestionDetail{}, err
	}

	inEntries, err := os.ReadDir(inputDir)
	if err != nil {
		return QuestionDetail{}, fmt.Errorf("não foi possível ler os arquivos de entrada em %q: %w", inputDir, err)
	}

	var names []string
	for _, entry := range inEntries {
		if strings.HasPrefix(entry.Name(), ".") || entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	tests := make([]TestCaseSummary, 0, len(names))
	for i, name := range names {
		inPath := filepath.Join(inputDir, name)
		inInfo, err := os.Stat(inPath)
		if err != nil || !inInfo.Mode().IsRegular() {
			return QuestionDetail{}, fmt.Errorf("arquivo de entrada inválido %q", inPath)
		}

		outPath := filepath.Join(outputDir, name)
		outInfo, err := os.Stat(outPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return QuestionDetail{}, fmt.Errorf("caso de teste inconsistente em %s %s: saída esperada não encontrada para %q", contest, question, name)
			}
			return QuestionDetail{}, fmt.Errorf("não foi possível acessar o arquivo de saída %q: %w", outPath, err)
		}
		if !outInfo.Mode().IsRegular() {
			return QuestionDetail{}, fmt.Errorf("arquivo de saída inválido %q", outPath)
		}

		tests = append(tests, TestCaseSummary{
			Index:      i + 1,
			Name:       name,
			InputPath:  inPath,
			OutputPath: outPath,
			InputSize:  inInfo.Size(),
			OutputSize: outInfo.Size(),
		})
	}

	return QuestionDetail{
		Contest:  contest,
		Question: question,
		Meta:     meta,
		Tests:    tests,
	}, nil
}

// GetTest verifies and loads the 1-based test case and its input/output contents.
func GetTest(contest, question string, testNumber int) (QuestionDetail, TestCaseDetail, error) {
	detail, err := ListTests(contest, question)
	if err != nil {
		return QuestionDetail{}, TestCaseDetail{}, err
	}

	if len(detail.Tests) == 0 {
		return QuestionDetail{}, TestCaseDetail{}, fmt.Errorf("nenhum caso de teste encontrado para %s %s", contest, question)
	}

	if testNumber < 1 || testNumber > len(detail.Tests) {
		return QuestionDetail{}, TestCaseDetail{}, fmt.Errorf(
			"teste #%d não encontrado para %s %s (disponíveis: 1 a %d)",
			testNumber, contest, question, len(detail.Tests),
		)
	}

	summary := detail.Tests[testNumber-1]
	inBytes, err := os.ReadFile(summary.InputPath)
	if err != nil {
		return QuestionDetail{}, TestCaseDetail{}, fmt.Errorf("não foi possível ler a entrada do teste #%d (%s): %w", testNumber, summary.InputPath, err)
	}

	outBytes, err := os.ReadFile(summary.OutputPath)
	if err != nil {
		return QuestionDetail{}, TestCaseDetail{}, fmt.Errorf("não foi possível ler a saída do teste #%d (%s): %w", testNumber, summary.OutputPath, err)
	}

	return detail, TestCaseDetail{
		TestCaseSummary: summary,
		Input:           string(inBytes),
		Output:          string(outBytes),
	}, nil
}
