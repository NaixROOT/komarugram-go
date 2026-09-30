// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"strconv"
	"strings"
	"time"

	"gio-mw/token"
	"gio-mw/widget/button"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// audioBarHeight is the height of the bar of what plays, under the chat's
// header and over its pinned bar, where Telegram Desktop has its own.
const audioBarHeight = unit.Dp(48)

// speedMinLength is the shortest music whose speed changes, as in Telegram
// Desktop (kMinLengthForChangeablePlaybackSpeed): a voice message's always
// does, a song's does not, a podcast's does.
const speedMinLength = 20 * time.Minute

// audioSpeeds are the speeds the bar's button goes through.
var audioSpeeds = []float64{1, 1.5, 2}

// speedChanges reports whether m plays at the speed chosen.
func speedChanges(m model.Message) bool {
	return m.Kind == model.MessageVoice || m.Media != nil && m.Media.Duration >= speedMinLength
}

// current is the message that plays or loads, what the player tells of it,
// and the speed chosen; false when there is none.
func (v *audioPlayer) current() (model.Message, audioState, float64, bool) {
	if v == nil {
		return model.Message{}, audioState{}, 1, false
	}
	v.mu.Lock()
	m := v.msg
	speed := v.speed
	v.mu.Unlock()
	if speed == 0 {
		speed = 1
	}
	state := v.state(m)
	return m, state, speed, state.active
}

// shownIn is the chat, or the thread, the message that plays is shown in.
func (v *audioPlayer) shownIn() int64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.list
}

// setSpeed chooses the speed of what changes its speed, 0 standing for 1,
// without keeping the choice: it comes from what was kept.
func (v *audioPlayer) setSpeed(speed float64) {
	if speed == 0 {
		speed = 1
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	v.speed = speed
	if v.stretch != nil && speedChanges(v.msg) {
		v.stretch.SetTempo(speed)
	}
}

// nextSpeed goes to the speed after the one chosen, and keeps the choice.
func (v *audioPlayer) nextSpeed() {
	v.mu.Lock()
	next := audioSpeeds[0]
	for i, s := range audioSpeeds {
		if s == max(v.speed, 1) {
			next = audioSpeeds[(i+1)%len(audioSpeeds)]
		}
	}
	save := v.saveSpeed
	v.mu.Unlock()
	v.setSpeed(next)
	if save != nil {
		save(next)
	}
}

// watch waits for m to play to its end, and plays what follows it then.
func (v *audioPlayer) watch(ctx context.Context, m model.Message, playback audioPlayback) {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		// Asking whether it plays is what notices that it ended.
		if !playback.Playing() && playback.Ended() {
			v.advance(m)
			return
		}
	}
}

// advance plays the voice message, or the music, that follows m where it
// is shown, as Telegram Desktop goes on through a chat's voice messages
// and through its music; after the last one the player has nothing.
func (v *audioPlayer) advance(m model.Message) {
	v.mu.Lock()
	if v.key != m.Key {
		v.mu.Unlock()
		return
	}
	page, list := v.page, v.list
	v.mu.Unlock()
	if next, ok := nextAudio(page.source.History(list).Messages, m); ok {
		v.toggleIn(page, list, next, -1)
	} else {
		v.mu.Lock()
		if v.key == m.Key {
			v.stopLocked()
		}
		v.mu.Unlock()
	}
	page.invalidate()
}

// nextAudio is the message after m among messages that is of m's kind, a
// voice message or music, if it plays in the client.
func nextAudio(messages []model.Message, m model.Message) (model.Message, bool) {
	after := false
	for _, next := range messages {
		switch {
		case next.Key == m.Key:
			after = true
		case after && next.Kind == m.Kind && next.Media != nil && !next.Deleted:
			return next, internalAudio(next)
		}
	}
	return model.Message{}, false
}

// audioBar is the bar of what the window's player plays, as a chat page
// shows it: play or pause, whose it is, the speed of what changes its
// speed, and a button that ends it. A click on it goes to the message, in
// the chat it is in.
type audioBar struct {
	bar                surface
	play, speed, close *button.Button
}

// audioBarSize is how much of the page's top the bar takes: nothing while
// nothing plays.
func (p *chatPage) audioBarSize(gtx layout.Context) int {
	if _, _, _, ok := p.audio.current(); !ok || p.audioExternal != nil && p.audioExternal() {
		return 0
	}
	return gtx.Dp(audioBarHeight)
}

// audioTitle is what the bar tells of m in two lines: a voice message's
// sender and when it was sent, as Telegram Desktop does; music's title and
// performer.
func audioTitle(m model.Message, now time.Time, l localization.Catalog) (string, string) {
	if m.Kind == model.MessageMusic {
		title := m.Media.Title
		if title == "" {
			title = m.Media.FileName
		}
		if title == "" {
			title = l.T("history.file")
		}
		return title, m.Media.Performer
	}
	name := m.SenderName
	if m.Outgoing {
		name = l.T("history.you")
	}
	if name == "" {
		name = l.T("history.voice")
	}
	clock := m.Date.Format("15:04")
	day := func(t time.Time) time.Time {
		y, mo, d := t.Date()
		return time.Date(y, mo, d, 0, 0, 0, 0, t.Location())
	}
	sent := day(m.Date.In(now.Location()))
	switch {
	case sent.Equal(day(now)):
		return name, strings.ReplaceAll(l.T("audio.today"), "{time}", clock)
	case sent.Equal(day(now).AddDate(0, 0, -1)):
		return name, strings.ReplaceAll(l.T("audio.yesterday"), "{time}", clock)
	}
	return name, strings.NewReplacer("{date}", m.Date.Format("02.01.2006"), "{time}", clock).Replace(l.T("audio.date"))
}

// speedText is a speed as its button tells it: 1×, 1.5×.
func speedText(speed float64) string {
	return strconv.FormatFloat(speed, 'g', 3, 64) + "×"
}

// layoutAudioBar draws the bar across the top of gtx.
func (p *chatPage) layoutAudioBar(gtx layout.Context, l localization.Catalog) {
	b := &p.audioBar
	if b.play == nil {
		b.play, b.speed, b.close = button.Text(), button.Text(), button.Text()
	}
	m, state, speed, ok := p.audio.current()
	if !ok {
		return
	}
	if b.close.Clicked(gtx) {
		p.audio.stop()
		return
	}
	if b.play.Clicked(gtx) {
		p.audio.toggle(p, m, -1)
		state = p.audio.state(m)
	}
	if b.speed.Clicked(gtx) {
		p.audio.nextSpeed()
		_, _, speed, _ = p.audio.current()
	}
	if b.bar.Clicked(gtx) && p.audio.shownIn() == p.chat {
		p.jumpTo(m.Key.MessageID)
	}
	if state.playing || state.loading {
		// The line of what was heard moves while it plays.
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(100 * time.Millisecond)})
	}

	sc := scheme(gtx)
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(audioBarHeight))
	fillRect(gtx, sc.Surface.Color, size)
	gtx.Constraints = layout.Exact(size)
	b.bar.Layout(gtx, size, surfaceStyle{content: sc.Surface.OnColor}, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
	button := gtx.Dp(48)
	// iconButton draws one of the bar's buttons in a square at x.
	iconButton := func(x int, w layout.Widget) {
		offset(gtx, image.Pt(x, (size.Y-button)/2), func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(button, button))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return w(gtx)
			})
		})
	}
	left := gtx.Dp(4)
	iconButton(left, func(gtx layout.Context) layout.Dimensions {
		if state.playing {
			return b.play.LayoutIconOnly(gtx, l.T("audio.pause"), iconPauseFile)
		}
		return b.play.LayoutIconOnly(gtx, l.T("audio.play"), iconPlayFile)
	})
	right := size.X - gtx.Dp(4) - button
	iconButton(right, func(gtx layout.Context) layout.Dimensions {
		return b.close.LayoutIconOnly(gtx, l.T("audio.close"), iconClear)
	})
	if speedChanges(m) {
		width := gtx.Dp(56)
		right -= width
		offset(gtx, image.Pt(right, 0), func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(width, size.Y))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return b.speed.Layout(gtx, speedText(speed))
			})
		})
	}
	title, subtitle := audioTitle(m, gtx.Now, l)
	textX := left + button + gtx.Dp(8)
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(max(right-textX-gtx.Dp(8), 0), size.Y)}
	offset(textGtx, image.Pt(textX, 0), func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.Y = size.Y
		return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, title, token.TypestyleLabelLargeEmphasized, sc.Surface.OnColor, 1)
				}),
				layout.Rigid(layout.Spacer{Width: 8}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, subtitle, token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 1)
				}),
			)
		})
	})
	// The line under the bar: what was heard, in the primary color.
	line := gtx.Dp(2)
	offset(gtx, image.Pt(0, size.Y-line), func(gtx layout.Context) layout.Dimensions {
		fillRect(gtx, sc.OutlineVariant, image.Pt(size.X, line))
		fillRect(gtx, sc.Primary.Color, image.Pt(int(state.progress*float32(size.X)), line))
		return layout.Dimensions{}
	})
}
