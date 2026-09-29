# AGENTS.md — plugin-openclaw

Standalone out-of-tree plugin repo serving the `openclaw:` gateway check verb
(`verb:openclaw`). The plugin is a Go module at `candy/plugin-openclaw/` (module
path `github.com/opencharly/plugin-openclaw/candy/plugin-openclaw`); the root
`charly.yml` declares `discover: candy` (so the repo is a project and its candy is
scanned) and carries the embedded `openclaw-skill:` skill entity.

Canonical files:

- `candy/plugin-openclaw/charly.yml` — the `plugin-openclaw:` candy entity
  (`plugin:` block, `plan:` check).
- `candy/plugin-openclaw/` — the Go source: `plugin.go`, `provider.go`,
  `methods.go`, `schema/openclaw.cue`, `params/cue_types_gen.go`,
  `cmd/serve/main.go`.
- `charly.yml` — the root manifest (`discover: candy`) + the `openclaw-skill:`
  skill entity.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.
- `README.md` — user overview only; never agent guidance.

## Load these skills first (R0)

- `/charly-check:openclaw` — the `openclaw:` check verb reference (projected
  from this candy's own `openclaw-skill:` entity). Load before changing a method
  or its input schema.
- `/charly-internals:plugin` — the plugin authoring reference: the `plugin:`
  block, the unified Provider model, the per-plugin CUE-schema contract,
  placement.
- `/charly-check:check` — the declarative check-step surface the `openclaw:` verb
  is authored through.
- `/charly-internals:git-workflow` — before any git/PR action.

## Build / validate / test

- `go build ./...` in `candy/plugin-openclaw/` — compile the plugin module.
- `go test ./...` in `candy/plugin-openclaw/` — the plugin's Go tests.
- `charly box validate` at the repo root — the structural check (the candy +
  `plugin:` block, CUE schema).
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- The live R10 witness is an openclaw-bearing bed (pod) whose check composes this
  plugin alongside the `openclaw` candy from `opencharly/pod-openclaw`.

## Modify this repo

- Edit the `plugin-openclaw:` candy entity, the Go source, and
  `schema/openclaw.cue` **together** — the schema is the single source for the
  verb's `params/` struct, so a method change not mirrored in the schema desyncs
  the generated types.
- Keep the `openclaw-skill:` entity in step with any method-surface change — it
  is the projected source for `/charly-check:openclaw`.

## Landing

Load `/charly-internals:git-workflow` before any git/PR action; it owns the
landing mechanics. The authoritative rulebook is the umbrella `AGENTS.md` in
`opencharly/opencharly` and `charly/AGENTS.md` in the charly repo.
