package main

import (
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// htmlPolicy is a UGC (User Generated Content) sanitizer that allows safe HTML
// from Quill rich text editor while stripping dangerous elements like <script>,
// event handlers (onerror, onclick, etc.), and other XSS vectors.
var htmlPolicy = createHTMLPolicy()

func createHTMLPolicy() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()

	// Allow Quill editor classes
	p.AllowAttrs("class").Matching(bluemonday.Paragraph).OnElements(
		"p", "span", "div", "pre", "blockquote", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "code", "em", "strong", "u", "s", "sub", "sup",
	)

	// Allow Quill inline styles for alignment, indentation, colors
	p.AllowStyles("color", "background-color", "text-align", "padding-left").Globally()

	// Allow data attributes Quill uses for list types, indent
	p.AllowDataAttributes()

	// Allow MathLive/KaTeX elements
	p.AllowElements("math", "annotation", "semantics", "mrow", "mi", "mn", "mo",
		"msup", "msub", "mfrac", "msqrt", "mover", "munder", "mtable", "mtr", "mtd")
	p.AllowAttrs("class", "style", "aria-hidden").Globally()

	return p
}

// SafeContent sanitizes user-generated HTML content (from Quill editor) before
// rendering in templates. It strips dangerous elements/attributes while preserving
// safe formatting.
func SafeContent(s string) template.HTML {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return template.HTML("")
	}
	// If it doesn't look like HTML, escape it
	if s[0] != '<' {
		return template.HTML(template.HTMLEscapeString(s))
	}
	// Sanitize HTML through bluemonday
	return template.HTML(htmlPolicy.Sanitize(s))
}
