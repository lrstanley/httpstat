// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package httpstat

import (
	"sync"
	"time"
)

type HistoryOptions struct {
	Enabled       bool
	MaxResolution time.Duration
	Resolution    time.Duration
}

type HistoryElem struct {
	Born          time.Time
	TimeTotal     float64
	TimeDiff      float64
	RequestErrors int64
	RequestsTotal int64
	RequestsDiff  int64
	RPS           int64
}

type History struct {
	Opts  HistoryOptions
	mu    sync.RWMutex
	elems []HistoryElem
}

func (h *History) Elems() []HistoryElem {
	h.mu.RLock()
	elems := h.elems
	h.mu.RUnlock()

	return elems
}

func (h *History) add(stats *HTTPStats) {
	h.truncate()

	elem := HistoryElem{
		Born:          time.Now(),
		TimeTotal:     stats.TimeTotal.Value(),
		RequestErrors: stats.RequestErrorsTotal.Value(),
		RequestsTotal: stats.RequestsTotal.Value(),
	}

	h.mu.RLock()
	if len(h.elems) > 0 {
		elem.RequestsDiff = elem.RequestsTotal - h.elems[len(h.elems)-1].RequestsTotal
		elem.TimeDiff = elem.TimeTotal - h.elems[len(h.elems)-1].TimeTotal

		if elem.RequestsDiff > 0 {
			elem.RPS = elem.RequestsDiff / int64(h.Opts.Resolution.Seconds())
		}
	} else if elem.RequestsTotal > 0 {
		elem.RPS = elem.RequestsTotal / int64(h.Opts.Resolution.Seconds())
	}
	h.mu.RUnlock()

	h.mu.Lock()
	h.elems = append(h.elems, elem)
	h.mu.Unlock()
}

func (h *History) truncate() {
	h.mu.Lock()

	truncateTo := -1
	for i := 0; i < len(h.elems); i++ {
		if time.Since(h.elems[i].Born) > h.Opts.MaxResolution {
			truncateTo = i
			continue
		}

		break
	}

	if truncateTo > -1 {
		h.elems = append(h.elems[:0], h.elems[truncateTo+1:]...)
	}
	h.mu.Unlock()
}
