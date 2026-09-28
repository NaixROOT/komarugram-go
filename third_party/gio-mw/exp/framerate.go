// SPDX-License-Identifier: Unlicense OR MIT

package exp

import (
	"gioui.org/layout"
	"gioui.org/op"
	"time"
)

type FrameRateCounter struct {
	Rate  float64
	Times [240]time.Time
}

func (c *FrameRateCounter) Update(gtx layout.Context) {
	frameTime := gtx.Now
	copy(c.Times[1:], c.Times[:len(c.Times)-1])
	c.Times[0] = frameTime
	if oldestTime := c.Times[len(c.Times)-1]; !oldestTime.IsZero() {
		seconds := frameTime.Sub(oldestTime).Seconds()
		if seconds > 0 {
			c.Rate = float64(len(c.Times)) / seconds
		}
	}
	gtx.Execute(op.InvalidateCmd{})
}
