package http

import (
	"encoding/json"
	"net/http"

	"backend/internal/generated/api/generated"
)

const scalarHTML = `<!doctype html>
<html>
<head>
  <title>BlessThun Food System – API Reference</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  <script
    id="api-reference"
    data-url="/docs/openapi.json"
    data-configuration='{"darkMode":false}'></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

func DocsScalarHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(scalarHTML))
	}
}

func DocsOpenAPIHandler() http.HandlerFunc {
	spec, err := generated.GetSwagger()
	if err != nil {
		panic("failed to load OpenAPI spec: " + err.Error())
	}

	specJSON, err := json.Marshal(spec)
	if err != nil {
		panic("failed to marshal OpenAPI spec: " + err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specJSON)
	}
}
