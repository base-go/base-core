package template

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (e *Engine) RegisterDefaultHelpers() {
	// String helpers
	e.AddHelper("upper", strings.ToUpper)
	e.AddHelper("lower", strings.ToLower)
	e.AddHelper("title", cases.Title(language.AmericanEnglish).String)
	e.AddHelper("truncate", func(s string, length int) string {
		if len(s) <= length {
			return s
		}
		return s[:length] + "..."
	})

	// HTML helpers
	e.AddHelper("safe", func(s string) template.HTML {
		return template.HTML(s)
	})

	e.AddHelper("escape", template.HTMLEscapeString)

	// URL helpers
	e.AddHelper("url_for", func(path string) string {
		return path
	})

	e.AddHelper("link_to", func(text, url string, attrs ...string) template.HTML {
		var attrStr string
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return template.HTML(fmt.Sprintf(`<a href="%s"%s>%s</a>`, url, attrStr, text))
	})

	// Form helpers
	e.AddHelper("form_tag", func(action, method string, attrs ...string) template.HTML {
		var attrStr string
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return template.HTML(fmt.Sprintf(`<form action="%s" method="%s"%s>`, action, method, attrStr))
	})

	e.AddHelper("input_tag", func(inputType, name, value string, attrs ...string) template.HTML {
		var attrStr string
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return template.HTML(fmt.Sprintf(`<input type="%s" name="%s" value="%s"%s>`, inputType, name, value, attrStr))
	})

	e.AddHelper("submit_tag", func(value string, attrs ...string) template.HTML {
		var attrStr string
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return template.HTML(fmt.Sprintf(`<input type="submit" value="%s"%s>`, value, attrStr))
	})

	// Time helpers
	e.AddHelper("time_ago", func(t time.Time) string {
		return time.Since(t).String() + " ago"
	})

	e.AddHelper("format_time", func(t time.Time, layout string) string {
		return t.Format(layout)
	})

	// Conditional helpers
	e.AddHelper("if_eq", func(a, b interface{}) bool {
		return a == b
	})

	e.AddHelper("if_ne", func(a, b interface{}) bool {
		return a != b
	})

	// Collection helpers
	e.AddHelper("join", func(sep string, items []string) string {
		return strings.Join(items, sep)
	})

	// Alpine.js helpers
	e.AddHelper("alpine_data", func(name string, data interface{}) template.HTML {
		return template.HTML(fmt.Sprintf(`x-data="%s"`, name))
	})

	e.AddHelper("alpine_show", func(condition string) template.HTML {
		return template.HTML(fmt.Sprintf(`x-show="%s"`, condition))
	})

	e.AddHelper("alpine_click", func(action string) template.HTML {
		return template.HTML(fmt.Sprintf(`@click="%s"`, action))
	})

	e.AddHelper("alpine_model", func(variable string) template.HTML {
		return template.HTML(fmt.Sprintf(`x-model="%s"`, variable))
	})

	// CSS/JS helpers
	e.AddHelper("stylesheet_link_tag", func(href string, attrs ...string) template.HTML {
		var attrStr string
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return template.HTML(fmt.Sprintf(`<link rel="stylesheet" href="%s"%s>`, href, attrStr))
	})

	e.AddHelper("javascript_include_tag", func(src string, attrs ...string) template.HTML {
		var attrStr string
		if len(attrs) > 0 {
			attrStr = " " + strings.Join(attrs, " ")
		}
		return template.HTML(fmt.Sprintf(`<script src="%s"%s></script>`, src, attrStr))
	})
}
