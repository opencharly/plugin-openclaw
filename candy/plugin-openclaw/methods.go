package openclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opencharly/plugin-openclaw/candy/plugin-openclaw/params"
	"github.com/opencharly/sdk"
	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// methods.go is the single table mapping an `openclaw:` method to one probe. Keeping
// it declarative (rather than a switch that also builds requests) means the mapping
// is unit-testable without a live gateway, and adding a method is one entry.

// httpProbe is a resolved gateway HTTP probe: the path under the gateway root.
type httpProbe struct {
	Path string
}

// cliCall is a resolved in-venue CLI invocation.
type cliCall struct {
	Args []string
}

// isCLIMethod reports whether a method runs the `openclaw` CLI inside the venue
// rather than an HTTP probe from the host.
func isCLIMethod(m string) bool {
	switch m {
	case "status", "models", "channels", "version":
		return true
	}
	return false
}

// resolveHTTPProbe builds the gateway HTTP probe for a method.
func resolveHTTPProbe(m string) (httpProbe, error) {
	switch m {
	case "health":
		// Liveness — the OpenClaw 2.0 container-probe endpoint. Unauthenticated.
		return httpProbe{Path: "/healthz"}, nil
	case "ready":
		// Readiness — configured channels healthy.
		return httpProbe{Path: "/readyz"}, nil
	case "startup":
		// Startup status — the gateway owns its listener.
		return httpProbe{Path: "/startupz"}, nil
	}
	return httpProbe{}, fmt.Errorf("openclaw: unknown HTTP probe method %q", m)
}

// resolveCLICall builds the in-venue CLI invocation for a method.
func resolveCLICall(m string) (cliCall, error) {
	switch m {
	case "status":
		return cliCall{Args: []string{"status", "--all"}}, nil
	case "models":
		return cliCall{Args: []string{"models", "status"}}, nil
	case "channels":
		return cliCall{Args: []string{"channels", "status"}}, nil
	case "version":
		return cliCall{Args: []string{"--version"}}, nil
	}
	return cliCall{}, fmt.Errorf("openclaw: unknown CLI method %q", m)
}

// dispatchHTTP runs one HTTP probe method from the host against the resolved gateway
// endpoint and returns the body (or the json_path-extracted value) for the shared
// matcher pipeline. A non-2xx status is NOT an error here — the caller's stdout
// matcher asserts the body, and the shared pipeline compares the exit status.
func dispatchHTTP(ctx context.Context, cc kit.CheckContext, addr string, op *spec.Op, in *params.OpenclawInput) (string, error) {
	probe, err := resolveHTTPProbe(in.Method)
	if err != nil {
		return "", err
	}
	timeout := in.Timeout
	if timeout == "" {
		timeout = "10s"
	}
	req := spec.CheckHTTPRequest{
		Method:  "GET",
		URL:     "http://" + addr + probe.Path,
		Timeout: timeout,
	}
	resp, err := cc.HTTPDo(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openclaw: %s: %v", in.Method, err)
	}
	body := string(resp.Body)
	if in.JSONPath != "" {
		body = extractJSONPath(body, in.JSONPath)
	}
	return body, nil
}

// invokeCLI runs one CLI method inside the venue over the reverse channel and
// returns the shared verdict.
func invokeCLI(ctx context.Context, cc kit.CheckContext, op *spec.Op, in *params.OpenclawInput) (*pb.InvokeReply, error) {
	call, err := resolveCLICall(in.Method)
	if err != nil {
		return sdk.ResultJSON("fail", err.Error())
	}
	args := append([]string{"openclaw"}, call.Args...)
	script := shellJoin(args)
	stdout, stderr, exit, err := cc.Exec().RunCapture(ctx, script)
	if err != nil {
		return sdk.ResultJSON("fail", fmt.Sprintf("openclaw: %s: %v", in.Method, err))
	}
	out := stdout
	if in.JSONPath != "" {
		out = extractJSONPath(stdout, in.JSONPath)
	}
	// The CLI's exit code is the verdict signal; a non-zero exit is reported as a
	// runErr so the shared pipeline maps it to exit=1.
	var runErr error
	if exit != 0 {
		runErr = fmt.Errorf("openclaw %s exited %d: %s", in.Method, exit, strings.TrimSpace(stderr))
	}
	return sdk.VerbVerdict("openclaw", in.Method, out, runErr, op, false)
}

// extractJSONPath pulls a dotted path out of a JSON document (e.g. "status" from
// {"status":"ok"}). Returns the raw document when the path is absent or unparsable.
func extractJSONPath(doc, path string) string {
	if path == "" {
		return doc
	}
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		return doc
	}
	cur := v
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return doc
		}
		cur, ok = m[part]
		if !ok {
			return doc
		}
	}
	b, err := json.Marshal(cur)
	if err != nil {
		return doc
	}
	return string(b)
}

// shellJoin quotes each argument for a single shell invocation.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = "'" + strings.ReplaceAll(a, "'", "'\\''") + "'"
	}
	return strings.Join(quoted, " ")
}
