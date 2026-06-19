package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderERB(t *testing.T) {
	t.Setenv("PGDATABASE", "noticehub-app-dev")
	t.Setenv("PGHOST", "db.internal")
	// PGUSER and PGPASSWORD are intentionally left unset to exercise defaults.

	cases := []struct {
		name string
		in   string
		out  string
	}{
		{"fetch with block default, env set", `<%= ENV.fetch("PGDATABASE") { "fallback" } %>`, "noticehub-app-dev"},
		{"fetch with block default, env unset", `<%= ENV.fetch("PGUSER") { "localhost" } %>`, "localhost"},
		{"fetch with empty block default", `<%= ENV.fetch("PGPASSWORD") { "" } %>`, ""},
		{"fetch with bare integer default", `<%= ENV.fetch("RAILS_MAX_THREADS") { 10 } %>`, "10"},
		{"fetch with argument default", `<%= ENV.fetch("PGUSER", "postgres") %>`, "postgres"},
		{"fetch without default, env unset", `<%= ENV.fetch("PGUSER") %>`, ""},
		{"index, env set", `<%= ENV["PGHOST"] %>`, "db.internal"},
		{"index, env unset", `<%= ENV["MISSING"] %>`, ""},
		{"index with || fallback, env set", `<%= ENV["PGHOST"] || "localhost" %>`, "db.internal"},
		{"index with || fallback, env unset", `<%= ENV["MISSING"] || "localhost" %>`, "localhost"},
		{"index with || fallback, single quotes", `<%= ENV['MISSING'] || 'fallback' %>`, "fallback"},
		{"index with ||= fallback, env set", `<%= ENV["PGHOST"] ||= "localhost" %>`, "db.internal"},
		{"index with ||= fallback, env unset", `<%= ENV['DB_ADAPTER'] ||= 'postgresql' %>`, "postgresql"},
		{"single quotes", `<%= ENV.fetch('PGHOST') { 'x' } %>`, "db.internal"},
		{"trim markers", `<%=- ENV.fetch("PGHOST") { "x" } -%>`, "db.internal"},
		{"comment tag stripped", `before<%# secret %>after`, "beforeafter"},
		{"no erb passes through", "database: plain-name", "database: plain-name"},
		{"unrecognised expr left verbatim", `<%= Rails.env %>`, `<%= Rails.env %>`},
		{"fetch with backtick command default", `<%= ENV.fetch("PGUSER", ` + "`whoami`.strip" + `) %>`, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _ := renderERB(tc.in, "")
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestRenderERBReportsUnresolved(t *testing.T) {
	t.Setenv("PGHOST", "db.internal")

	out, unresolved := renderERB(`a: <%= ENV["PGHOST"] %>
b: <%= Rails.application.credentials.dig(:db) %>
c: <%= ENV["MISSING"] || "ok" %>`, "")

	assert.Equal(t, "a: db.internal\nb: <%= Rails.application.credentials.dig(:db) %>\nc: ok", out)
	assert.Equal(t, []string{`<%= Rails.application.credentials.dig(:db) %>`}, unresolved)
}

func TestRenderERBFullConfig(t *testing.T) {
	t.Setenv("PGDATABASE", "noticehub-app-dev")

	in := `development:
  host: "localhost"
  database: <%= ENV.fetch("PGDATABASE") { "noticehub-app-dev" } %>
  username: <%= ENV.fetch("PGUSER") { "" } %>
  pool: <%= ENV.fetch("RAILS_MAX_THREADS") { 10 } %>
`
	out := "development:\n" +
		"  host: \"localhost\"\n" +
		"  database: noticehub-app-dev\n" +
		"  username: \n" + // empty value leaves the trailing space; YAML reads it as nil
		"  pool: 10\n"
	rendered, _ := renderERB(in, "")
	assert.Equal(t, out, rendered)
}

func TestRenderERBFileInclude(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	os.MkdirAll(configDir, 0o755)

	os.WriteFile(filepath.Join(configDir, "database.creds.yml"), []byte(`  username: "admin"
  password: "secret"`), 0o644)

	in := `development:
  adapter: postgresql
  database: myapp_dev
<%= ERB.new(File.read("#{Rails.root}/config/database.creds.yml")).result rescue nil %>
`
	rendered, unresolved := renderERB(in, root)

	expected := `development:
  adapter: postgresql
  database: myapp_dev
  username: "admin"
  password: "secret"
`
	assert.Equal(t, expected, rendered)
	assert.Empty(t, unresolved)
}

func TestRenderERBFileIncludeMissing(t *testing.T) {
	root := t.TempDir()

	in := `development:
  database: myapp_dev
<%= ERB.new(File.read("#{Rails.root}/config/database.creds.yml")).result rescue nil %>
`
	rendered, unresolved := renderERB(in, root)

	expected := `development:
  database: myapp_dev

`
	assert.Equal(t, expected, rendered)
	assert.Empty(t, unresolved)
}

func TestRenderERBFileIncludeWithERB(t *testing.T) {
	t.Setenv("DB_USER", "fromenv")
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	os.MkdirAll(configDir, 0o755)

	os.WriteFile(filepath.Join(configDir, "creds.yml"), []byte(`  username: <%= ENV["DB_USER"] %>`), 0o644)

	in := `development:
  database: myapp_dev
<%= ERB.new(File.read("#{Rails.root}/config/creds.yml")).result rescue nil %>
`
	rendered, unresolved := renderERB(in, root)

	expected := `development:
  database: myapp_dev
  username: fromenv
`
	assert.Equal(t, expected, rendered)
	assert.Empty(t, unresolved)
}
