// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"testing"
	"time"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// The avatar in a chat's header opens its photo; the rest of the header
// still opens the chat's info.
func TestChatHeaderAvatarOpensPhoto(t *testing.T) {
	store := mockstore.New(time.Now(), 0)
	page := newChatPage(store, func() {})
	defer page.Close()
	var router input.Router
	var avatar, header int
	frame := func() {
		ops := new(op.Ops)
		gtx := sharedContext(ops, image.Pt(900, 800))
		gtx.Source = router.Source()
		if page.headAvatar.Clicked(gtx) {
			avatar++
		}
		if page.header.Clicked(gtx) {
			header++
		}
		layoutChatPage(gtx, model.Chat{ID: 2, Title: "Chat"}, localization.For("en"), func(layout.Context, int64, model.ChatKind, string, unit.Dp) layout.Dimensions {
			return layout.Dimensions{}
		}, nil, func(layout.Context) layout.Dimensions { return layout.Dimensions{} }, page)
		router.Frame(ops)
	}
	click := func(x, y float32) {
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x, y)})
		frame()
		frame()
	}
	frame()
	// The avatar is 40 dp at 16, 8 in the 56 dp header.
	click(36, 28)
	if avatar != 1 || header != 0 {
		t.Fatalf("a click on the avatar: avatar %d, header %d", avatar, header)
	}
	click(200, 25)
	if avatar != 1 || header != 1 {
		t.Fatalf("a click beside the avatar: avatar %d, header %d", avatar, header)
	}
}

// The viewer opens the photo a chat shows at once, and has the older ones
// of its profile after them when the store has told them.
func TestViewerOpensProfilePhotos(t *testing.T) {
	store := mockstore.New(time.Now(), 0)
	var images imageOps
	v := newPhotoViewer(store, &images, func() {})
	defer v.Destroy()
	if !v.OpenProfile(2) || !v.open {
		t.Fatal("the viewer did not open")
	}
	if v.current != model.ProfilePhotoID(0) {
		t.Fatalf("shows %d, not the photo the chat shows", v.current)
	}
	until := time.Now().Add(3 * time.Second)
	for {
		items, _, _, _, _ := v.snapshot()
		if len(items) == 3 {
			for i, m := range items {
				if m.Key.MessageID != model.ProfilePhotoID(i) {
					t.Fatalf("photo %d has id %d", i, m.Key.MessageID)
				}
			}
			break
		}
		if time.Now().After(until) {
			t.Fatalf("%d photos of the profile", len(items))
		}
		time.Sleep(time.Millisecond)
	}
	// A store that has no profile photos opens nothing.
	plain := newPhotoViewer(newGalleryStore(), &images, func() {})
	defer plain.Destroy()
	if plain.OpenProfile(2) || plain.open {
		t.Fatal("a viewer without profile photos opened")
	}
}

// What a profile photo is saved as tells the chat and its place, since it
// is no message.
func TestProfilePhotoFileName(t *testing.T) {
	m := model.Message{Key: model.MessageKey{ChatID: 7, MessageID: model.ProfilePhotoID(2)}}
	if got := photoFileName(m); got != "komarugram-go-avatar-7-3" {
		t.Fatalf("file name %q", got)
	}
	if !model.IsProfilePhoto(m.Key.MessageID) || model.IsProfilePhoto(-5) || model.IsProfilePhoto(42) {
		t.Fatal("profile photos told from messages and stories wrongly")
	}
}

// The avatar at the top of a chat's info opens the photo too.
func TestChatInfoAvatarOpensPhoto(t *testing.T) {
	store := mockstore.New(time.Now(), 0)
	var images imageOps
	p := newChatInfo(store, &images, func() {})
	defer p.Destroy()
	var asked int64
	p.openAvatar = func(chat int64) { asked = chat }
	p.drawAvatar = func(gtx layout.Context, _ int64, _ model.ChatKind, _ string, size unit.Dp) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(gtx.Dp(size), gtx.Dp(size))}
	}
	var router input.Router
	frame := func() {
		ops := new(op.Ops)
		gtx := sharedContext(ops, image.Pt(900, 800))
		gtx.Source = router.Source()
		p.Layout(gtx, localization.For("en"), false)
		router.Frame(ops)
	}
	p.Open(model.Chat{ID: 2, Title: "Chat"})
	awaitShared(t, p)
	for range 30 {
		frame()
		time.Sleep(10 * time.Millisecond)
	}
	// The dialog is centred: the avatar is at its top, under its title.
	for y := float32(60); y < 300 && asked == 0; y += 4 {
		router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(450, y)})
		frame()
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(450, y)}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(450, y)})
		frame()
		frame()
	}
	if asked != 2 {
		t.Fatalf("the avatar asked for the photos of chat %d", asked)
	}
}
