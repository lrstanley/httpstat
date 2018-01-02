// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package httpstat

import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// HTTPStats holds the state of the current recorder middleware.
type HTTPStats struct {
	namespace string
	closer    chan struct{}

	PID         *expvar.Int
	Invoked     *expvar.String
	InvokedUnix *expvar.Int
	Uptime      *UptimeVar

	TimeTotal          *expvar.Float
	RequestErrorsTotal *expvar.Int
	RequestsTotal      *expvar.Int
	StatusTotal        *expvar.Map

	History History
}

// New creates a new middleware http stat recorder. Note that because httpstat
// uses expvar to track data, expvar variables can only be created ONCE, with
// the same namespace. If you know you are only using one invokation of
// httpstat, leave namespace blank. Otherwise use it to identify which thing
// it is recording (e.g. auth, frontend, etc).
//
// histOpts are options which you can use to enable history snapshots (see
// History.Elems(), and HistoryElem) which can be used to track historical
// records on how the request count and similar is increasing over time.
// History is disabled by default as it has an almost negligible performance
// hit. Make sure if History is being used, that HTTPStats.Close() is called
// when closing the server.
func New(namespace string, histOpts *HistoryOptions) *HTTPStats {
	if namespace != "" {
		namespace = strings.ToLower(strings.Trim(namespace, "_")) + "_"
	}

	s := &HTTPStats{
		namespace:   namespace,
		closer:      make(chan struct{}),
		PID:         expvar.NewInt("httpstat_" + namespace + "pid"),
		Invoked:     expvar.NewString("httpstat_" + namespace + "invoked"),
		InvokedUnix: expvar.NewInt("httpstat_" + namespace + "invoked_unix"),

		TimeTotal:          expvar.NewFloat("httpstat_" + namespace + "request_total_seconds"),
		RequestErrorsTotal: expvar.NewInt("httpstat_" + namespace + "request_error_total"),
		RequestsTotal:      expvar.NewInt("httpstat_" + namespace + "request_total"),
		StatusTotal:        expvar.NewMap("httpstat_" + namespace + "status_total"),
	}

	started := time.Now()

	s.PID.Set(int64(os.Getpid()))
	s.Invoked.Set(started.Format(time.RFC3339))
	s.InvokedUnix.Set(started.Unix())

	// Custom variables we need to publish ourselves.
	s.Uptime = &UptimeVar{started: started}
	expvar.Publish("httpstat_"+namespace+"invoked_seconds", s.Uptime)

	if histOpts == nil {
		histOpts = &HistoryOptions{Enabled: false}
	}

	// TODO: sanitize this a bit more.
	// TODO: have this lifecycle managed by the History type?
	if histOpts.Enabled {
		if histOpts.MaxResolution < 10*time.Second {
			histOpts.MaxResolution = 5 * time.Minute
		}
		if histOpts.Resolution < 5*time.Second {
			histOpts.Resolution = 5 * time.Second
		}

		s.History = History{Opts: *histOpts}

		// TODO: use history for averaging?
		go s.updateHistory()
	}

	return s
}

// Close is required if using History, it will close the goroutine which
// manages taking snapshots.
func (s *HTTPStats) Close() {
	close(s.closer)
}

func (s *HTTPStats) update(r ResponseWriter, dur time.Duration) {
	statusKey := strconv.FormatInt(int64(r.Status()), 10)

	s.TimeTotal.Add(dur.Seconds())
	s.RequestsTotal.Add(1)
	s.StatusTotal.Add(statusKey, 1)

	if r.Status() >= 500 {
		s.RequestErrorsTotal.Add(1)
	}
}

func (s *HTTPStats) updateHistory() {
	ticker := time.NewTicker(s.History.Opts.Resolution)

	for {
		select {
		case <-s.closer:
			return
		case <-ticker.C:
			s.History.add(s)
		}
	}
}

// MarshalJSON implements the json.Marshaler interface, allowing all httpstats
// expvars that match the configured namespace of the current HTTPStats, are
// returned in JSON form.
func (s *HTTPStats) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}

	fmt.Fprint(buf, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !strings.HasPrefix(kv.Key, "httpstat_"+s.namespace) {
			return
		}

		if !first {
			fmt.Fprint(buf, ",\n")
		}
		first = false
		fmt.Fprintf(buf, "%q: %s", kv.Key, kv.Value)
	})

	fmt.Fprint(buf, "\n}\n")

	return buf.Bytes(), nil
}

// Record is the handler wrapper method used to invoke tracking of all child
// handlers. Note that this should be invoked early in the handler chain,
// otherwise the handlers invoked before this, will not be recorded/tracked.
// Also note that if one of the children handlers writes to the ResponseWriter
// after the handler is returned (e.g. from a goroutine), the time and bytes
// written will not be updated after the handler returns.
func (s *HTTPStats) Record(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := NewResponseRecorder(w)
		start := time.Now()
		next.ServeHTTP(rr, r)
		s.update(rr, time.Since(start))
	})
}

// ServeHTTP is a way of invoking/showing the JSON version of httpstats without
// mounting an expvar endpoint (e.g. if you don't want all of the other expvar
// stats). If you mount /debug/vars via expvar, this isn't needed.
func (s *HTTPStats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// UptimeVar is a type which implements the expvar.Var interface, and shows
// the (or time since invokation) of the struct when calling String().
type UptimeVar struct {
	started time.Time
}

// String returns the amount of seconds that UptimeVar has recorded, in integer
// form.
func (u *UptimeVar) String() string {
	return fmt.Sprintf("%d", int(time.Since(u.started).Seconds()))
}
