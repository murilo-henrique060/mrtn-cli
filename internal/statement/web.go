package statement

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OpenInBrowser generates a standalone HTML file with KaTeX math rendering for the question statement
// and opens it in the system's default web browser.
func OpenInBrowser(contest, question, title, markdown string) (string, error) {
	htmlContent := buildHTML(contest, question, title, markdown)

	tempDir := filepath.Join(os.TempDir(), "mrtn-statements")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return "", fmt.Errorf("não foi possível criar diretório temporário para visualização web: %w", err)
	}

	fileName := fmt.Sprintf("mrtn-%s-%s.html", sanitizeName(contest), sanitizeName(question))
	filePath := filepath.Join(tempDir, fileName)
	if err := os.WriteFile(filePath, []byte(htmlContent), 0o644); err != nil {
		return "", fmt.Errorf("não foi possível gravar o arquivo HTML do enunciado: %w", err)
	}

	if err := launchBrowser(filePath); err != nil {
		return filePath, err
	}

	return filePath, nil
}

func sanitizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func launchBrowser(filePath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}
	return cmd.Start()
}

func buildHTML(contest, question, title, markdown string) string {
	label := strings.ToUpper(question)
	headerTitle := fmt.Sprintf("Questão %s", label)
	if title != "" {
		headerTitle = fmt.Sprintf("Questão %s - %s", label, title)
	}

	mdJSON, _ := json.Marshal(markdown)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s (%s) - mrtn</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.css">
  <script defer src="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.js"></script>
  <script defer src="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/contrib/auto-render.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
  <style>
    :root {
      --bg: #0f172a;
      --card: #1e293b;
      --text: #e2e8f0;
      --muted: #94a3b8;
      --accent: #38bdf8;
      --border: #334155;
      --code-bg: #0b1120;
    }
    @media (prefers-color-scheme: light) {
      :root {
        --bg: #f8fafc;
        --card: #ffffff;
        --text: #1e293b;
        --muted: #64748b;
        --accent: #0284c7;
        --border: #e2e8f0;
        --code-bg: #f1f5f9;
      }
    }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Inter, sans-serif;
      background: var(--bg);
      color: var(--text);
      line-height: 1.65;
      margin: 0;
      padding: 2rem 1rem;
    }
    .container {
      max-width: 820px;
      margin: 0 auto;
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 2rem 2.5rem;
      box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.2);
    }
    .badge {
      display: inline-block;
      font-size: 0.8rem;
      font-weight: 600;
      color: var(--accent);
      text-transform: uppercase;
      letter-spacing: 0.05em;
      margin-bottom: 0.25rem;
    }
    h1, h2, h3 {
      color: var(--text);
      margin-top: 1.5rem;
      margin-bottom: 0.75rem;
    }
    h1 {
      font-size: 1.75rem;
      border-bottom: 1px solid var(--border);
      padding-bottom: 0.75rem;
      margin-top: 0;
    }
    pre {
      background: var(--code-bg);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 1rem;
      overflow-x: auto;
      font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
      font-size: 0.9rem;
    }
    code {
      font-family: "JetBrains Mono", "Fira Code", Consolas, monospace;
      background: var(--code-bg);
      padding: 0.15rem 0.35rem;
      border-radius: 4px;
      font-size: 0.9em;
    }
    pre code {
      padding: 0;
      background: transparent;
    }
    table {
      width: 100%%;
      border-collapse: collapse;
      margin: 1rem 0;
    }
    th, td {
      border: 1px solid var(--border);
      padding: 0.6rem 0.85rem;
      text-align: left;
      vertical-align: top;
    }
    th {
      background: var(--code-bg);
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="badge">Maratona SBC • Prova %s</div>
    <div id="content"></div>
  </div>
  <script>
    const rawMarkdown = %s;
    document.addEventListener("DOMContentLoaded", function() {
      const contentEl = document.getElementById("content");
      if (window.marked) {
        contentEl.innerHTML = marked.parse(rawMarkdown);
      } else {
        contentEl.innerText = rawMarkdown;
      }
      if (window.renderMathInElement) {
        renderMathInElement(contentEl, {
          delimiters: [
            {left: "$$", right: "$$", display: true},
            {left: "$", right: "$", display: false},
            {left: "\\(", right: "\\)", display: false},
            {left: "\\[", right: "\\]", display: true}
          ],
          throwOnError: false
        });
      }
    });
  </script>
</body>
</html>`,
		html.EscapeString(headerTitle),
		html.EscapeString(contest),
		html.EscapeString(contest),
		string(mdJSON),
	)
}
