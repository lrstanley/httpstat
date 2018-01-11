# httpstat [![godoc](https://godoc.org/github.com/lrstanley/httpstat?status.png)](https://godoc.org/github.com/lrstanley/httpstat) [![goreport](https://goreportcard.com/badge/github.com/lrstanley/httpstat)](https://goreportcard.com/report/github.com/lrstanley/httpstat)

httpstat is a `net/http` handler for Go, which reports various metrics about
a given http server. It natively supports exporting stats to `expvar` (this
is actually how it tracks all of it's stats). httpstat also has support for
taking historical snapshots of the data. An example usecase and feature of
this is the `statgraph` subpackage, as documented [here](https://godoc.org/github.com/lrstanley/httpstat/statgraph).

An example usecase (also see [_examples/](https://github.com/lrstanley/httpstat/tree/master/_examples))
is shown below:

```go
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello Gopher")
    })

	http.ListenAndServe(":8080", stats.Record(http.DefaultServeMux))
}
```

Example output, using the expvar's default endpoint, `/debug/vars`:

```json
{
    "cmdline": [
        "/tmp/go-build807700005/command-line-arguments/_obj/exe/main"
    ],
    "httpstat_invoked": "2018-01-11T07:58:40-05:00",
    "httpstat_invoked_seconds": 1170,
    "httpstat_invoked_unix": 1515675520,
    "httpstat_pid": 17135,
    "httpstat_request_bytes_total": 8894,
    "httpstat_request_errors_total": 0,
    "httpstat_request_total": 23,
    "httpstat_request_total_seconds": 0.004517135,
    "httpstat_response_bytes_total": 50748,
    "httpstat_status_total": {
        "200": 22,
        "404": 1
    },
    "memstats": "..."
    }
}
```

You can also mount the `ServeHTTP` endpoint ([here](https://godoc.org/github.com/lrstanley/httpstat#HTTPStats.ServeHTTP))
to a custom location, which will only return the expvar variables that were
created by this application.

Note:

   * Make sure you register the handler/middleware as far up the stack that
   you want to track metrics on. Also make sure that your handlers do not
   return early, and continue writing to the ResponseWriter after they have
   returned, as httpstat cannot monitor those writes.
   * Using this library will introduce a ~550ns overhead to the total request
   processing time, however it shouldn't introduce any measurable delay during
   the server->client response times, since tracking measurement compilation
   occurs after the child middleware/handler has finished being invoked.
   * Using `History` will add a minor amount of additional overhead during
   snapshot pauses. This may be reduced in future iterations.
   * Each invocation of a new `HTTPStats` struct must be under it's own namespace,
   as `expvar` only allows variables with a given name to be registered once.
   See [httpstat.New](https://godoc.org/github.com/lrstanley/httpstat#New) for
   details.
   * The request size ("bytes in") can only be roughly calculated, as `net/http`
   strips some of the data of the request as it is being processed.

## Why?

There are a few packages similar to this one ([thoas/stats](https://github.com/thoas/stats)
and [felixge/httpsnoop](https://github.com/felixge/httpsnoop) to name a few),
however I wanted one with builtin support for historical snapshots, as well
as one which was built with `expvar` in mind. With support for `expvar`, this
can allow other packages to introspect the data that this package exports, in
a standardized way.

## License

    LICENSE: The MIT License (MIT)
    Copyright (c) 2018 Liam Stanley <me@liamstanley.io>

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in
    all copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE.
