# plugin-openclaw

The `plugin-openclaw` candy of the [opencharly/charly](https://github.com/opencharly/charly)
plugin library, as a standalone repo (kind-prefixed naming). The candy manifest lives at
`candy/plugin-openclaw/`; the charly resolver fetches this repo at the pinned tag,
go-builds the provider on the host, and serves it out-of-process over go-plugin gRPC.

Serves the `openclaw:` check verb, which probes a running
[OpenClaw](https://openclaw.ai) 2.0 gateway — liveness/readiness/startup via the
`/healthz` `/readyz` `/startupz` HTTP probe endpoints, and gateway state
(status, models, channels) via the in-venue `openclaw` CLI.

Install the gateway itself with the
[`openclaw`](https://github.com/opencharly/pod-openclaw) candy.
