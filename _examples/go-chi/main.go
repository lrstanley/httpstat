package main

import (
	"expvar"
	_ "expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/lrstanley/httpstat"
	"github.com/lrstanley/httpstat/statgraph"
)

func main() {
	// This example includes usage of the statgraph sub-package!
	stats := httpstat.New("", &httpstat.HistoryOptions{Enabled: true})
	defer stats.Close()

	r := chi.NewRouter()
	r.Use(stats.Record)

	r.Mount("/debug/vars", expvar.Handler())
	r.Mount("/graphs*", http.StripPrefix("/graphs", statgraph.New(stats)))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "Hello Gopher") })

	http.ListenAndServe(":8080", r)
}
