package main

import (
	"go/importer"
	"go/token"
	"testing"
)

func TestCamelToKebab(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Contains", "contains"},
		{"HasPrefix", "has-prefix"},
		{"NewReader", "new-reader"},
		{"EqualFold", "equal-fold"},
		{"ToUpper", "to-upper"},
		{"ReplaceAll", "replace-all"},
		{"TrimSpace", "trim-space"},
		{"NewReplacer", "new-replacer"},
		{"IndexByte", "index-byte"},
		{"Cut", "cut"},
	}
	for _, c := range cases {
		got := camelToKebab(c.in)
		if got != c.want {
			t.Errorf("camelToKebab(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestExportedFuncs_strings(t *testing.T) {
	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "gc", nil)
	pkg, err := imp.Import("strings")
	if err != nil {
		t.Fatal(err)
	}
	funcs := exportedFuncs(pkg)
	if len(funcs) == 0 {
		t.Fatal("expected exported functions from strings package")
	}
	byName := make(map[string]bool)
	for _, f := range funcs {
		byName[f.GoName] = true
	}
	for _, want := range []string{"Contains", "HasPrefix", "Split", "Join", "ToUpper"} {
		if !byName[want] {
			t.Errorf("expected %q in exported funcs", want)
		}
	}
}

func TestExportedFuncs_skipsUnbindable(t *testing.T) {
	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "gc", nil)
	// os package has functions with chan/map params that should be skipped
	pkg, err := imp.Import("os")
	if err != nil {
		t.Fatal(err)
	}
	funcs := exportedFuncs(pkg)
	for _, f := range funcs {
		// none should be chan/map/unsafe typed — just a smoke check that we get results
		if f.GoName == "" || f.LispName == "" {
			t.Errorf("got empty name in func entry: %+v", f)
		}
	}
	if len(funcs) == 0 {
		t.Error("expected some bindable functions from os package")
	}
}
