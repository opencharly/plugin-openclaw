package openclaw

import (
	"testing"
)

// TestResolveHTTPProbe pins the HTTP probe method → endpoint mapping.
func TestResolveHTTPProbe(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{"health", "/healthz"},
		{"ready", "/readyz"},
		{"startup", "/startupz"},
	}
	for _, c := range cases {
		probe, err := resolveHTTPProbe(c.method)
		if err != nil {
			t.Fatalf("resolveHTTPProbe(%q): %v", c.method, err)
		}
		if probe.Path != c.path {
			t.Errorf("resolveHTTPProbe(%q).Path = %q, want %q", c.method, probe.Path, c.path)
		}
	}
	if _, err := resolveHTTPProbe("bogus"); err == nil {
		t.Error("resolveHTTPProbe(bogus) = nil error, want error")
	}
}

// TestResolveCLICall pins the CLI method → argv mapping.
func TestResolveCLICall(t *testing.T) {
	cases := []struct {
		method string
		args   []string
	}{
		{"status", []string{"status", "--all"}},
		{"models", []string{"models", "status"}},
		{"channels", []string{"channels", "status"}},
		{"version", []string{"--version"}},
	}
	for _, c := range cases {
		call, err := resolveCLICall(c.method)
		if err != nil {
			t.Fatalf("resolveCLICall(%q): %v", c.method, err)
		}
		if len(call.Args) != len(c.args) {
			t.Errorf("resolveCLICall(%q).Args = %v, want %v", c.method, call.Args, c.args)
			continue
		}
		for i := range c.args {
			if call.Args[i] != c.args[i] {
				t.Errorf("resolveCLICall(%q).Args[%d] = %q, want %q", c.method, i, call.Args[i], c.args[i])
			}
		}
	}
	if _, err := resolveCLICall("bogus"); err == nil {
		t.Error("resolveCLICall(bogus) = nil error, want error")
	}
}

// TestIsCLIMethod pins the HTTP-vs-CLI split.
func TestIsCLIMethod(t *testing.T) {
	for _, m := range []string{"status", "models", "channels", "version"} {
		if !isCLIMethod(m) {
			t.Errorf("isCLIMethod(%q) = false, want true", m)
		}
	}
	for _, m := range []string{"health", "ready", "startup"} {
		if isCLIMethod(m) {
			t.Errorf("isCLIMethod(%q) = true, want false", m)
		}
	}
}

// TestExtractJSONPath pins the json_path extraction.
func TestExtractJSONPath(t *testing.T) {
	doc := `{"status":"ok","gateway":{"version":"2026.8.2"}}`
	if got := extractJSONPath(doc, "status"); got != `"ok"` {
		t.Errorf("extractJSONPath(status) = %s, want \"ok\"", got)
	}
	if got := extractJSONPath(doc, "gateway.version"); got != `"2026.8.2"` {
		t.Errorf("extractJSONPath(gateway.version) = %s, want \"2026.8.2\"", got)
	}
	// Missing path falls back to the raw document.
	if got := extractJSONPath(doc, "nope"); got != doc {
		t.Errorf("extractJSONPath(nope) = %s, want raw doc", got)
	}
	// Non-JSON falls back to the raw document.
	if got := extractJSONPath("not json", "status"); got != "not json" {
		t.Errorf("extractJSONPath(non-json) = %s, want raw", got)
	}
	// Empty path returns the document unchanged.
	if got := extractJSONPath(doc, ""); got != doc {
		t.Errorf("extractJSONPath(empty) = %s, want raw doc", got)
	}
}

// TestShellJoin pins the argv quoting.
func TestShellJoin(t *testing.T) {
	got := shellJoin([]string{"openclaw", "status", "--all"})
	want := "'openclaw' 'status' '--all'"
	if got != want {
		t.Errorf("shellJoin = %s, want %s", got, want)
	}
}
