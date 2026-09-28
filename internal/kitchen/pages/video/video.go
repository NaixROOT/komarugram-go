// SPDX-License-Identifier: Unlicense OR MIT

package video

import (
	"image"
	"sync"

	"komarugram/pkg/video"

	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

const (
	// videoFile is resolved relative to the working directory the example is
	// started from.
	videoFile = "assets/video.mp4"

	// decodeHeight is the height ffmpeg scales frames down to before they are
	// piped in. It caps how much pixel data crosses the pipe every frame.
	decodeHeight = 720

	// viewHeight is how tall the video surface is laid out on the page.
	viewHeight = unit.Dp(360)
)

// The router builds a fresh page on every navigation, so the player is kept
// per package instead of per page: one ffmpeg process serves every visit, and
// it suspends itself while the page is not on screen.
var (
	playerOnce   sync.Once
	sharedPlayer *video.Player
	sharedErr    error
)

func loadPlayer() (*video.Player, error) {
	playerOnce.Do(func() {
		player, err := video.NewPlayer(videoFile, decodeHeight)
		if err != nil {
			sharedErr = err
			return
		}
		sharedPlayer = player
		player.Start()
	})
	return sharedPlayer, sharedErr
}

type Page struct {
	player    *video.Player
	playPause *button.Button
	playerErr error
}

func NewPage() router.PageWidget {
	p := &Page{
		playPause: button.Filled(),
	}
	player, err := loadPlayer()
	if err != nil {
		p.playerErr = err
		p.playPause.Disable()
		return p
	}
	p.player = player
	return p
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	if p.player == nil {
		return
	}
	if p.playPause.Clicked(gtx) {
		p.player.SetPaused(!p.player.Paused())
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Video"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "An mp4 file decoded by an ffmpeg subprocess"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionConfig),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionVideo),
		)
	})
}

func (p *Page) sectionConfig(gtx layout.Context) layout.Dimensions {
	label := "Pause"
	if p.player == nil || p.player.Paused() {
		label = "Play"
	}
	details := "ffmpeg is unavailable"
	if p.player != nil {
		details = p.player.Info().String()
	}

	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowWrap,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.playPause.Layout(gtx, label)
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return exp.BodyL(gtx, details)
		}),
	)
}

// sectionVideo paints the most recent frame, scaled to fit the page width
// while keeping its aspect ratio.
func (p *Page) sectionVideo(gtx layout.Context) layout.Dimensions {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(viewHeight)}

	if p.playerErr != nil {
		return p.message(gtx, size, p.playerErr.Error())
	}

	frame, err := p.player.Frame()
	if err != nil {
		return p.message(gtx, size, err.Error())
	}
	if !p.player.Paused() {
		// Ask for another frame for as long as the video is running. The page
		// stops doing so when it is no longer shown, and the player then
		// suspends ffmpeg on its own.
		gtx.Execute(op.InvalidateCmd{})
	}
	if frame == nil {
		return p.message(gtx, size, "Decoding…")
	}

	frameSize := frame.Bounds().Size()
	scale := min(
		float32(size.X)/float32(frameSize.X),
		float32(size.Y)/float32(frameSize.Y),
	)
	shown := image.Point{
		X: int(float32(frameSize.X) * scale),
		Y: int(float32(frameSize.Y) * scale),
	}
	offset := image.Point{X: (size.X - shown.X) / 2}

	defer op.Offset(offset).Push(gtx.Ops).Pop()
	defer clip.Rect{Max: shown}.Push(gtx.Ops).Pop()

	imageOp := paint.NewImageOp(frame)
	imageOp.Filter = paint.FilterLinear
	imageOp.Add(gtx.Ops)
	defer op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(scale, scale))).Push(gtx.Ops).Pop()
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Dimensions{Size: size}
}

// message fills the video surface with a placeholder and centers txt in it.
func (p *Page) message(gtx layout.Context, size image.Point, txt string) layout.Dimensions {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surface := clip.Rect{Max: size}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, materialTheme.Scheme.SurfaceContainerHighest.AsNRGBA())
	surface.Pop()

	gtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return exp.BodyL(gtx, txt)
	})
}
