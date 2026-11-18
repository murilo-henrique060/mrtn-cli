package statement

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	inlineCodeRe = regexp.MustCompile("`([^`\\n]+)`")
	boldRe       = regexp.MustCompile(`(?s)\*\*(.+?)\*\*`)
	italicRe     = regexp.MustCompile(`(?s)\*([^\*\s][^\*]*?)\*`)
)

func colorsEnabled() bool {
	return os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
}

func ansiBold(s string) string {
	if !colorsEnabled() {
		return s
	}
	return wrapLinesANSI(s, "\033[1m", "\033[22m")
}

func ansiItalic(s string) string {
	if !colorsEnabled() {
		return s
	}
	return wrapLinesANSI(s, "\033[3m", "\033[23m")
}

func ansiCode(s string) string {
	if !colorsEnabled() {
		return s
	}
	return "\033[1;36m" + s + "\033[0m"
}

func wrapLinesANSI(s, openSeq, closeSeq string) string {
	if !strings.Contains(s, "\n") {
		return openSeq + s + closeSeq
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = openSeq + line + closeSeq
	}
	return strings.Join(lines, "\n")
}

// FormatForTerminal formats a Markdown problem statement into clean, readable terminal text,
// converting inline and block LaTeX math into Unicode characters and styling bold, italic, and code spans.
func FormatForTerminal(markdown string) string {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")

	// Split by fenced code blocks first so we don't apply inline markdown/math inside ```...```
	rawLines := strings.Split(normalized, "\n")
	var segments []struct {
		isCodeBlock bool
		text        string
	}

	var buf []string
	inCodeBlock := false
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if len(buf) > 0 {
				segments = append(segments, struct {
					isCodeBlock bool
					text        string
				}{isCodeBlock: inCodeBlock, text: strings.Join(buf, "\n")})
				buf = nil
			}
			inCodeBlock = !inCodeBlock
			continue
		}
		buf = append(buf, line)
	}
	if len(buf) > 0 {
		segments = append(segments, struct {
			isCodeBlock bool
			text        string
		}{isCodeBlock: inCodeBlock, text: strings.Join(buf, "\n")})
	}

	var out []string
	for _, seg := range segments {
		if seg.isCodeBlock {
			out = append(out, "  ┌────────────────────────────────────────")
			for _, line := range strings.Split(seg.text, "\n") {
				out = append(out, "  │ "+ansiCode(line))
			}
			out = append(out, "  └────────────────────────────────────────")
			continue
		}

		formattedText := formatProseBlock(seg.text)
		for _, line := range strings.Split(formattedText, "\n") {
			trimmed := strings.TrimSpace(line)

			// Horizontal rule
			if trimmed == "---" || trimmed == "***" || trimmed == "___" {
				out = append(out, "  "+strings.Repeat("─", 56))
				continue
			}

			// Headings
			if strings.HasPrefix(trimmed, "#") {
				level := 0
				for level < len(trimmed) && trimmed[level] == '#' {
					level++
				}
				headingText := strings.TrimSpace(trimmed[level:])
				if level == 1 {
					out = append(out, fmt.Sprintf("\n  === %s ===", ansiBold(strings.ToUpper(headingText))))
				} else if level == 2 {
					out = append(out, fmt.Sprintf("\n  ── %s ──", ansiBold(headingText)))
				} else {
					out = append(out, fmt.Sprintf("\n  • %s:", ansiBold(headingText)))
				}
				continue
			}

			// Bullet lists (preserving sub-list indentation)
			leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
				item := trimmed[2:]
				if leadingSpaces >= 2 {
					out = append(out, "    ◦ "+item)
				} else {
					out = append(out, "  • "+item)
				}
				continue
			}

			if trimmed == "" {
				out = append(out, "")
				continue
			}

			out = append(out, "  "+line)
		}
	}

	result := strings.Join(out, "\n")
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.Trim(result, "\n")
}

func formatProseBlock(text string) string {
	// 1. Extract inline code spans `...` so their contents (like `*` or `$`) are untouched by math/italic/bold
	var codeSpans []string
	s := inlineCodeRe.ReplaceAllStringFunc(text, func(m string) string {
		sub := inlineCodeRe.FindStringSubmatch(m)
		inner := m[1 : len(m)-1]
		if len(sub) == 2 {
			inner = sub[1]
		}
		idx := len(codeSpans)
		codeSpans = append(codeSpans, ansiCode(inner))
		return fmt.Sprintf("\x00CODE_%d\x00", idx)
	})

	// 2. Render LaTeX math ($...$ and $$...$$)
	s = RenderMathInText(s)

	// 3. Render Bold (**...**)
	s = boldRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[2 : len(m)-2]
		return ansiBold(inner)
	})

	// 4. Render Italic (*...*)
	s = italicRe.ReplaceAllStringFunc(s, func(m string) string {
		inner := m[1 : len(m)-1]
		return ansiItalic(inner)
	})

	// 5. Restore inline code spans
	for i, code := range codeSpans {
		placeholder := fmt.Sprintf("\x00CODE_%d\x00", i)
		s = strings.ReplaceAll(s, placeholder, code)
	}

	return s
}
