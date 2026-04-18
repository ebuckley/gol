//go:build !cgo

// demo-http shows how to expose net/http as gol callables.
// fetch returns the response body as a string.
// http-get returns a wrapped *http.Response so lisp can inspect Status/StatusCode.
//
// Build: CGO_ENABLED=0 go run ./cmd/demo-http
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ebuckley/gol/lisp"
)

// fetch is a convenience wrapper: returns the full response body as a string.
func fetch(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

// get returns a wrapped *http.Response so the lisp program can inspect fields.
func get(url string) (lisp.ObjectNode, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return lisp.ObjectNode{}, err
	}
	resp.Body.Close()
	return lisp.WrapObject(resp)
}

const program = `
(do
  (println "--- simple fetch ---")
  (:= (body err) (fetch "https://httpbin.org/get"))
  (if err (println err) (println (contains body "httpbin")))

  (println "--- inspecting response ---")
  (:= (resp err2) (http-get "https://httpbin.org/status/418"))
  (if err2 (println err2) (do
    (println (get resp "Status"))
    (println (get resp "StatusCode")))))
`

func main() {
	scope := lisp.DefaultScope()

	scope.Set("fetch", lisp.GoFunc(fetch))
	scope.Set("http-get", lisp.GoFunc(get))
	scope.Set("contains", lisp.GoFunc(strings.Contains))
	_, err := lisp.EvalProgram(program, scope)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
