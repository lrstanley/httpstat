package main

import (
	_ "expvar"
	"fmt"
	"net/http"

	"github.com/lrstanley/httpstat"
)

func main() {
	stats := httpstat.New("", nil)
	defer stats.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "Hello Gopher") })

	http.ListenAndServe(":8080", stats.Record(http.DefaultServeMux))
}
