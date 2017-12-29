// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package httpstat

import (
	"fmt"
	"time"
)

// UptimeVar is a type which implements the expvar.Var interface, and shows
// the (or time since invokation) of the struct when calling String().
type UptimeVar struct {
	started time.Time
}

func (u *UptimeVar) String() string {
	return fmt.Sprintf("%d", int(time.Since(u.started).Seconds()))
}
