package main

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

func renderMarkdown(source string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return source
	}
	return buf.String()
}
