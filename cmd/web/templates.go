package main

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"html/template"
	"time"
)

func humanTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006")
}

func RFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func markdownProcessor(args ...interface{}) template.HTML {
	md := []byte(fmt.Sprintf("%s", args...))
	htmlBytes := markdown.ToHTML(md, nil, nil)

	return template.HTML(htmlBytes)
}

func limitText(t string) string {
	limit := 2000
	if len(t) > limit {
		return t[:limit] + "..."
	}
	return t
}

var functions = template.FuncMap{
	"humanTime":         humanTime,
	"humanDate":         humanDate,
	"markdownProcessor": markdownProcessor,
	"limitText":         limitText,
	"RFC3339":           RFC3339,
}
