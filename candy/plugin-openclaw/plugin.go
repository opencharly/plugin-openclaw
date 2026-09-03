// Package openclaw is the charly plugin serving the `openclaw` check verb — an
// importable root package plus its own go.mod. It probes a running OpenClaw 2.0
// gateway so a candy or box plan can assert gateway health declaratively:
//
//   - HTTP probe methods (health/ready/startup) hit the gateway's /healthz /readyz
//     /startupz endpoints — the container-probe surface OpenClaw 2.0 added for
//     orchestrators — from the charly HOST against the resolved gateway endpoint
//     (cc.ResolveEndpoint + cc.HTTPDo).
//   - CLI methods (status/models/channels/version) run the `openclaw` binary INSIDE
//     the venue over the reverse channel (cc.Exec), where the gateway's own CLI
//     connects to the loopback gateway over WebSocket.
//
// Why this is not the generic `http:` verb. `http:` already does URL + status +
// body/header matchers, and duplicating that would be pointless. This verb exists
// for the two things `http:` structurally cannot do:
//
//   - ENDPOINT RESOLUTION. cc.ResolveEndpoint turns the in-venue gateway port into a
//     host-reachable address, so ONE authored step works unchanged against a pod's
//     published port and a VM's forwarded one.
//   - DOMAIN SEMANTICS. Methods map to the gateway's real probe surface and return
//     real verdicts, with `json_path:` to assert one field instead of
//     pattern-matching a JSON blob; the CLI methods run in-venue where the gateway
//     CLI lives, which `http:` cannot reach at all.
//
// Dual-placement by construction: the SAME NewProvider()/NewMeta() compile INTO
// charly in-process when listed in compiled_plugins, or cmd/serve serves them
// OUT-OF-PROCESS over go-plugin gRPC when they are not — placement is invisible
// above the provider registry.
package openclaw

import (
	"embed"

	"github.com/opencharly/sdk"
	pb "github.com/opencharly/spec/proto"
)

//go:embed schema/*.cue
var schemaFS embed.FS

// pluginCalVer is this candy's CalVer, advertised over Describe. It must match the
// `version:` in charly.yml — the host reports it when the verb resolves.
const pluginCalVer = "2026.246.0100"

// NewProvider returns the openclaw provider.
func NewProvider() pb.ProviderServer { return &provider{} }

// NewMeta advertises verb:openclaw plus the plugin's self-contained CUE schema (via
// sdk.NewMeta → BuildCapabilities). The verb's whole authoring contract — the method
// enum and every openclaw-exclusive modifier — lives in the served #OpenclawInput
// (schema/openclaw.cue), which the host splices onto the base and validates every
// authored `openclaw:` step's plugin_input against.
//
// Primary is "method", so the scalar sugar `openclaw: health` desugars to
// {method: "health"} the same way the other live probe verbs do.
func NewMeta() pb.PluginMetaServer {
	return sdk.NewMeta(pluginCalVer,
		[]sdk.ProvidedCapability{{
			Class:    "verb",
			Word:     "openclaw",
			InputDef: "#OpenclawInput",
			Primary:  "method",
		}},
		schemaFS)
}
