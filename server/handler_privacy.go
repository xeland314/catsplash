package server

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func (s *Server) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("docs/privacy_policy.md")
	if err != nil {
		http.Error(w, "Privacy policy not available", http.StatusInternalServerError)
		return
	}

	html := mdToHTML(string(content))

	data := struct {
		Content string
	}{
		Content: html,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "privacy.html", data); err != nil {
		log.Printf("Failed to render privacy template: %v", err)
	}
}

func mdToHTML(md string) string {
	var out strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(md))
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024)

	inList := false
	inTable := false
	inBlockquote := false
	var listType string // "ul" or "ol"

	flushList := func() {
		if inList {
			out.WriteString("</" + listType + ">\n")
			inList = false
		}
	}
	flushTable := func() {
		if inTable {
			out.WriteString("</tbody></table>\n")
			inTable = false
		}
	}
	flushBlockquote := func() {
		if inBlockquote {
			out.WriteString("</blockquote>\n")
			inBlockquote = false
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Empty line: close open blocks
		if strings.TrimSpace(line) == "" {
			flushList()
			flushTable()
			flushBlockquote()
			continue
		}

		// Blockquote
		if strings.HasPrefix(line, "> ") {
			if !inBlockquote {
				flushList()
				flushTable()
				out.WriteString("<blockquote>\n")
				inBlockquote = true
			}
			out.WriteString("<p>" + inline(strings.TrimPrefix(line, "> ")) + "</p>\n")
			continue
		}
		flushBlockquote()

		// HR
		if strings.HasPrefix(line, "---") {
			flushList()
			flushTable()
			out.WriteString("<hr>\n")
			continue
		}

		// Headers
		if strings.HasPrefix(line, "### ") {
			flushList()
			flushTable()
			out.WriteString("<h3>" + inline(strings.TrimPrefix(line, "### ")) + "</h3>\n")
			continue
		}
		if strings.HasPrefix(line, "## ") {
			flushList()
			flushTable()
			out.WriteString("<h2>" + inline(strings.TrimPrefix(line, "## ")) + "</h2>\n")
			continue
		}
		if strings.HasPrefix(line, "# ") {
			flushList()
			flushTable()
			out.WriteString("<h1>" + inline(strings.TrimPrefix(line, "# ")) + "</h1>\n")
			continue
		}

		// Table
		if strings.HasPrefix(line, "|") {
			cells := parseTableRow(line)
			if isTableSeparator(cells) {
				continue
			}
			if !inTable {
				flushList()
				out.WriteString("<table><thead><tr>")
				for _, c := range cells {
					out.WriteString("<th>" + inline(c) + "</th>")
				}
				out.WriteString("</tr></thead><tbody>\n")
			inTable = true
			continue
			}
			out.WriteString("<tr>")
			for _, c := range cells {
				out.WriteString("<td>" + inline(c) + "</td>")
			}
			out.WriteString("</tr>\n")
			continue
		}
		flushTable()

		// Unordered list
		if strings.HasPrefix(line, "- ") {
			if !inList || listType != "ul" {
				flushList()
				flushTable()
				out.WriteString("<ul>\n")
				inList = true
				listType = "ul"
			}
			out.WriteString("<li>" + inline(strings.TrimPrefix(line, "- ")) + "</li>\n")
			continue
		}

		// Ordered list
		if len(line) > 3 && line[0] >= '0' && line[0] <= '9' && strings.HasPrefix(line[1:], ". ") {
			if !inList || listType != "ol" {
				flushList()
				flushTable()
				out.WriteString("<ol>\n")
				inList = true
				listType = "ol"
			}
			content := line[strings.Index(line, ". ")+2:]
			out.WriteString("<li>" + inline(content) + "</li>\n")
			continue
		}
		flushList()

		// Paragraph
		flushTable()
		out.WriteString("<p>" + inline(line) + "</p>\n")
	}

	flushList()
	flushTable()
	flushBlockquote()

	return out.String()
}

func inline(s string) string {
	// Bold: **text**
	for {
		start := strings.Index(s, "**")
		if start == -1 {
			break
		}
		end := strings.Index(s[start+2:], "**")
		if end == -1 {
			break
		}
		end += start + 2
		tag := "<strong>" + s[start+2:end] + "</strong>"
		s = s[:start] + tag + s[end+2:]
	}

	// Inline code: `text`
	for {
		start := strings.Index(s, "`")
		if start == -1 {
			break
		}
		end := strings.Index(s[start+1:], "`")
		if end == -1 {
			break
		}
		end += start + 1
		tag := "<code>" + s[start+1:end] + "</code>"
		s = s[:start] + tag + s[end+1:]
	}

	// Links: [text](url)
	for {
		start := strings.Index(s, "[")
		if start == -1 {
			break
		}
		closeBracket := strings.Index(s[start:], "]")
		if closeBracket == -1 {
			break
		}
		closeBracket += start
		if closeBracket+1 >= len(s) || s[closeBracket+1] != '(' {
			break
		}
		closeParen := strings.Index(s[closeBracket+2:], ")")
		if closeParen == -1 {
			break
		}
		closeParen += closeBracket + 2
		text := s[start+1 : closeBracket]
		url := s[closeBracket+2 : closeParen]
		tag := fmt.Sprintf("<a href=\"%s\">%s</a>", url, text)
		s = s[:start] + tag + s[closeParen+1:]
	}

	return s
}

func parseTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

func isTableSeparator(cells []string) bool {
	for _, c := range cells {
		trimmed := strings.TrimSpace(c)
		if trimmed == "" || !strings.Contains(trimmed, "---") {
			return false
		}
	}
	return true
}
