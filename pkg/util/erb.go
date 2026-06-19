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
// ERB that database.yml files actually use: reading environment variables (with
// an optional default) and file includes via ERB.new(File.read("...")).result.
// Files without ERB tags (e.g. non-Ruby projects) pass through untouched.

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
// `<%= ENV["DB_HOST"] || "localhost" %>` is as common as the fetch form. The
// `||=` assign-default form (`<%= ENV["DB_ADAPTER"] ||= "postgresql" %>`) renders
// to the same value, so it is accepted too.
var envIndexRe = regexp.MustCompile(`^ENV\[\s*['"]([^'"]+)['"]\s*\]\s*(?:\|\|=?\s*(.+?)\s*)?$`)

// fileReadRe matches `ERB.new(File.read("path")).result` with an optional
// `rescue nil` suffix — a Rails idiom for conditionally including another
// YAML fragment. The path may contain `#{Rails.root}` interpolation.
var fileReadRe = regexp.MustCompile(`^ERB\.new\(\s*File\.read\(\s*"([^"]+)"\s*\)\s*\)\.result\s*(?:rescue\s+nil\s*)?$`)

// rubyRootInterpolRe matches Ruby's `#{Rails.root}` string interpolation.
var rubyRootInterpolRe = regexp.MustCompile(`#\{Rails\.root\}`)

// renderERB evaluates ERB expressions in a database.yml file and returns the
// rendered YAML alongside any output tags it could not evaluate. Supported
// expressions: ENV[...], ENV.fetch(...), and ERB.new(File.read("...")).result.
// rootDir is the project root used to resolve #{Rails.root} in file paths.
// An unrecognised tag is left verbatim; the caller is expected to warn about
// the returned tags so the limitation is visible.
func renderERB(content string, rootDir string) (string, []string) {
	return renderERBAt(content, rootDir, 0)
}

const maxIncludeDepth = 5

func renderERBAt(content string, rootDir string, depth int) (string, []string) {
	var unresolved []string

	out := erbOutputRe.ReplaceAllStringFunc(content, func(tag string) string {
		expr := erbOutputRe.FindStringSubmatch(tag)[1]
		if value, ok := evalEnvExpr(expr); ok {
			return value
		}
		if m := fileReadRe.FindStringSubmatch(expr); m != nil && depth < maxIncludeDepth {
			path := rubyRootInterpolRe.ReplaceAllString(m[1], rootDir)
			data, err := os.ReadFile(path)
			if err != nil {
				return ""
			}
			rendered, nested := renderERBAt(string(data), rootDir, depth+1)
			unresolved = append(unresolved, nested...)
			return rendered
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
// Backtick expressions (`cmd`) are Ruby shell commands we cannot evaluate; those
// return "" so the result stays valid YAML.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	if len(s) > 0 && s[0] == '`' {
		return ""
	}
	return s
}
