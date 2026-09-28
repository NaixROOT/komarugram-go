// SPDX-License-Identifier: Unlicense OR MIT

package video_grid

import (
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"komarugram/pkg/video"

	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

const (
	// videoDir holds the clips shown in the grid, resolved relative to the
	// working directory the example is started from.
	videoDir = "videos"

	// tileHeight is the decode height of a tile: everything in the cache is
	// stored at this size, so it also sets the memory cost per frame.
	tileHeight = 240

	// tileFPS is the frame rate the cache is resampled to. Tiles in a grid
	// are small, and halving the rate halves the memory.
	tileFPS = 15

	// maxFrames caps a clip at six seconds of loop, the way a preview in a
	// chat grid is capped. Without it a long video would dominate the cache.
	maxFrames = 90

	tileGap = unit.Dp(8)
)

// The cache lives per package: the router rebuilds pages on every navigation,
// and decoding is far too expensive to repeat. Reloading it in another storage
// format is explicit, through the button on the page.
var cache videoCache

type videoCache struct {
	mu      sync.Mutex
	clips   []*video.Clip
	format  video.Format
	err     error
	status  string
	took    time.Duration
	loading bool
	loaded  bool
}

// load decodes every clip in videoDir in the background, replacing whatever
// the cache held before.
func (c *videoCache) load(format video.Format) {
	c.mu.Lock()
	if c.loading {
		c.mu.Unlock()
		return
	}
	c.loading = true
	c.loaded = false
	c.clips = nil
	c.err = nil
	c.format = format
	c.status = "Looking for clips…"
	c.mu.Unlock()

	go func() {
		paths, err := filepath.Glob(filepath.Join(videoDir, "*.mp4"))
		if err != nil || len(paths) == 0 {
			c.finish(nil, 0, fmt.Errorf("no *.mp4 found in %s/", videoDir))
			return
		}
		sort.Strings(paths)

		started := time.Now()
		var clips []*video.Clip
		for i, path := range paths {
			c.setStatus(fmt.Sprintf("Decoding %d/%d as %v: %s", i+1, len(paths), format, filepath.Base(path)))
			clip, err := video.DecodeClip(path, tileHeight, tileFPS, maxFrames, format)
			if err != nil {
				c.finish(nil, 0, err)
				return
			}
			clip.Name = filepath.Base(path)
			clips = append(clips, clip)
		}
		c.finish(clips, time.Since(started), nil)
	}()
}

func (c *videoCache) setStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = status
}

func (c *videoCache) finish(clips []*video.Clip, took time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clips, c.took, c.err = clips, took, err
	c.loading, c.loaded = false, true
}

// snapshot returns what the layout needs without holding the lock while it
// paints.
func (c *videoCache) snapshot() (clips []*video.Clip, ready bool, summary string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.loaded {
		return nil, false, c.status
	}
	if c.err != nil {
		return nil, true, c.err.Error()
	}
	var bytes, frames int
	for _, clip := range c.clips {
		bytes += clip.Bytes()
		frames += clip.Len()
	}
	return c.clips, true, fmt.Sprintf("%d clips · %d frames · %v · %.0f MB cached in %.1f s",
		len(c.clips), frames, c.format, float64(bytes)/1e6, c.took.Seconds())
}

// otherFormat is the storage format the toggle switches to.
func (c *videoCache) otherFormat() video.Format {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.format == video.FormatRGBA {
		return video.FormatYCbCr
	}
	return video.FormatRGBA
}

type Page struct {
	started time.Time
	restart *button.Button
	swap    *button.Button
}

func NewPage() router.PageWidget {
	cache.mu.Lock()
	fresh := !cache.loaded && !cache.loading
	cache.mu.Unlock()
	if fresh {
		cache.load(video.FormatRGBA)
	}
	return &Page{
		started: time.Now(),
		restart: button.Text(),
		swap:    button.Text(),
	}
}

func (p *Page) IsWide() bool {
	return true
}

func (p *Page) Update(gtx layout.Context) {
	if p.restart.Clicked(gtx) {
		p.started = gtx.Now
	}
	if p.swap.Clicked(gtx) {
		cache.load(cache.otherFormat())
		p.started = gtx.Now
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Video grid"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Clips decoded once into memory and looped without a running ffmpeg"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionStatus),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionGrid),
		)
	})
}

func (p *Page) sectionStatus(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowWrap,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.restart.Layout(gtx, "Restart loops")
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.swap.Layout(gtx, "Store as "+cache.otherFormat().String())
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			_, _, summary := cache.snapshot()
			return exp.BodyL(gtx, summary)
		}),
	)
}

func (p *Page) sectionGrid(gtx layout.Context) layout.Dimensions {
	clips, ready, _ := cache.snapshot()
	if !ready {
		// Decoding runs in the background; keep repainting until it lands.
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(100 * time.Millisecond)})
		return layout.Dimensions{}
	}
	if len(clips) == 0 {
		return layout.Dimensions{}
	}

	elapsed := gtx.Now.Sub(p.started)
	gap := gtx.Dp(tileGap)
	var (
		pos     image.Point
		rowMax  int
		nextDue = time.Duration(1<<62 - 1)
	)
	for _, clip := range clips {
		size := clip.Size
		if pos.X > 0 && pos.X+size.X > gtx.Constraints.Max.X {
			pos = image.Point{Y: pos.Y + rowMax + gap}
			rowMax = 0
		}
		p.paintTile(gtx, clip, pos, elapsed)
		nextDue = min(nextDue, untilNextFrame(clip, elapsed))
		pos.X += size.X + gap
		rowMax = max(rowMax, size.Y)
	}

	// Repaint when the next tile is due for a new frame, not at display rate:
	// the clips run at 15 fps, so most display frames would be identical.
	gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(nextDue)})

	return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: pos.Y + rowMax}}
}

func (p *Page) paintTile(gtx layout.Context, c *video.Clip, at image.Point, elapsed time.Duration) {
	defer op.Offset(at).Push(gtx.Ops).Pop()

	materialTheme := wdk.GetMaterialTheme(gtx)
	shape := wdk.Box{
		Shape:    wdk.UniformCornerShapes(wdk.CornerShape{Kind: wdk.CornerKindRound, Size: 12}),
		EndPoint: c.Size,
	}
	defer shape.Outline(gtx).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, materialTheme.Scheme.SurfaceContainerHighest.AsNRGBA())

	frame := c.FrameAt(elapsed)
	imageOp := paint.NewImageOp(frame)
	imageOp.Filter = paint.FilterLinear
	imageOp.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

// untilNextFrame is how long the clip keeps showing its current frame.
func untilNextFrame(c *video.Clip, elapsed time.Duration) time.Duration {
	period := time.Duration(float64(time.Second) / c.FPS)
	if period <= 0 {
		return time.Millisecond
	}
	return period - elapsed%period
}
