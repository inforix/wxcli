package markup

import (
	"strings"
	"testing"
)

func TestMarkdownToHTML(t *testing.T) {
	out, err := MarkdownToHTML("# Title\n\nHello")
	if err != nil {
		t.Fatalf("markdown to html: %v", err)
	}
	if !strings.Contains(out, "<h1>Title</h1>") {
		t.Fatalf("expected h1 in output, got %q", out)
	}
}

func TestInlineCSS(t *testing.T) {
	out, err := InlineCSS(`<p class="x">Hi</p>`, `.x { color: red; }`)
	if err != nil {
		t.Fatalf("inline css: %v", err)
	}
	if !strings.Contains(out, `style="color: red;`) {
		t.Fatalf("expected inline style in output, got %q", out)
	}
}

func TestTightenListParagraphs(t *testing.T) {
	in := "- a\n\n- b\n"
	html, err := MarkdownToHTML(in)
	if err != nil {
		t.Fatalf("markdown to html: %v", err)
	}
	out, err := TightenListParagraphs(html)
	if err != nil {
		t.Fatalf("tighten list: %v", err)
	}
	if strings.Contains(out, "<li><p>") {
		t.Fatalf("expected list items without paragraph wrapper, got %q", out)
	}
}

func TestRemoveListBreaks(t *testing.T) {
	in := "<ul><li>a<br/>b</li></ul>"
	out, err := RemoveListBreaks(in)
	if err != nil {
		t.Fatalf("remove list breaks: %v", err)
	}
	if strings.Contains(out, "<br") {
		t.Fatalf("expected no br inside lists, got %q", out)
	}
}

func TestWrapHeadingContent(t *testing.T) {
	in := "<h3>Title</h3>"
	out, err := WrapHeadingContent(in)
	if err != nil {
		t.Fatalf("wrap heading content: %v", err)
	}
	if !strings.Contains(out, `<span class="content">Title</span>`) {
		t.Fatalf("expected content span, got %q", out)
	}
}

func TestConstrainCodeWidth(t *testing.T) {
	html, err := MarkdownToHTML("```\ncode\n```")
	if err != nil {
		t.Fatalf("markdown to html: %v", err)
	}
	out, err := ConstrainCodeWidth(html)
	if err != nil {
		t.Fatalf("constrain code width: %v", err)
	}
	if !strings.Contains(out, "max-width: 100%") {
		t.Fatalf("expected max-width style, got %q", out)
	}
}

func TestNormalizeCodeLeaf(t *testing.T) {
	in := `<pre><code><section leaf=""><span leaf="">x</span></section></code></pre>`
	out, err := NormalizeCodeLeaf(in)
	if err != nil {
		t.Fatalf("normalize code leaf: %v", err)
	}
	if strings.Contains(out, "leaf=") {
		t.Fatalf("expected leaf attributes removed, got %q", out)
	}
	if !strings.Contains(out, "<div>") {
		t.Fatalf("expected div wrapper, got %q", out)
	}
	if strings.Contains(out, "<span") {
		t.Fatalf("expected span replaced, got %q", out)
	}
}

func TestConvertPreToDiv(t *testing.T) {
	in := `<pre style="white-space: pre;">code</pre>`
	out, err := ConvertPreToDiv(in)
	if err != nil {
		t.Fatalf("convert pre to div: %v", err)
	}
	if strings.Contains(out, "<pre") {
		t.Fatalf("expected pre removed, got %q", out)
	}
	if !strings.Contains(out, "data-code-block") {
		t.Fatalf("expected data-code-block, got %q", out)
	}
	if !strings.Contains(out, "font-family: monospace") {
		t.Fatalf("expected fixed-width font, got %q", out)
	}
}

func TestNormalizeCodeBlocks(t *testing.T) {
	in := "```\nline1\n\nline3\n```"
	html, err := MarkdownToHTML(in)
	if err != nil {
		t.Fatalf("markdown to html: %v", err)
	}
	out, err := NormalizeCodeBlocks(html)
	if err != nil {
		t.Fatalf("normalize code blocks: %v", err)
	}
	if !strings.Contains(out, `class="code-snippet__js`) {
		t.Fatalf("expected code-snippet class, got %q", out)
	}
	if !strings.Contains(out, `<pre class="code-snippet__js code-snippet code-snippet_nowrap" data-lang="json">`) {
		t.Fatalf("expected pre wrapper, got %q", out)
	}
	if strings.Count(out, "<code>") < 2 {
		t.Fatalf("expected per-line code tags, got %q", out)
	}
	if strings.Contains(out, "<br") {
		t.Fatalf("expected no br tags, got %q", out)
	}
	if !strings.Contains(out, "&nbsp;") && !strings.Contains(out, "\u00a0") {
		t.Fatalf("expected blank line preserved, got %q", out)
	}
}

func TestCleanupProseMirrorArtifacts(t *testing.T) {
	in := `<section class="code-snippet__js"><pre class="code-snippet__js code-snippet code-snippet_nowrap" data-lang="json"><code><span leaf=""><br class="ProseMirror-trailingBreak"></span></code></pre></section>`
	out, err := CleanupProseMirrorArtifacts(in)
	if err != nil {
		t.Fatalf("cleanup artifacts: %v", err)
	}
	if strings.Contains(out, "ProseMirror-trailingBreak") || strings.Contains(out, "leaf=") {
		t.Fatalf("expected artifacts removed, got %q", out)
	}
}

func TestStripNewlines(t *testing.T) {
	// Test that newlines are removed from regular text
	in := "<p>Hello\nWorld</p><p>Test</p>"
	out, err := StripNewlines(in)
	if err != nil {
		t.Fatalf("strip newlines: %v", err)
	}
	if strings.Contains(out, "\n") {
		t.Fatalf("expected no newlines in output, got %q", out)
	}
	if !strings.Contains(out, "HelloWorld") {
		t.Fatalf("expected HelloWorld without newline, got %q", out)
	}

	// Test that newlines in code blocks are converted to <br>
	in2 := "<pre>line1\nline2</pre>"
	out2, err := StripNewlines(in2)
	if err != nil {
		t.Fatalf("strip newlines: %v", err)
	}
	if strings.Contains(out2, "\n") {
		t.Fatalf("expected no newlines in code block, got %q", out2)
	}
	if !strings.Contains(out2, "<br/>") {
		t.Fatalf("expected <br/> in code block, got %q", out2)
	}
}
