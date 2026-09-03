// Command serve is the OUT-OF-PROCESS entrypoint for the openclaw verb plugin: a
// thin shim serving the importable provider over go-plugin gRPC via sdk.Serve. The
// SAME NewProvider()/NewMeta() compile INTO charly in-process when listed in
// compiled_plugins; this binary is host-built and connected only when they are not —
// placement is invisible above the registry.
package main

import (
	openclaw "github.com/opencharly/plugin-openclaw/candy/plugin-openclaw"
	"github.com/opencharly/sdk"
)

func main() { sdk.Serve(openclaw.NewProvider(), openclaw.NewMeta()) }
