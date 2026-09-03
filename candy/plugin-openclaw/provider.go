package openclaw

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/opencharly/plugin-openclaw/candy/plugin-openclaw/params"
	"github.com/opencharly/sdk"
	"github.com/opencharly/sdk/kit"
	pb "github.com/opencharly/spec/proto"
	"github.com/opencharly/spec/spec"
)

// provider.go is the out-of-process openclaw verb provider. charly's host dispatches a
// `openclaw:` step to it through the registry (ResolveVerb("openclaw") → grpcProvider →
// Provider.Invoke) with the FULL #Op marshaled as params_json and a CheckEnv snapshot as
// env_json. The out-of-process path runs no host-side matcher pipeline, so this Invoke
// OWNS the whole verdict: resolve the gateway endpoint, dispatch the method (HTTP probe
// from the host, or in-venue CLI), then evaluate the stdout/stderr/exit_status matchers
// through the shared sdk implementation (R3).

// defaultGatewayPort is the openclaw gateway's published port (the openclaw candy's
// port 18789, relayed from the loopback-bound gateway by socat).
const defaultGatewayPort = 18789

// openclawEnv is the plugin-side decode of the CheckEnv the host ships as the
// Operation env for a `openclaw:` step.
type openclawEnv struct {
	Box  string `json:"box"`
	Mode string `json:"mode"` // "live" | "box"
}

type provider struct{ pb.UnimplementedProviderServer }

// Invoke runs one `openclaw:` operation.
func (provider) Invoke(ctx context.Context, req *pb.InvokeRequest) (*pb.InvokeReply, error) {
	var op spec.Op
	if len(req.GetParamsJson()) > 0 {
		if err := json.Unmarshal(req.GetParamsJson(), &op); err != nil {
			return sdk.ResultJSON("fail", "openclaw: decode op: "+err.Error())
		}
	}
	var in params.OpenclawInput
	kit.DecodeInput(op.PluginInput, &in)
	var env openclawEnv
	if len(req.GetEnvJson()) > 0 {
		_ = json.Unmarshal(req.GetEnvJson(), &env)
	}
	method := in.Method

	// Live-deployment verb: there is no running gateway during `charly check box` (a
	// disposable build container), so skip rather than fail — the same contract the
	// other live probe verbs follow.
	if env.Mode == "box" {
		return sdk.ResultJSON("skip", fmt.Sprintf(
			"openclaw: %s requires a running gateway (skip under charly check box)", method))
	}

	cc, err := sdk.NewCheckContext(req.GetExecutorBrokerId(), req.GetEnvJson())
	if err != nil {
		return sdk.ResultJSON("fail", fmt.Sprintf("openclaw: %s: %v", method, err))
	}

	// CLI methods run inside the venue where the `openclaw` binary is installed; HTTP
	// probe methods run from the host against the resolved gateway endpoint. Branching
	// here keeps the two contracts (argv + exit code vs HTTP status + body) from
	// bleeding into each other.
	if isCLIMethod(method) {
		return invokeCLI(ctx, cc, &op, &in)
	}

	port := int(in.Port)
	if port == 0 {
		port = defaultGatewayPort
	}
	addr, err := cc.ResolveEndpoint(ctx, port)
	if err != nil {
		return sdk.ResultJSON("fail", fmt.Sprintf("openclaw: %s: %v", method, err))
	}
	// No live venue (no-box context) → skip, the analogue of the host's empty-box skip.
	if addr == "" {
		return sdk.ResultJSON("skip", fmt.Sprintf("openclaw: %s has no resolved gateway endpoint (box=%q)", method, env.Box))
	}

	out, runErr := dispatchHTTP(ctx, cc, addr, &op, &in)

	// The shared exit/stdout/stderr verdict pipeline (R3).
	return sdk.VerbVerdict("openclaw", method, out, runErr, &op, false)
}
