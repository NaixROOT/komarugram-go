// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"gio-mw/token"
	"gio-mw/widget/scroll"

	"gioui.org/gpu/headless"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// selectionBar holds the buttons over selected messages, as Telegram
// Desktop's top bar has them: forward, delete and snapshot, each with the
// number of messages, and cancel at the end. A button shows only when its
// action is allowed for every selected message.
type selectionBar struct {
	forward, remove, snapshot, cancel surface
}

// selectedParts are the selected messages in order, with the parts of an
// album as messages of their own, as Telegram counts them.
func (p *chatPage) selectedParts() []model.Message {
	var out []model.Message
	for _, m := range p.messages {
		if p.selection.selected[m.Key.MessageID] {
			out = append(out, messageParts(m)...)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key.MessageID < out[j].Key.MessageID })
	return out
}

// messageParts are the messages m shows: the parts of an album, or m.
func messageParts(m model.Message) []model.Message {
	if len(m.Attachments) > 0 {
		return m.Attachments
	}
	return []model.Message{m}
}

// selectionRights tells what may be done with the selected messages.
func (p *chatPage) selectionRights() model.MessageRights {
	return p.rightsFor(p.selectedParts())
}

// rightsFor tells what may be done with parts, messages of the open chat.
func (p *chatPage) rightsFor(parts []model.Message) model.MessageRights {
	var r model.MessageRights
	if source, ok := p.source.(model.RightsSource); ok {
		r = source.MessageRights(p.chat, parts)
	} else {
		// Without a store that knows the rights, nothing is protected, and
		// channels delete for everyone.
		r = model.MessageRights{Forward: true, Save: true, Delete: true, Revoke: true, Everyone: p.chat <= -1000000000000}
		for _, m := range parts {
			if m.Kind == model.MessageService || m.NoForwards {
				r.Forward = false
			}
			if m.NoForwards {
				r.Save = false
			}
		}
	}
	if _, ok := p.source.(model.MessageForwarder); !ok {
		r.Forward = false
	}
	if _, ok := p.source.(model.MessageDeleter); !ok {
		r.Delete = false
	}
	if p.frozen.Frozen() {
		// A frozen account can only read.
		r.Forward, r.Delete = false, false
	}
	return r
}

// selectionHeader draws the selection's buttons over the chat's header.
func (p *chatPage) selectionHeader(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	b := &p.actions
	rights := p.selectionRights()
	if b.cancel.Clicked(gtx) {
		p.clearSelection()
		gtx.Execute(op.InvalidateCmd{})
	}
	if b.forward.Clicked(gtx) && rights.Forward {
		p.openForward(gtx)
	}
	if b.remove.Clicked(gtx) && rights.Delete {
		p.openDelete(gtx, rights)
	}
	if b.snapshot.Clicked(gtx) && rights.Save {
		p.snapshotDue = true
		gtx.Execute(op.InvalidateCmd{})
	}
	count := p.selectionCount()
	size := gtx.Constraints.Max
	gap := layout.Rigid(layout.Spacer{Width: 8}.Layout)
	var buttons []layout.FlexChild
	add := func(s *surface, key string) {
		if len(buttons) > 0 {
			buttons = append(buttons, gap)
		}
		buttons = append(buttons, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return tonalButton(gtx, s, l.T(key), count)
		}))
	}
	if rights.Forward {
		add(&b.forward, "history.select_forward")
	}
	if rights.Delete {
		add(&b.remove, "history.select_delete")
	}
	if rights.Save {
		add(&b.snapshot, "history.select_snapshot")
	}
	// Cancel sits at the end, the actions at the start, both in the middle
	// of the header's height.
	measure := func(w layout.Widget) (layout.Dimensions, op.CallOp) {
		macro := op.Record(gtx.Ops)
		dims := w(gtx)
		return dims, macro.Stop()
	}
	gtx.Constraints.Min = image.Point{}
	cancel, cancelCall := measure(func(gtx layout.Context) layout.Dimensions {
		return textButton(gtx, &b.cancel, l.T("history.cancel"))
	})
	end := size.X - gtx.Dp(8) - cancel.Size.X
	offset(gtx, image.Pt(end, (size.Y-cancel.Size.Y)/2), func(gtx layout.Context) layout.Dimensions {
		cancelCall.Add(gtx.Ops)
		return cancel
	})
	start := gtx.Dp(12)
	area := gtx
	area.Constraints = layout.Constraints{Max: image.Pt(max(end-start-gtx.Dp(8), 0), size.Y)}
	actions, actionsCall := measure(func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints = area.Constraints
		if len(buttons) == 0 {
			// Nothing may be done with this selection: it only counts.
			return label(gtx, l.Count("history.selection_count", count, nil), token.TypestyleTitleSmall, scheme(gtx).Surface.OnColor, 1)
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, buttons...)
	})
	func() {
		defer clip.Rect{Max: image.Pt(end-gtx.Dp(8), size.Y)}.Push(gtx.Ops).Pop()
		offset(gtx, image.Pt(start, (size.Y-actions.Size.Y)/2), func(gtx layout.Context) layout.Dimensions {
			actionsCall.Add(gtx.Ops)
			return actions
		})
	}()
	return layout.Dimensions{Size: size}
}

// forwardPicker is the dialog that picks the chat to forward messages to.
type forwardPicker struct {
	modal   modal
	search  widget.Editor
	list    scroll.List
	rows    map[int64]*surface
	cancel  surface
	busy    bool
	err     error
	results chan error
	from    int64
	ids     []model.MessageID
	target  model.Chat
	loader  loadingIndicator
	// selection is set when the selected messages are forwarded, which
	// are unselected once they are.
	selection bool
}

func (p *chatPage) openForward(gtx layout.Context) {
	p.openForwardParts(gtx, p.selectedParts(), true)
}

// openForwardParts opens the picker to forward parts, messages of the open
// chat; selection tells that they are the selected ones.
func (p *chatPage) openForwardParts(gtx layout.Context, parts []model.Message, selection bool) {
	f := &p.forwarding
	if f.busy {
		return
	}
	f.ids = f.ids[:0]
	for _, m := range parts {
		f.ids = append(f.ids, m.Key.MessageID)
	}
	f.selection = selection
	if len(f.ids) == 0 {
		return
	}
	f.from, f.err = p.chat, nil
	f.search.SingleLine = true
	f.search.SetText("")
	f.list = scroll.List{List: layout.List{Axis: layout.Vertical}}
	if f.rows == nil {
		f.rows = map[int64]*surface{}
	}
	f.modal.Open()
	gtx.Execute(key.FocusCmd{Tag: &f.search})
	gtx.Execute(op.InvalidateCmd{})
}

// forwardTargets are the chats messages may be forwarded to, matching the
// search, Saved Messages first.
func (p *chatPage) forwardTargets() []model.Chat {
	if p.chats == nil {
		return nil
	}
	rights, _ := p.source.(model.RightsSource)
	query := strings.ToLower(strings.TrimSpace(p.forwarding.search.Text()))
	var saved, out []model.Chat
	for _, c := range p.chats() {
		if rights != nil && !rights.CanSend(c.ID) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(c.Title), query) {
			continue
		}
		if c.Kind == model.KindSaved {
			saved = append(saved, c)
		} else {
			out = append(out, c)
		}
	}
	return append(saved, out...)
}

// updateForward takes the result of a forward.
func (p *chatPage) updateForward(l localization.Catalog) {
	f := &p.forwarding
	if f.results == nil {
		return
	}
	select {
	case err := <-f.results:
		f.busy, f.err, f.results = false, err, nil
		if err == nil {
			f.modal.Close()
			if f.from == p.chat && f.selection {
				p.clearSelection()
			}
			p.selectionNotice = l.Format("history.forwarded", map[string]string{"chat": f.target.Title})
		}
	default:
	}
}

func (p *chatPage) forward(target model.Chat) {
	f := &p.forwarding
	source, ok := p.source.(model.MessageForwarder)
	if !ok || f.busy {
		return
	}
	f.busy, f.err, f.target = true, nil, target
	f.results = make(chan error, 1)
	results, from, ids := f.results, f.from, append([]model.MessageID(nil), f.ids...)
	ctx := context.Background()
	if p.composer != nil {
		ctx = p.composer.ctx
	}
	go func() {
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		results <- source.ForwardMessages(ctx, from, ids, target.ID)
		p.invalidate()
	}()
}

// forwardDialog draws the picker when it is open.
func (p *chatPage) forwardDialog(gtx layout.Context, l localization.Catalog) {
	f := &p.forwarding
	if !f.modal.Shown() {
		return
	}
	targets := p.forwardTargets()
	if !f.busy {
		if f.cancel.Clicked(gtx) {
			f.modal.Close()
		}
		for _, c := range targets {
			if r := f.rows[c.ID]; r != nil && r.Clicked(gtx) {
				p.forward(c)
			}
		}
	}
	sc := scheme(gtx)
	f.modal.Layout(gtx, f.busy, func(gtx layout.Context) layout.Dimensions {
		width := min(gtx.Constraints.Max.X, gtx.Dp(420))
		height := min(gtx.Constraints.Max.Y-gtx.Dp(48), gtx.Dp(560))
		gtx.Constraints = layout.Exact(image.Pt(width, max(height, 0)))
		return card(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, l.T("history.forward_to"), token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
				}),
				vspace(12),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(40))
					fillRounded(gtx, sc.SurfaceContainerHigh, size, gtx.Dp(20))
					return layout.Inset{Top: 10, Bottom: 8, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return flatEditor(gtx, &f.search, l.T("chat.search"))
					})
				}),
				vspace(8),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if f.busy {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return f.loader.sized(gtx, l, 32)
						})
					}
					if len(targets) == 0 {
						return layout.N.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return label(gtx, l.T("chat.not_found"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 1)
							})
						})
					}
					return f.list.Layout(gtx, len(targets), func(gtx layout.Context, i int) layout.Dimensions {
						return p.forwardRow(gtx, targets[i])
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if f.err == nil {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return label(gtx, l.T("history.forward_failed")+": "+mediaErrorText(f.err), token.TypestyleBodySmall, sc.Error.Color, 3)
					})
				}),
				vspace(8),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if f.busy {
							gtx = gtx.Disabled()
						}
						return textButton(gtx, &f.cancel, l.T("history.cancel"))
					})
				}),
			)
		}, defaultCardPadding)
	})
}

// forwardRow draws a chat of the picker: its avatar and title.
func (p *chatPage) forwardRow(gtx layout.Context, c model.Chat) layout.Dimensions {
	f := &p.forwarding
	r := f.rows[c.ID]
	if r == nil {
		r = new(surface)
		f.rows[c.ID] = r
	}
	sc := scheme(gtx)
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(52))
	style := surfaceStyle{radius: gtx.Dp(10), background: sc.Surface.OnColor.SetOpacity(0), content: sc.Surface.OnColor}
	return r.Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
		avatarPx := gtx.Dp(40)
		offset(gtx, image.Pt(gtx.Dp(6), (size.Y-avatarPx)/2), func(gtx layout.Context) layout.Dimensions {
			if p.avatar != nil {
				return p.avatar(gtx, c.ID, c.Kind, c.Title, 40)
			}
			return avatar(gtx, c.ID, c.Kind, c.Title, 40)
		})
		textX := gtx.Dp(6) + avatarPx + gtx.Dp(12)
		text := gtx
		text.Constraints = layout.Constraints{Max: image.Pt(max(size.X-textX-gtx.Dp(6), 0), size.Y)}
		macro := op.Record(gtx.Ops)
		dims := label(text, c.Title, token.TypestyleBodyLarge, sc.Surface.OnColor, 1)
		call := macro.Stop()
		offset(gtx, image.Pt(textX, (size.Y-dims.Size.Y)/2), func(gtx layout.Context) layout.Dimensions {
			call.Add(gtx.Ops)
			return dims
		})
		return layout.Dimensions{Size: size}
	})
}

// snapshotMaxHeight bounds a snapshot to what a GPU texture holds.
const snapshotMaxHeight = 16384

// snapshotResult is a saved snapshot, or why it was not saved.
type snapshotResult struct {
	path string
	err  error
}

// takeSnapshot renders the selected messages as they are shown, on the
// chat's background, to a PNG in the user's pictures. The rows are laid out
// again into operations of their own, without the selection's highlight,
// and with the avatars of private chats, which the history leaves out; a
// headless GPU context renders them off the frame.
func (p *chatPage) takeSnapshot(gtx layout.Context, l localization.Catalog) {
	var msgs []model.Message
	for _, m := range p.messages {
		if p.selection.selected[m.Key.MessageID] {
			msgs = append(msgs, m)
		}
	}
	if len(msgs) == 0 {
		return
	}
	width := gtx.Constraints.Max.X
	ops := new(op.Ops)
	sgtx := gtx.Disabled()
	sgtx.Ops = ops
	sgtx.Constraints = layout.Constraints{Min: image.Pt(width, 0), Max: image.Pt(width, snapshotMaxHeight)}
	pad := gtx.Dp(8)
	type row struct {
		call   op.CallOp
		height int
	}
	var rows []row
	height := 2 * pad
	p.snapshotting = true
	joins := messageJoins(msgs)
	for i, m := range msgs {
		date := i == 0 || !sameDay(m.Date, msgs[i-1].Date)
		macro := op.Record(ops)
		dims := p.row(sgtx, m, date, joins[i], l, false)
		if r := p.rows[m.Key.MessageID]; r != nil && p.avatar != nil && joins[i]&joinBelow == 0 {
			if id := p.senderAvatar(m); id != 0 {
				offset(sgtx, r.avatarPoint, func(gtx layout.Context) layout.Dimensions {
					return p.avatar(gtx, id, model.KindUser, p.avatarName(m), 34)
				})
			}
		}
		rows = append(rows, row{macro.Stop(), dims.Size.Y})
		height += dims.Size.Y
	}
	p.snapshotting = false
	if height > snapshotMaxHeight {
		p.selectionNotice = l.T("history.snapshot_too_tall")
		return
	}
	size := image.Pt(width, height)
	sgtx.Constraints = layout.Exact(size)
	fillRect(sgtx, scheme(gtx).SurfaceContainerLow, size)
	if p.appearance != nil {
		p.appearance.Background(sgtx)
	}
	y := pad
	for _, r := range rows {
		stack := op.Offset(image.Pt(0, y)).Push(ops)
		r.call.Add(ops)
		stack.Pop()
		y += r.height
	}
	results := make(chan snapshotResult, 1)
	p.snapshots = results
	name := fmt.Sprintf("komarugram-go-%s.png", time.Now().Format("2006-01-02-150405"))
	go func() {
		path, err := renderSnapshot(ops, size, name)
		results <- snapshotResult{path, err}
		p.invalidate()
	}()
}

// updateSnapshot tells where the last snapshot was saved.
func (p *chatPage) updateSnapshot(l localization.Catalog) {
	if p.snapshots == nil {
		return
	}
	select {
	case r := <-p.snapshots:
		p.snapshots = nil
		if r.err != nil {
			p.selectionNotice = l.T("history.snapshot_failed") + ": " + mediaErrorText(r.err)
		} else {
			p.selectionNotice = l.Format("history.snapshot_saved", map[string]string{"path": r.path})
		}
	default:
	}
}

// renderSnapshot draws ops of size in a headless GPU context and saves the
// picture as name in the user's pictures.
func renderSnapshot(ops *op.Ops, size image.Point, name string) (string, error) {
	// The context is current on this thread only.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	win, err := headless.NewWindow(size.X, size.Y)
	if err != nil {
		return "", err
	}
	defer win.Release()
	if err := win.Frame(ops); err != nil {
		return "", err
	}
	img := image.NewRGBA(image.Rectangle{Max: size})
	if err := win.Screenshot(img); err != nil {
		return "", err
	}
	dir := picturesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return "", err
	}
	return path, f.Close()
}

// picturesDir is the user's pictures directory: XDG_PICTURES_DIR on Linux,
// Pictures in the home directory elsewhere, or the home directory itself.
func picturesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd" {
		config, err := os.UserConfigDir()
		if err == nil {
			if b, err := os.ReadFile(filepath.Join(config, "user-dirs.dirs")); err == nil {
				for _, line := range strings.Split(string(b), "\n") {
					value, ok := strings.CutPrefix(strings.TrimSpace(line), "XDG_PICTURES_DIR=")
					if !ok {
						continue
					}
					value = strings.Trim(value, `"`)
					value = strings.Replace(value, "$HOME", home, 1)
					if value != "" && value != home {
						return value
					}
				}
			}
		}
	}
	dir := filepath.Join(home, "Pictures")
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	return home
}
