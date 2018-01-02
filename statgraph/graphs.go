// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package statgraph

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lrstanley/httpstat"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type renderer struct {
	stats *httpstat.HTTPStats
	mux   *http.ServeMux
}

// New returns a new http handler which allows viewing of httpstat data using
// png/svg rendered graphs. Note that if history isn't enabled for the
// requested HTTPStats, New will panic. History must be enabled.
//
// The following endpoints are registered with the return handler:
//   /{requests,rps,latency}
//   /{requests,rps,latency}.{svg,png}
//
// For example the following returns the average latency in svg form:
//   /latency.svg
//
// When viewing the main registered endpoint (/), you can view all three
// SVG versions of the graphs, which get updated automatically every
// History.Resolution.
func New(stats *httpstat.HTTPStats) http.Handler {
	if !stats.History.Opts.Enabled {
		panic("cannot create graph handler: requested HTTPStats has history disabled")
	}

	rn := &renderer{stats: stats, mux: http.NewServeMux()}
	rn.mux.HandleFunc("/requests", rn.requests)
	rn.mux.HandleFunc("/requests.svg", rn.requests)
	rn.mux.HandleFunc("/requests.png", rn.requests)
	rn.mux.HandleFunc("/rps", rn.requestsPerSecond)
	rn.mux.HandleFunc("/rps.svg", rn.requestsPerSecond)
	rn.mux.HandleFunc("/rps.png", rn.requestsPerSecond)
	rn.mux.HandleFunc("/latency", rn.latency)
	rn.mux.HandleFunc("/latency.svg", rn.latency)
	rn.mux.HandleFunc("/latency.png", rn.latency)
	rn.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprintf(w, htmlTemplate, int(stats.History.Opts.Resolution.Seconds())*1000)
	})
	return rn
}

func (rn *renderer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rn.mux.ServeHTTP(w, r)
}

func (rn *renderer) latency(w http.ResponseWriter, r *http.Request) {
	elems := rn.stats.History.Elems()
	spark := wantsSpark(r)

	reqTime := []time.Time{}
	reqLatency := []float64{}
	for i := 0; i < len(elems); i++ {
		reqTime = append(reqTime, elems[i].Born)
		diff := elems[i].TimeDiff / float64(elems[i].RequestsDiff)
		if math.IsNaN(diff) {
			diff = 0
		}
		reqLatency = append(reqLatency, diff)
	}

	ts := chart.TimeSeries{
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(2),
			FillColor:   chart.GetAlternateColor(3),
		},
		XValues: reqTime,
		YValues: reqLatency,
	}

	if spark {
		ts.Style.FillColor = drawing.ColorTransparent
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:           "time",
			NameStyle:      chart.Style{Show: !spark},
			Style:          chart.Style{Show: !spark},
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04:05"),
		},
		YAxis: chart.YAxis{
			Name:      "req time",
			NameStyle: chart.Style{Show: !spark},
			Style:     chart.Style{Show: !spark},
			ValueFormatter: func(v interface{}) string {
				f := v.(float64)

				dur, err := time.ParseDuration(fmt.Sprintf("%fs", f))
				if err != nil {
					panic(fmt.Sprintf("attempted to parse '%fs' into time.Duration: %s", f, err))
				}

				return dur.String()
			},
		},
		Series: []chart.Series{ts},
	}

	if spark {
		graph.Background.FillColor = drawing.ColorTransparent
		graph.Canvas.FillColor = drawing.ColorTransparent
	}

	renderGraph(w, r, graph)
}

func (rn *renderer) requestsPerSecond(w http.ResponseWriter, r *http.Request) {
	elems := rn.stats.History.Elems()
	spark := wantsSpark(r)

	reqTime := []time.Time{}
	reqDiff := []float64{}
	for i := 0; i < len(elems); i++ {
		reqTime = append(reqTime, elems[i].Born)
		reqDiff = append(reqDiff, float64(elems[i].RPS))
	}

	ts := chart.TimeSeries{
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(0),
			FillColor:   chart.GetAlternateColor(0),
		},
		XValues: reqTime,
		YValues: reqDiff,
	}

	if spark {
		ts.Style.FillColor = drawing.ColorTransparent
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:           "time",
			NameStyle:      chart.Style{Show: !spark},
			Style:          chart.Style{Show: !spark},
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04:05"),
		},
		YAxis: chart.YAxis{
			Name:           "req/sec",
			NameStyle:      chart.Style{Show: !spark},
			Style:          chart.Style{Show: !spark},
			ValueFormatter: func(v interface{}) string { return chart.FloatValueFormatterWithFormat(v, "%.0f") },
		},
		Series: []chart.Series{ts},
	}

	if spark {
		graph.Background.FillColor = drawing.ColorTransparent
		graph.Canvas.FillColor = drawing.ColorTransparent
	}

	renderGraph(w, r, graph)
}

func (rn *renderer) requests(w http.ResponseWriter, r *http.Request) {
	elems := rn.stats.History.Elems()
	spark := wantsSpark(r)

	reqTime := []time.Time{}
	reqDiff := []float64{}
	for i := 0; i < len(elems); i++ {
		reqTime = append(reqTime, elems[i].Born)
		reqDiff = append(reqDiff, float64(elems[i].RequestsDiff))
	}

	ts := chart.TimeSeries{
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(5),
			FillColor:   chart.GetAlternateColor(5),
		},
		XValues: reqTime,
		YValues: reqDiff,
	}

	if spark {
		ts.Style.FillColor = drawing.ColorTransparent
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:           "time",
			NameStyle:      chart.Style{Show: !spark},
			Style:          chart.Style{Show: !spark},
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04:05"),
		},
		YAxis: chart.YAxis{
			Name:           "req/ingest",
			NameStyle:      chart.Style{Show: !spark},
			Style:          chart.Style{Show: !spark},
			ValueFormatter: func(v interface{}) string { return chart.FloatValueFormatterWithFormat(v, "%.0f") },
		},
		Series: []chart.Series{ts},
	}

	if spark {
		graph.Background.FillColor = drawing.ColorTransparent
		graph.Canvas.FillColor = drawing.ColorTransparent
	}

	renderGraph(w, r, graph)
}

func renderGraph(w http.ResponseWriter, r *http.Request, graph chart.Chart) {
	graph.Width, graph.Height = getDimensions(r)

	if strings.HasSuffix(strings.ToLower(r.URL.Path), ".svg") {
		w.Header().Set("Content-Type", "image/svg+xml")
		_ = graph.Render(chart.SVG, w)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	_ = graph.Render(chart.PNG, w)
}

func getDimensions(r *http.Request) (width, height int) {
	w, _ := strconv.ParseInt(r.FormValue("w"), 10, 64)
	if w == 0 {
		w, _ = strconv.ParseInt(r.FormValue("width"), 10, 64)
	}
	h, _ := strconv.ParseInt(r.FormValue("h"), 10, 64)
	if h == 0 {
		h, _ = strconv.ParseInt(r.FormValue("height"), 10, 64)
	}

	if w == 0 && h == 0 {
		w = 1024
		h = 400
	}

	if w < 256 {
		w = 256
	} else if w > 2048 {
		w = 2048
	}

	if h < 100 {
		h = 100
	} else if h > 800 {
		h = 800
	}

	return int(w), int(h)
}

func wantsSpark(r *http.Request) bool {
	wants, _ := strconv.ParseBool(r.FormValue("spark"))
	return wants
}
