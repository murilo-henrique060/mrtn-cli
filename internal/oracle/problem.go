package oracle

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ProblemLimits stores resource limits parsed from problem.yaml.
type ProblemLimits struct {
	Time             int
	Memory           int
	Output           int
	CompilationTime  int
	CompilationMem   int
	ValidationTime   int
	ValidationMem    int
	ValidationOutput int
}

// ProblemMeta stores problem metadata parsed from problem.yaml.
type ProblemMeta struct {
	Title   string
	UUID    string
	Type    string
	Limits  ProblemLimits
	Authors []string
	Source  string
	License string
}

// LoadProblemYAML reads and parses a problem.yaml file from disk.
func LoadProblemYAML(path string) (ProblemMeta, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ProblemMeta{}, fmt.Errorf("arquivo 'problem.yaml' não encontrado (%s): %w", path, err)
	}
	if info.IsDir() {
		return ProblemMeta{}, fmt.Errorf("esperado arquivo em %q, mas é um diretório", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ProblemMeta{}, fmt.Errorf("não foi possível ler %q: %w", path, err)
	}

	return ParseProblemYAML(data)
}

// ParseProblemYAML parses the subset of problem.yaml fields used by mrtn.
func ParseProblemYAML(data []byte) (ProblemMeta, error) {
	var meta ProblemMeta
	localizedNames := make(map[string]string)
	var firstLocalized string

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	section := ""
	subsection := ""
	sectionIndent := 0

	for _, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := countLeadingSpaces(rawLine)
		if indent <= sectionIndent && !strings.HasPrefix(trimmed, "-") {
			section = ""
			subsection = ""
		}

		if strings.HasPrefix(trimmed, "-") && section == "credits" && subsection == "authors" {
			author := unquoteYAML(strings.TrimSpace(strings.TrimPrefix(trimmed, "-")))
			if author != "" {
				meta.Authors = append(meta.Authors, author)
			}
			continue
		}

		key, val, ok := splitYAMLKeyValue(trimmed)
		if !ok {
			continue
		}

		if indent == 0 {
			section = key
			subsection = ""
			sectionIndent = indent
			switch key {
			case "name":
				if val != "" {
					meta.Title = val
				}
			case "uuid":
				meta.UUID = val
			case "type":
				meta.Type = val
			case "license":
				meta.License = val
			}
			continue
		}

		switch section {
		case "name":
			if val != "" {
				localizedNames[key] = val
				if firstLocalized == "" {
					firstLocalized = val
				}
			}
		case "limits":
			assignLimit(&meta.Limits, key, val)
		case "credits":
			if key == "authors" {
				subsection = "authors"
			}
		case "source":
			if key == "name" && val != "" {
				meta.Source = val
			}
		}
	}

	if meta.Title == "" {
		if pt, ok := localizedNames["pt-BR"]; ok {
			meta.Title = pt
		} else if en, ok := localizedNames["en"]; ok {
			meta.Title = en
		} else {
			meta.Title = firstLocalized
		}
	}

	return meta, nil
}

func countLeadingSpaces(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' {
			n++
		} else if r == '\t' {
			n += 2
		} else {
			break
		}
	}
	return n
}

func splitYAMLKeyValue(line string) (key, val string, ok bool) {
	idx := strings.Index(line, ":")
	if idx == -1 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	rawVal := strings.TrimSpace(line[idx+1:])
	val = unquoteYAML(stripInlineComment(rawVal))
	return key, val, true
}

func stripInlineComment(s string) string {
	if s == "" {
		return ""
	}
	if s[0] == '"' || s[0] == '\'' {
		return s
	}
	if idx := strings.Index(s, " #"); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

func unquoteYAML(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func assignLimit(l *ProblemLimits, key, val string) {
	n, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	switch key {
	case "time":
		l.Time = n
	case "memory":
		l.Memory = n
	case "output":
		l.Output = n
	case "compilation_time":
		l.CompilationTime = n
	case "compilation_memory":
		l.CompilationMem = n
	case "validation_time":
		l.ValidationTime = n
	case "validation_memory":
		l.ValidationMem = n
	case "validation_output":
		l.ValidationOutput = n
	}
}
