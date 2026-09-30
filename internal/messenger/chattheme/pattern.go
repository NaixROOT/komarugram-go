// SPDX-License-Identifier: Unlicense OR MIT

package chattheme

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"sync"
)

// doodles are line drawings in a 24×24 box, stroked as Telegram's pattern
// wallpapers are drawn: small objects scattered over the chat.
var doodles = []string{
	// A heart.
	"M12 19.5C5 14.5 3.5 10 5.5 7.2C7.5 4.6 10.8 5.2 12 8C13.2 5.2 16.5 4.6 18.5 7.2C20.5 10 19 14.5 12 19.5Z",
	// A star.
	"M12 3.5L14.4 9L20.3 9.4L15.8 13.2L17.2 19L12 15.9L6.8 19L8.2 13.2L3.7 9.4L9.6 9Z",
	// A ring.
	"M12 5A7 7 0 1 1 11.99 5Z",
	// A wave.
	"M2.5 12C4.5 8 7 8 9 12S13.5 16 15.5 12S20 8 21.5 10",
	// A plus.
	"M12 6V18M6 12H18",
	// A paper plane.
	"M3.5 11.5L20.5 4.5L16.5 19.5L11 15ZM11 15L20.5 4.5M11 15V20L13.5 17",
	// A crescent.
	"M15 4.5A8 8 0 1 0 19.5 16A6.5 6.5 0 0 1 15 4.5Z",
	// A note.
	"M9 17.5V6L18.5 4V15.5M9 17.5A2.2 2.2 0 1 1 8.99 17.5M18.5 15.5A2.2 2.2 0 1 1 18.49 15.5",
	// A flower.
	"M12 9.5A2.5 2.5 0 1 1 11.99 9.5ZM12 9.5C10 7 11 4 12 4S14 7 12 9.5M14.5 12C17 10 20 11 20 12S17 14 14.5 12M12 14.5C14 17 13 20 12 20S10 17 12 14.5M9.5 12C7 14 4 13 4 12S7 10 9.5 12",
	// A diamond.
	"M12 3.5L18.5 12L12 20.5L5.5 12Z",
	// A cloud.
	"M7 18H17C20 18 20.8 14.2 18 13.2C18 9.5 13.5 8.2 12 11C10.2 9.2 6.5 10.2 7 13.2C4 13.5 4.2 18 7 18Z",
	// A lightning bolt.
	"M13 3.5L6.5 13H11.5L10.5 20.5L17.5 10.5H12.5Z",
	// A drop.
	"M12 3.5C15.5 8.5 17.5 11.5 17.5 14.8C17.5 18 15 20.5 12 20.5S6.5 18 6.5 14.8C6.5 11.5 8.5 8.5 12 3.5Z",
	// A smile.
	"M12 4.5A7.5 7.5 0 1 1 11.99 4.5ZM8.5 13.5C10 16 14 16 15.5 13.5M9.5 9.8V10.2M14.5 9.8V10.2",
	// An envelope.
	"M4 7H20V17.5H4ZM4 7L12 13L20 7",
	// A leaf.
	"M5 19C5 10 10 5 19 5C19 14 14 19 5 19ZM5 19L13 11",
	// A spiral.
	"M12 12C12 11 13.5 11 13.5 12.2C13.5 14 11 14.5 10 13C8.5 10.8 10.5 8.5 13 8.8C16.5 9.3 17 13.5 15 15.8C12.5 18.5 7.5 17.5 6.5 13.5",
	// A cup.
	"M6 9H16V14C16 17 14 19 11 19S6 17 6 14ZM16 10.5H17.5C19.5 10.5 19.5 14 17.5 14H16M9 4.5C8.5 5.5 10 6 9.5 7M12.5 4.5C12 5.5 13.5 6 13 7",
}

// patternWidth and patternHeight are the pattern's box, portrait as the
// wallpapers of Telegram's patterns are.
const patternWidth, patternHeight = 1440, 2560

var doodlePattern = sync.OnceValue(func() []byte {
	// The same pattern every time: a fixed seed.
	rnd := rand.New(rand.NewPCG(0x6b6f6d617275, 0x6772616d))
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d">`, patternWidth, patternHeight)
	b.WriteString(`<g fill="none" stroke="#000000" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">`)
	const cell = 104
	rows := patternHeight / cell
	cols := patternWidth / cell
	for y := 0; y <= rows; y++ {
		for x := 0; x <= cols; x++ {
			// Every other row is shifted, so that the doodles do not line up.
			cx := float64(x*cell) + float64(cell)*.5*float64(y%2) + (rnd.Float64()-.5)*cell*.45
			cy := float64(y*cell) + (rnd.Float64()-.5)*cell*.45
			scale := 2.7 + rnd.Float64()*1.2
			angle := math.Round((rnd.Float64() - .5) * 80)
			d := doodles[rnd.IntN(len(doodles))]
			fmt.Fprintf(&b, `<path transform="translate(%.1f %.1f) rotate(%.0f) scale(%.2f %.2f) translate(-12 -12)" d="%s"/>`, cx, cy, angle, scale, scale, d)
		}
	}
	b.WriteString(`</g></svg>`)
	return []byte(b.String())
})

// DoodlePattern is the SVG of a pattern of doodles, for a wallpaper's
// Image with Pattern set: the preset Classic's and the demo's.
func DoodlePattern() []byte { return doodlePattern() }
