package indicator

import (
	"math"
	"testing"
	"time"
)

func TestCircularArcGrowsShrinksAndRemainsContinuous(t *testing.T) {
	_, short := circularAngles(0)
	_, long := circularAngles(circularCycle / 2)
	if short >= long || long < math.Pi {
		t.Fatal("spinner never grows")
	}
	for elapsed := time.Duration(0); elapsed < 10*circularCycle; elapsed += time.Millisecond {
		start, sweep := circularAngles(elapsed)
		if math.IsNaN(start) || sweep < math.Pi/18-.00001 || sweep > 1.5*math.Pi+.00001 {
			t.Fatal("invalid arc", elapsed, start, sweep)
		}
	}
	for n := 1; n < 5; n++ {
		boundary := time.Duration(n) * circularCycle
		before, arcBefore := circularAngles(boundary - time.Nanosecond)
		after, arcAfter := circularAngles(boundary)
		if math.Abs(before-after) > 1e-5 || math.Abs(arcBefore-arcAfter) > 1e-5 {
			t.Fatal("spinner jumps at loop boundary")
		}
	}
}
