// Package openapi embeds the hand-maintained OpenAPI spec (openapi.yaml) so
// it ships inside the compiled binary — no extra file to copy into the
// Docker image or keep in sync with a separate deploy step.
package openapi

import _ "embed"

//go:embed openapi.yaml
var OpenAPISpec []byte
