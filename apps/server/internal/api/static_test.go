package api

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestEmbeddedSPAHandlerServesNestedAsset(t *testing.T) {
	files := fstest.MapFS{
		"index.html":         {Data: []byte("<main>app</main>")},
		"assets/client.js":   {Data: []byte("globalThis.loaded = true")},
		"assets/client.css":  {Data: []byte("body { color: white }")},
		"assets/nested/icon": {Data: []byte("icon")},
	}
	handler := embeddedSPAHandler(files)

	for requestPath, expected := range map[string]string{
		"/assets/client.js":   "globalThis.loaded = true",
		"//assets/client.css": "body { color: white }",
		"/unknown/route":      "<main>app</main>",
	} {
		request := httptest.NewRequest(http.MethodGet, requestPath, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", requestPath, response.Code)
		}
		if response.Body.String() != expected {
			t.Fatalf("%s: expected %q, got %q", requestPath, expected, response.Body.String())
		}
	}
}

func TestCleanURLPathIsPortableAndRelative(t *testing.T) {
	for input, expected := range map[string]string{
		"/assets/client.js":   "assets/client.js",
		"//assets/client.js":  "assets/client.js",
		"assets/client.js":    "assets/client.js",
		"/../assets/icon.png": "assets/icon.png",
		"/":                   "",
	} {
		if actual := cleanURLPath(input); actual != expected {
			t.Fatalf("cleanURLPath(%q): expected %q, got %q", input, expected, actual)
		}
		if !fs.ValidPath(expected) && expected != "" {
			t.Fatalf("test expectation %q is not a valid fs path", expected)
		}
	}
}
