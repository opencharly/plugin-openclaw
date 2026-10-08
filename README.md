# plugin-openclaw

The `openclaw:` check verb for OpenCharly — probe a running
[OpenClaw](https://openclaw.ai) 2.0 gateway from a candy or box plan. The probe
surface is verified against npm `openclaw@2026.9.8`.

The plugin is an out-of-tree Go module: charly fetches this repo at the pinned
tag, go-builds the provider on the host, and serves it **out-of-process** over
go-plugin gRPC via the plugin SDK. An authored `openclaw:` step dispatches
through the provider registry exactly like a built-in.

## What it provides

| Capability | Surface |
|---|---|
| `verb:openclaw` | the `openclaw:` check verb — probe a running gateway |

The verb probes two ways:

- **HTTP probes** (`health` / `ready` / `startup`) hit the `/healthz` `/readyz`
  `/startupz` endpoints from the charly host against the resolved gateway
  endpoint. Per upstream, `/healthz` shows only that the process is alive, while
  `/readyz` is the endpoint that reports recorded terminal database failures —
  prefer `ready` when a check must gate on the gateway being usable.
- **In-venue CLI** (`status` / `models` / `channels` / `version`) run the
  `openclaw` binary inside the venue over the reverse channel.

All methods are read-only — there are no mutating methods in this first cut.

### Why it is not the `http:` verb

`http:` already covers URL + status + body/header matchers. This verb exists for
what `http:` structurally cannot do: **endpoint resolution** (one authored step
works unchanged against a pod's published port and a VM's forwarded one),
**domain semantics** (methods map to the gateway's real probe surface, with
`json_path:` to assert one field), and **in-venue CLI** (`http:` cannot reach the
gateway CLI at all).

## How to use it

Compose the plugin candy in a box or check bed's `candy:` list, then author the
verb in a plan:

```yaml
- check: the gateway's /healthz liveness endpoint reports ok
  openclaw: health
  stdout: [{contains: ok}]
  eventually: 60s
  retry_interval: 5s
  context: [runtime]
```

| Field | Meaning |
|---|---|
| `method` | the operation (also the scalar-sugar primary: `openclaw: health`) |
| `port` | gateway port (default 18789) |
| `timeout` | HTTP probe timeout (default 10s) |
| `json_path` | dotted path into the response; stdout becomes that value |

## Layout

- `candy/plugin-openclaw/` — the plugin module: `plugin.go` (provider + meta),
  `provider.go`, `methods.go`, `schema/openclaw.cue` (the self-contained
  `#OpenclawInput`), `params/cue_types_gen.go`, and `cmd/serve/main.go`.
- `charly.yml` — the root project manifest (`discover: candy`) + the embedded
  `openclaw-skill:` skill entity.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.

## Related

- Owning skill: `/charly-check:openclaw` — the `openclaw:` check verb reference
  (projected from this candy's own `openclaw-skill:` entity).
- `/charly-openclaw:openclaw` — the candy that installs the gateway itself.
- `/charly-internals:plugin` — the out-of-process plugin model this verb follows.
- [`opencharly/charly`](https://github.com/opencharly/charly) — the charly CLI.
