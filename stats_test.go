// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package httpstat

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "Hello Gopher! This is some example text!")
}

func BenchmarkRequestBaseline(b *testing.B) {
	s := httptest.NewServer(http.HandlerFunc(dummyHandler))
	defer s.Close()

	var resp *http.Response
	var err error
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		resp, err = http.Get(s.URL)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(ioutil.Discard, resp.Body)
	}
}

func BenchmarkResponseBaseline(b *testing.B) {
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dummyHandler)
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkRequestStats(b *testing.B) {
	ts := time.Now().Nanosecond()
	stats := New(strconv.Itoa(ts), nil)
	s := httptest.NewServer(stats.Record(http.HandlerFunc(dummyHandler)))
	defer s.Close()

	var resp *http.Response
	var err error
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		resp, err = http.Get(s.URL)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(ioutil.Discard, resp.Body)
	}
}

func BenchmarkResponseStats(b *testing.B) {
	ts := time.Now().Nanosecond()
	stats := New(strconv.Itoa(ts), nil)

	rr := httptest.NewRecorder()
	handler := stats.Record(http.HandlerFunc(dummyHandler))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkRequestStatsWithHistory(b *testing.B) {
	ts := time.Now().Nanosecond()
	stats := New(strconv.Itoa(ts), &HistoryOptions{Enabled: true, Resolution: 10 * time.Second})
	s := httptest.NewServer(stats.Record(http.HandlerFunc(dummyHandler)))
	defer s.Close()

	var resp *http.Response
	var err error
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		resp, err = http.Get(s.URL)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(ioutil.Discard, resp.Body)
	}
}

func BenchmarkResponseStatsWithHistory(b *testing.B) {
	ts := time.Now().Nanosecond()
	stats := New(strconv.Itoa(ts), &HistoryOptions{Enabled: true, Resolution: 10 * time.Second})

	rr := httptest.NewRecorder()
	handler := stats.Record(http.HandlerFunc(dummyHandler))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}
