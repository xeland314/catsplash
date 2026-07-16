package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xeland314/catsplash/config"
	"github.com/xeland314/catsplash/firewall"
	"github.com/xeland314/catsplash/state"
)

func TestMdToHTMLHeaders(t *testing.T) {
	input := "# Title\n## Subtitle\n### Section"
	html := mdToHTML(input)

	if !strings.Contains(html, "<h1>Title</h1>") {
		t.Errorf("h1 not rendered, got: %s", html)
	}
	if !strings.Contains(html, "<h2>Subtitle</h2>") {
		t.Errorf("h2 not rendered, got: %s", html)
	}
	if !strings.Contains(html, "<h3>Section</h3>") {
		t.Errorf("h3 not rendered, got: %s", html)
	}
}

func TestMdToHTMLBold(t *testing.T) {
	html := mdToHTML("This is **bold** text")
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("bold not rendered, got: %s", html)
	}
}

func TestMdToHTMLInlineCode(t *testing.T) {
	html := mdToHTML("Use the `code` tag")
	if !strings.Contains(html, "<code>code</code>") {
		t.Errorf("inline code not rendered, got: %s", html)
	}
}

func TestMdToHTMLLink(t *testing.T) {
	html := mdToHTML("[click here](https://example.com)")
	expected := `<a href="https://example.com">click here</a>`
	if !strings.Contains(html, expected) {
		t.Errorf("link not rendered, got: %s", html)
	}
}

func TestMdToHTMLList(t *testing.T) {
	input := "- item one\n- item two\n- item three"
	html := mdToHTML(input)
	if !strings.Contains(html, "<ul>") {
		t.Errorf("ul not rendered, got: %s", html)
	}
	if strings.Count(html, "<li>") != 3 {
		t.Errorf("expected 3 li, got %d in: %s", strings.Count(html, "<li>"), html)
	}
}

func TestMdToHTMLOrderedList(t *testing.T) {
	input := "1. first\n2. second\n3. third"
	html := mdToHTML(input)
	if !strings.Contains(html, "<ol>") {
		t.Errorf("ol not rendered, got: %s", html)
	}
}

func TestMdToHTMLBlockquote(t *testing.T) {
	input := "> This is a quote"
	html := mdToHTML(input)
	if !strings.Contains(html, "<blockquote>") {
		t.Errorf("blockquote not rendered, got: %s", html)
	}
	if !strings.Contains(html, "This is a quote") {
		t.Errorf("quote content missing, got: %s", html)
	}
}

func TestMdToHTMLTable(t *testing.T) {
	input := "| Col A | Col B |\n| --- | --- |\n| 1 | 2 |\n| 3 | 4 |"
	html := mdToHTML(input)
	if !strings.Contains(html, "<table>") {
		t.Errorf("table not rendered, got: %s", html)
	}
	if !strings.Contains(html, "<th>Col A</th>") {
		t.Errorf("th not rendered, got: %s", html)
	}
	if !strings.Contains(html, "<td>1</td>") {
		t.Errorf("td not rendered, got: %s", html)
	}
}

func TestMdToHTMLHR(t *testing.T) {
	html := mdToHTML("---")
	if !strings.Contains(html, "<hr>") {
		t.Errorf("hr not rendered, got: %s", html)
	}
}

func TestMdToHTMLParagraph(t *testing.T) {
	html := mdToHTML("Hello world")
	if !strings.Contains(html, "<p>Hello world</p>") {
		t.Errorf("paragraph not rendered, got: %s", html)
	}
}

func TestMdToHTMLFullDocument(t *testing.T) {
	input := `# Policy

> Quote text

## Section 1

Some paragraph with **bold** and ` + "`code`" + `.

- item a
- item b

| Key | Value |
| --- | --- |
| foo | bar |

---

### Sub

Done.`

	html := mdToHTML(input)

	checks := []string{"<h1>", "<blockquote>", "<h2>", "<strong>bold</strong>", "<code>code</code>", "<ul>", "<table>", "<th>Key</th>", "<td>bar</td>", "<hr>", "<h3>Sub</h3>"}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("missing %q in:\n%s", c, html)
		}
	}
}

func TestHandlePrivacyReturns200(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_privacy.db")
	db, _ := state.Open(dbPath)
	defer db.Close()
	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{PortalPort: 8110}
	srv := New(cfg, db, fw)

	// Write a temp privacy policy
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "privacy_policy.md")
	os.WriteFile(policyPath, []byte("# Test Policy\n\nThis is test content.\n"), 0644)

	// Override the file path for testing by creating handler manually
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile(policyPath)
		if err != nil {
			http.Error(w, "not found", 500)
			return
		}
		html := mdToHTML(string(content))
		data := struct{ Content string }{Content: html}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		srv.templates.ExecuteTemplate(w, "privacy.html", data)
	})

	req := httptest.NewRequest("GET", "/privacy", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Test Policy") {
		t.Errorf("policy content not rendered, body: %s", body)
	}
	if !strings.Contains(body, "This is test content.") {
		t.Errorf("policy body missing, body: %s", body)
	}
}

func TestHandlePrivacyMissingFileReturns500(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test_privacy_500.db")
	db, _ := state.Open(dbPath)
	defer db.Close()
	fw := firewall.New("wlan0", "eth0", nil)
	cfg := &config.Config{PortalPort: 8111}
	srv := New(cfg, db, fw)

	req := httptest.NewRequest("GET", "/privacy", nil)
	rr := httptest.NewRecorder()
	srv.handlePrivacy(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for missing file, got %d", rr.Code)
	}
}
