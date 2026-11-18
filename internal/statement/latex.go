package statement

import (
	"regexp"
	"strings"
)

var (
	blockMathRe  = regexp.MustCompile(`(?s)\$\$(.+?)\$\$`)
	inlineMathRe = regexp.MustCompile(`\$([^\$\n]+?)\$`)
	fracRe       = regexp.MustCompile(`\\frac\{([^{}]*)\}\{([^{}]*)\}`)
	sqrtRe       = regexp.MustCompile(`\\sqrt\{([^{}]*)\}`)
	textMacroRe  = regexp.MustCompile(`\\(?:text|texttt|textbf|textit|mathrm|mathbf|mathit|mathsf|mathtt|operatorname)\{([^{}]*)\}`)
	supBraceRe   = regexp.MustCompile(`\^\{([^{}]+)\}`)
	supSingleRe  = regexp.MustCompile(`\^([0-9a-zA-Z\+\-])`)
	subBraceRe   = regexp.MustCompile(`_\{([^{}]+)\}`)
	subSingleRe  = regexp.MustCompile(`_([0-9a-zA-Z\+\-])`)
)

var superscriptMap = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
	'n': 'ⁿ', 'i': 'ⁱ', 'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ',
	'd': 'ᵈ', 'e': 'ᵉ', 'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ',
	'j': 'ʲ', 'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'o': 'ᵒ',
	'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ',
	'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
}

var subscriptMap = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
	'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
	'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
	'v': 'ᵥ', 'x': 'ₓ',
}

// symbolReplacements ordered from longer macro names to shorter ones to avoid prefix collisions.
var symbolReplacements = []struct {
	latex   string
	unicode string
}{
	{`\Leftrightarrow`, `⇔`},
	{`\leftrightarrow`, `↔`},
	{`\Rightarrow`, `⇒`},
	{`\rightarrow`, `→`},
	{`\leftarrow`, `←`},
	{`\subseteq`, `⊆`},
	{`\supseteq`, `⊇`},
	{`\emptyset`, `∅`},
	{`\varepsilon`, `ε`},
	{`\leqslant`, `≤`},
	{`\geqslant`, `≥`},
	{`\lfloor`, `⌊`},
	{`\rfloor`, `⌋`},
	{`\lceil`, `⌈`},
	{`\rceil`, `⌉`},
	{`\ldots`, `…`},
	{`\cdots`, `⋯`},
	{`\dots`, `…`},
	{`\times`, `×`},
	{`\cdot`, `·`},
	{`\div`, `÷`},
	{`\neq`, `≠`},
	{`\leq`, `≤`},
	{`\geq`, `≥`},
	{`\le`, `≤`},
	{`\ge`, `≥`},
	{`\ne`, `≠`},
	{`\pm`, `±`},
	{`\mp`, `∓`},
	{`\notin`, `∉`},
	{`\in`, `∈`},
	{`\subset`, `⊂`},
	{`\supset`, `⊃`},
	{`\cup`, `∪`},
	{`\cap`, `∩`},
	{`\sum`, `∑`},
	{`\prod`, `∏`},
	{`\infty`, `∞`},
	{`\approx`, `≈`},
	{`\equiv`, `≡`},
	{`\forall`, `∀`},
	{`\exists`, `∃`},
	{`\alpha`, `α`},
	{`\beta`, `β`},
	{`\gamma`, `γ`},
	{`\delta`, `δ`},
	{`\epsilon`, `ε`},
	{`\theta`, `θ`},
	{`\lambda`, `λ`},
	{`\mu`, `μ`},
	{`\pi`, `π`},
	{`\sigma`, `σ`},
	{`\tau`, `τ`},
	{`\phi`, `φ`},
	{`\omega`, `ω`},
	{`\Delta`, `Δ`},
	{`\Sigma`, `Σ`},
	{`\Omega`, `Ω`},
	{`\bmod`, ` mod `},
	{`\mod`, ` mod `},
	{`\log`, `log`},
	{`\min`, `min`},
	{`\max`, `max`},
	{`\gcd`, `gcd`},
	{`\lcm`, `lcm`},
	{`\sin`, `sin`},
	{`\cos`, `cos`},
	{`\tan`, `tan`},
	{`\ln`, `ln`},
	{`\to`, `→`},
	{`\gets`, `←`},
	{`\left`, ``},
	{`\right`, ``},
	{`\quad`, `  `},
	{`\qquad`, `    `},
	{`\,`, ` `},
	{`\;`, ` `},
	{`\!`, ``},
	{`\{`, `{`},
	{`\}`, `}`},
	{`\%`, `%`},
}

// RenderMathInText replaces $$...$$ and $...$ LaTeX segments in a text with readable Unicode.
func RenderMathInText(text string) string {
	out := blockMathRe.ReplaceAllStringFunc(text, func(match string) string {
		inner := match[2 : len(match)-2]
		converted := ConvertLaTeX(strings.TrimSpace(inner))
		return "\n  " + converted + "\n"
	})

	out = inlineMathRe.ReplaceAllStringFunc(out, func(match string) string {
		inner := match[1 : len(match)-1]
		return ConvertLaTeX(inner)
	})

	return out
}

// ConvertLaTeX converts a single LaTeX mathematical expression into Unicode text.
func ConvertLaTeX(expr string) string {
	s := expr

	// Process text macros like \text{...}
	for textMacroRe.MatchString(s) {
		s = textMacroRe.ReplaceAllString(s, "$1")
	}

	// Convert single character superscripts ^x and braced superscripts ^{...}
	s = convertSuperAndSub(s)

	// Process fractions \frac{a}{b} -> (a)/(b)
	for fracRe.MatchString(s) {
		s = fracRe.ReplaceAllStringFunc(s, func(m string) string {
			parts := fracRe.FindStringSubmatch(m)
			if len(parts) != 3 {
				return m
			}
			num := strings.TrimSpace(parts[1])
			den := strings.TrimSpace(parts[2])
			if len([]rune(num)) > 1 {
				num = "(" + num + ")"
			}
			if len([]rune(den)) > 1 {
				den = "(" + den + ")"
			}
			return num + "/" + den
		})
	}

	// Process square roots \sqrt{x} -> √(x)
	for sqrtRe.MatchString(s) {
		s = sqrtRe.ReplaceAllStringFunc(s, func(m string) string {
			parts := sqrtRe.FindStringSubmatch(m)
			if len(parts) != 2 {
				return m
			}
			inner := strings.TrimSpace(parts[1])
			if len([]rune(inner)) <= 1 {
				return "√" + inner
			}
			return "√(" + inner + ")"
		})
	}

	// Replace LaTeX symbol macros
	for _, item := range symbolReplacements {
		s = replaceMacro(s, item.latex, item.unicode)
	}

	// Run superscript/subscript pass again in case macros or fractions exposed new ones
	s = convertSuperAndSub(s)

	// Clean remaining curly braces used for grouping
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")

	return s
}

func convertSuperAndSub(s string) string {
	// Single-char first so inner terms like n_j inside I_{n_j} convert before outer braces
	s = supSingleRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := supSingleRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		if conv, ok := toMappedRunes(sub[1], superscriptMap); ok {
			return conv
		}
		return "^" + sub[1]
	})

	s = subSingleRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := subSingleRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		if conv, ok := toMappedRunes(sub[1], subscriptMap); ok {
			return conv
		}
		return "_" + sub[1]
	})

	s = supBraceRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := supBraceRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		if conv, ok := toMappedRunes(sub[1], superscriptMap); ok {
			return conv
		}
		return "^(" + sub[1] + ")"
	})

	s = subBraceRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := subBraceRe.FindStringSubmatch(m)
		if len(sub) != 2 {
			return m
		}
		if conv, ok := toMappedRunes(sub[1], subscriptMap); ok {
			return conv
		}
		return "_" + sub[1]
	})

	return s
}

func replaceMacro(s, macro, replacement string) string {
	// If macro ends with a letter, ensure we don't replace a prefix of another alphabetic macro
	lastChar := macro[len(macro)-1]
	isAlpha := (lastChar >= 'a' && lastChar <= 'z') || (lastChar >= 'A' && lastChar <= 'Z')
	if !isAlpha {
		return strings.ReplaceAll(s, macro, replacement)
	}

	var b strings.Builder
	idx := 0
	for {
		pos := strings.Index(s[idx:], macro)
		if pos == -1 {
			b.WriteString(s[idx:])
			break
		}
		absPos := idx + pos
		endPos := absPos + len(macro)
		b.WriteString(s[idx:absPos])

		if endPos < len(s) {
			next := s[endPos]
			nextAlpha := (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z')
			if nextAlpha {
				// Not a full macro match; keep original macro text
				b.WriteString(macro)
				idx = endPos
				continue
			}
		}

		b.WriteString(replacement)
		// Consume a trailing space if the next char is a space between macro and operand (e.g. "\le N" -> "≤ N")
		idx = endPos
	}
	return b.String()
}

func toMappedRunes(s string, table map[rune]rune) (string, bool) {
	alreadyTarget := make(map[rune]bool, len(table))
	for _, v := range table {
		alreadyTarget[v] = true
	}
	var b strings.Builder
	for _, r := range s {
		if r == ' ' {
			continue
		}
		if mapped, ok := table[r]; ok {
			b.WriteRune(mapped)
			continue
		}
		if alreadyTarget[r] {
			b.WriteRune(r)
			continue
		}
		return "", false
	}
	return b.String(), b.Len() > 0
}
