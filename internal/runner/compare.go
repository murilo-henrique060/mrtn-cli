package runner

import (
	"fmt"
	"strings"
)

const maxDiffLines = 20

// DiffDetail describes a single line difference between expected and actual output.
type DiffDetail struct {
	Line     int
	HasExp   bool
	Expected string
	HasAct   bool
	Actual   string
}

// Format returns a human-readable representation of the line difference.
func (d DiffDetail) Format() string {
	var b strings.Builder
	if d.HasExp {
		fmt.Fprintf(&b, "    - esperado: %q\n", truncateLine(d.Expected))
	} else {
		fmt.Fprintf(&b, "    - esperado: <EOF>\n")
	}
	if d.HasAct {
		fmt.Fprintf(&b, "    + obtido:   %q", truncateLine(d.Actual))
	} else {
		fmt.Fprintf(&b, "    + obtido:   <EOF>")
	}
	return b.String()
}

// NormalizeOutput standardizes line endings (CRLF -> LF) and trims trailing newlines at EOF.
func NormalizeOutput(s string) string {
	normalized := strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimRight(normalized, "\n")
}

// FirstDiff compares expected and actual outputs after line-ending normalization.
// When they differ, it returns false and the details of the very first differing line.
func FirstDiff(expected, actual string) (bool, *DiffDetail) {
	normExp := NormalizeOutput(expected)
	normAct := NormalizeOutput(actual)
	if normExp == normAct {
		return true, nil
	}

	expLines := splitLines(normExp)
	actLines := splitLines(normAct)

	maxLen := len(expLines)
	if len(actLines) > maxLen {
		maxLen = len(actLines)
	}

	for i := 0; i < maxLen; i++ {
		var expLine, actLine string
		hasExp := i < len(expLines)
		hasAct := i < len(actLines)
		if hasExp {
			expLine = expLines[i]
		}
		if hasAct {
			actLine = actLines[i]
		}

		if hasExp && hasAct && expLine == actLine {
			continue
		}

		return false, &DiffDetail{
			Line:     i + 1,
			HasExp:   hasExp,
			Expected: expLine,
			HasAct:   hasAct,
			Actual:   actLine,
		}
	}

	return true, nil
}

// CompareOutputs compares expected and actual outputs after line-ending normalization.
// When they differ, it returns a formatted line-by-line diff bounded to maxDiffLines.
func CompareOutputs(expected, actual string) (bool, string) {
	normExp := NormalizeOutput(expected)
	normAct := NormalizeOutput(actual)
	if normExp == normAct {
		return true, ""
	}

	expLines := splitLines(normExp)
	actLines := splitLines(normAct)

	maxLen := len(expLines)
	if len(actLines) > maxLen {
		maxLen = len(actLines)
	}

	var b strings.Builder
	diffsShown := 0
	totalDiffs := 0

	for i := 0; i < maxLen; i++ {
		var expLine, actLine string
		hasExp := i < len(expLines)
		hasAct := i < len(actLines)
		if hasExp {
			expLine = expLines[i]
		}
		if hasAct {
			actLine = actLines[i]
		}

		if hasExp && hasAct && expLine == actLine {
			continue
		}

		totalDiffs++
		if diffsShown >= maxDiffLines {
			continue
		}
		diffsShown++

		if hasExp {
			fmt.Fprintf(&b, "  L%d - esperado: %q\n", i+1, truncateLine(expLine))
		} else {
			fmt.Fprintf(&b, "  L%d - esperado: <EOF>\n", i+1)
		}
		if hasAct {
			fmt.Fprintf(&b, "  L%d + obtido:   %q\n", i+1, truncateLine(actLine))
		} else {
			fmt.Fprintf(&b, "  L%d + obtido:   <EOF>\n", i+1)
		}
	}

	if totalDiffs > diffsShown {
		fmt.Fprintf(&b, "  ... (+%d linha(s) diferente(s) omitida(s))\n", totalDiffs-diffsShown)
	}

	return false, strings.TrimRight(b.String(), "\n")
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func truncateLine(s string) string {
	const maxLen = 120
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
