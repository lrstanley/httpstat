package main

import (
	_ "expvar"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/lrstanley/httpstat"
	"github.com/lrstanley/httpstat/statgraph"
)

func main() {
	stats := httpstat.New("", nil)
	defer stats.Close()

	r := chi.NewRouter()
	r.Use(stats.Record)
	r.Use(middleware.DefaultLogger)

	r.Mount("/debug", middleware.Profiler())
	r.Mount("/graphs*", http.StripPrefix("/graphs", statgraph.New(stats)))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Gopher")
	})

	http.ListenAndServe(":8080", r)
}
