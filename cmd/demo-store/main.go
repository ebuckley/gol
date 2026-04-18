// demo-store shows WrapObject: a mutable Go struct manipulated entirely from lisp.
package main

import (
	"fmt"
	"os"

	"github.com/ebuckley/gol/lisp"
)

// Store is a simple in-memory key-value store with typed counters.
type Store struct {
	data     map[string]string
	counters map[string]int
}

func NewStore() *Store {
	return &Store{
		data:     make(map[string]string),
		counters: make(map[string]int),
	}
}

func (s *Store) Set(key, value string) { s.data[key] = value }
func (s *Store) Get(key string) string  { return s.data[key] }
func (s *Store) Has(key string) bool    { _, ok := s.data[key]; return ok }
func (s *Store) Delete(key string)      { delete(s.data, key) }

func (s *Store) Increment(key string) int {
	s.counters[key]++
	return s.counters[key]
}
func (s *Store) Count(key string) int { return s.counters[key] }

const program = `
(do
  (:= store-set  (get store "Set"))
  (:= store-get  (get store "Get"))
  (:= store-has  (get store "Has"))
  (:= store-del  (get store "Delete"))
  (:= store-incr (get store "Increment"))
  (:= store-cnt  (get store "Count"))

  (store-set "name" "Alice")
  (store-set "lang" "gol")

  (println (store-get "name"))
  (println (store-has "lang"))
  (println (store-has "missing"))

  (store-del "lang")
  (println (store-has "lang"))

  (store-incr "visits")
  (store-incr "visits")
  (store-incr "visits")
  (println (store-cnt "visits")))
`

func main() {
	scope := lisp.DefaultScope()

	store := NewStore()
	obj, err := lisp.WrapObject(store)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wrap error:", err)
		os.Exit(1)
	}
	scope.Set("store", obj)

	_, err = lisp.EvalProgram(program, scope)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
