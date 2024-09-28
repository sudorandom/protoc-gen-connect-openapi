package converter_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"time"

	"buf.build/gen/go/connectrpc/eliza/connectrpc/go/connectrpc/eliza/v1/elizav1connect"
	"github.com/sudorandom/protoc-gen-connect-openapi/converter"
)

var tmplElements = template.Must(template.New("name").Parse(`<!doctype html>
<html lang="en">
	<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<title>OpenAPI Documentation</title>
	<script src="https://unpkg.com/@stoplight/elements@8.3.4/web-components.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements@8.3.4/styles.min.css">
	</head>
	<body>

	<elements-api
		id="docs"
		router="hash"
		layout="sidebar"
	/>
	<script>
	(async () => {
		const docs = document.getElementById('docs');
		docs.apiDescriptionDocument = atob("{{ .DocumentBase64 }}");
	})();
	</script>

	</body>
</html>`))

func ExampleGenerateSingle_withEndpoints() {
	mux := http.NewServeMux()
	mux.Handle(elizav1connect.NewElizaServiceHandler(&elizav1connect.UnimplementedElizaServiceHandler{}))
	openapiBody, err := converter.GenerateSingle(
		converter.WithGlobal(),
		converter.WithContentTypes(
			"json",
			"proto",
		),
		converter.WithStreaming(true),
		converter.WithBaseOpenAPI([]byte(`
openapi: 3.1.0
info:
  title: OpenAPI Documentation of gRPC Services
  description: This is documentation that was generated from [protoc-gen-connect-openapi](https://github.com/sudorandom/protoc-gen-connect-openapi).
`)))
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	generationTime := time.Now()

	mux.Handle("GET /openapi.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := tmplElements.Execute(w, struct{ DocumentBase64 string }{
			DocumentBase64: base64.StdEncoding.EncodeToString(openapiBody),
		}); err != nil {
			slog.Error("rendering_template", "error", err)
		}
	}))
	mux.Handle("GET /openapi.yaml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "openapi.yaml", generationTime, bytes.NewReader(openapiBody))
	}))

	addr := "127.0.0.1:6660"
	log.Printf("Starting connectrpc on http://%s", addr)
	log.Printf("OpenAPI Doc Page http://%s/openapi.html", addr)
	log.Printf("OpenAPI Spec http://%s/openapi.yaml", addr)
	srv := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	if err := srv.ListenAndServeTLS("cert.crt", "cert.key"); err != nil {
		log.Fatalf("error: %s", err)
	}
}

func ExampleGenerateSingle() {
	openapiBody, _ := converter.GenerateSingle(
		converter.WithGlobal(),
		converter.WithContentTypes(
			"json",
			"proto",
		),
		converter.WithStreaming(true),
		converter.WithBaseOpenAPI([]byte(`
openapi: 3.1.0
info:
  title: OpenAPI Documentation of gRPC Services
  description: This is documentation that was generated from [protoc-gen-connect-openapi](https://github.com/sudorandom/protoc-gen-connect-openapi).
`)))
	fmt.Println(string(openapiBody))
}
