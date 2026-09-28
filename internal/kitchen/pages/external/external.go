// SPDX-License-Identifier: Unlicense OR MIT

package external

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"komarugram/pkg/player"

	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"

	"gioui.org/layout"
	"gioui.org/op"
)

// videoDir holds the files offered to the external player.
const videoDir = "videos"

type entry struct {
	path      string
	name      string
	openFile  *button.Button
	openRange *button.Button
}

type Page struct {
	entries []entry
	loadErr error

	// kind is the first installed player, "" if there is none.
	kind    player.Kind
	current player.Player
	stream  *player.Stream
	playing string
	viaHTTP bool
	lastErr error

	playPause *button.Button
	back      *button.Button
	forward   *button.Button
	stop      *button.Button
}

func NewPage() router.PageWidget {
	p := &Page{
		playPause: button.Text(),
		back:      button.Text(),
		forward:   button.Text(),
		stop:      button.Text(),
	}
	if installed := player.Installed(nil); len(installed) > 0 {
		p.kind = installed[0]
	}
	paths, err := filepath.Glob(filepath.Join(videoDir, "*.mp4"))
	if err != nil || len(paths) == 0 {
		p.loadErr = fmt.Errorf("no *.mp4 in %s/", videoDir)
		return p
	}
	sort.Strings(paths)
	for _, path := range paths {
		p.entries = append(p.entries, entry{
			path:      path,
			name:      filepath.Base(path),
			openFile:  button.Filled(),
			openRange: button.Outlined(),
		})
	}
	return p
}

func (p *Page) IsWide() bool {
	return true
}

func (p *Page) Update(gtx layout.Context) {
	for i := range p.entries {
		e := &p.entries[i]
		if e.openFile.Clicked(gtx) {
			p.open(e.path, e.name, false)
		}
		if e.openRange.Clicked(gtx) {
			p.open(e.path, e.name, true)
		}
	}
	if p.current == nil {
		return
	}
	if p.playPause.Clicked(gtx) {
		p.fail(p.current.TogglePause())
	}
	if p.back.Clicked(gtx) {
		p.fail(p.current.Seek(-10 * time.Second))
	}
	if p.forward.Clicked(gtx) {
		p.fail(p.current.Seek(10 * time.Second))
	}
	if p.stop.Clicked(gtx) {
		p.closePlayer()
	}
}

// open starts the player on the file, either directly or through a loopback stream
// that answers byte ranges the way a partially downloaded file would.
func (p *Page) open(path, name string, viaHTTP bool) {
	p.closePlayer()
	p.lastErr = nil

	source := path
	if viaHTTP {
		file, err := os.Open(path)
		if err != nil {
			p.lastErr = err
			return
		}
		info, err := file.Stat()
		if err != nil {
			p.lastErr = err
			_ = file.Close()
			return
		}
		stream, err := player.Serve(name, file, info.Size())
		if err != nil {
			p.lastErr = err
			_ = file.Close()
			return
		}
		p.stream = stream
		source = stream.URL()
	}

	current, err := player.Open(context.Background(), p.kind, "", source)
	if err != nil {
		p.lastErr = err
		p.closeStream()
		return
	}
	p.current = current
	p.playing = name
	p.viaHTTP = viaHTTP
}

func (p *Page) fail(err error) {
	if err != nil {
		p.lastErr = err
	}
}

func (p *Page) closePlayer() {
	if p.current != nil {
		_ = p.current.Close()
		p.current = nil
	}
	p.closeStream()
	p.playing = ""
}

func (p *Page) closeStream() {
	if p.stream != nil {
		_ = p.stream.Close()
		p.stream = nil
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "External player"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "mpv or VLC in its own window, driven over its IPC socket"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionStatus),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionControls),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionFiles),
		)
	})
}

func (p *Page) sectionStatus(gtx layout.Context) layout.Dimensions {
	return exp.BodyL(gtx, p.statusText(gtx))
}

func (p *Page) statusText(gtx layout.Context) string {
	switch {
	case p.kind == "":
		return "neither mpv nor VLC is installed"
	case p.loadErr != nil:
		return p.loadErr.Error()
	case p.lastErr != nil:
		return "error: " + p.lastErr.Error()
	case p.current == nil:
		return "no external window open"
	}

	status := p.current.Status()
	if !status.Running {
		p.closePlayer()
		return "the external window was closed"
	}
	// The player reports its state over the socket; keep repainting to show it.
	gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(200 * time.Millisecond)})

	state := "playing"
	if status.Paused {
		state = "paused"
	}
	source := "local file"
	if p.viaHTTP {
		source = fmt.Sprintf("loopback stream, %d range request(s)", p.stream.Requests())
	}
	return fmt.Sprintf("%s · %s · %s · %s / %s · %s",
		p.kind.Title(), p.playing, state, clock(status.Position), clock(status.Duration), source)
}

func (p *Page) sectionControls(gtx layout.Context) layout.Dimensions {
	if p.current == nil {
		gtx = gtx.Disabled()
	}
	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowWrap,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.playPause.Layout(gtx, "Play / Pause")
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.back.Layout(gtx, "-10 s")
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.forward.Layout(gtx, "+10 s")
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.stop.Layout(gtx, "Close window")
		}),
	)
}

func (p *Page) sectionFiles(gtx layout.Context) layout.Dimensions {
	segments := make([]block.Segment, 0, len(p.entries)*2)
	for i := range p.entries {
		e := &p.entries[i]
		segments = append(segments, block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return e.openFile.Layout(gtx, e.name)
				}),
				block.NewHorizontalSpacer(examples.SpacingSmall),
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return e.openRange.Layout(gtx, "stream "+e.name)
				}),
			)
		}))
		segments = append(segments, block.NewVerticalSpacer(examples.SpacingSmall))
	}
	return block.Line{Axis: block.AxisVertical, Overflow: block.OverflowClip}.Layout(gtx, segments...)
}

func clock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}
