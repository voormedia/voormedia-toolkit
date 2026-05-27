package util

import (
	"os"
	"regexp"
	"strings"
)

// Rails database.yml files are ERB templates: values like the database name are
// commonly written as `<%= ENV.fetch("PGDATABASE") { "myapp-dev" } %>`. Rails
// renders these before parsing the YAML; a plain YAML parser sees the raw tag
// instead and tries to connect to a database literally named `<%= ... %>`.
//
// vmt has no Ruby runtime to lean on, so renderERB evaluates the small subset of
// ERB that database.yml files actually use: reading environment variables, with
// an optional default. Files without ERB tags (e.g. non-Ruby projects) pass
// through untouched.

// erbOutputRe matches an ERB output tag `<%= expr %>`, tolerating the
// whitespace-trim markers Rails permits (`<%=- ... -%>`).
var erbOutputRe = regexp.MustCompile(`(?s)<%=-?\s*(.*?)\s*-?%>`)

// erbOtherRe matches non-output tags — comments `<%# ... %>` and bare code
// `<% ... %>` — which produce no output in Rails and are stripped here.
var erbOtherRe = regexp.MustCompile(`(?s)<%[^=].*?%>|<%%>`)

// envFetchRe matches `ENV.fetch("NAME")` with an optional default supplied
// either as a second argument (`, "default"`) or a block (`{ "default" }`).
var envFetchRe = regexp.MustCompile(`^ENV\.fetch\(\s*['"]([^'"]+)['"]\s*(?:,\s*(.+?)\s*)?\)\s*(?:\{\s*(.+?)\s*\})?$`)

// envIndexRe matches `ENV["NAME"]`, optionally with a `|| "default"` fallback —
// `<%= ENV["DB_HOST"] || "localhost" %>` is as common as the fetch form.
var envIndexRe = regexp.MustCompile(`^ENV\[\s*['"]([^'"]+)['"]\s*\]\s*(?:\|\|\s*(.+?)\s*)?$`)

// renderERB evaluates the ENV-reading ERB expressions in a database.yml file and
// returns the rendered YAML alongside any output tags it could not evaluate. An
// unrecognised tag is left verbatim, but since that verbatim text is itself valid
// YAML it would not fail at parse time — it would silently become the value and
// resurface later as an opaque connection error. The caller is expected to warn
// about the returned tags so the limitation is visible.
func renderERB(content string) (string, []string) {
	var unresolved []string

	out := erbOutputRe.ReplaceAllStringFunc(content, func(tag string) string {
		expr := erbOutputRe.FindStringSubmatch(tag)[1]
		if value, ok := evalEnvExpr(expr); ok {
			return value
		}
		unresolved = append(unresolved, tag)
		return tag
	})

	return erbOtherRe.ReplaceAllString(out, ""), unresolved
}

// evalEnvExpr evaluates a single ERB expression. The bool result reports whether
// the expression was recognised; an unrecognised expression is left untouched.
func evalEnvExpr(expr string) (string, bool) {
	if m := envIndexRe.FindStringSubmatch(expr); m != nil {
		if value, ok := os.LookupEnv(m[1]); ok {
			return value, true
		}
		// Key absent: Ruby's `ENV["X"]` is nil, so `|| default` applies. When
		// there's no fallback, m[2] is empty and unquote yields "".
		return unquote(m[2]), true
	}

	if m := envFetchRe.FindStringSubmatch(expr); m != nil {
		if value, ok := os.LookupEnv(m[1]); ok {
			return value, true
		}
		if m[3] != "" { // block default: ENV.fetch("X") { "default" }
			return unquote(m[3]), true
		}
		if m[2] != "" { // argument default: ENV.fetch("X", "default")
			return unquote(m[2]), true
		}
		return "", true
	}

	return "", false
}

// unquote strips a single matched pair of surrounding quotes, leaving bare
// literals (numbers, booleans) as-is. Ruby's `""` default becomes an empty string.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
