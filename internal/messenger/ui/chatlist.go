// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"strings"
	"time"

	"komarugram/internal/diagnostics"

	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/widget/scroll"
	"gio-mw/widget/search"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

const (
	chatRowHeight  = unit.Dp(64)
	chatAvatarSize = unit.Dp(48)
	listTopGap     = unit.Dp(6)
	searchHeader   = unit.Dp(72)
	chatRowPadding = unit.Dp(10)
	chatRowMargin  = unit.Dp(6)
	// chatAvatarInset is where avatars start in wide and narrow lists alike,
	// so that they stay in place when the list collapses.
	chatAvatarInset  = chatRowMargin + chatRowPadding
	chatSelectRadius = unit.Dp(10)
)

// chatList is the list of chats of a section.
type chatList struct {
	// toast tells, at the bottom of the list, what happened in it.
	toast  toast
	avatar avatarLayout
	// badges draw the marks around a chat's name: see App.badges.
	badges badgesLayout
	list   scroll.List
	rows   map[int64]*chatRow
	// shown are the chats laid out in the last frame; only they can have
	// been clicked.
	shown  []int64
	search *search.Search
	// panel searches messages too, when the store can; nil filters the
	// chats by title only.
	panel *searchPanel
	// items are the search results laid out in the last frame.
	items []searchItem
}

// chatPick is a chat picked in the list: a chat of the list or found by a
// search, or a message found in one.
type chatPick struct {
	ID int64
	// Chat is the chat picked from search results, which may not be in the
	// account's chat list.
	Chat *model.Chat
	// Message is the message to open the chat at, 0 for none.
	Message model.MessageID
}

type chatRow struct {
	surface
}

func newChatList() *chatList {
	bar := search.Bar()
	bar.SupportingText = "Поиск чатов"
	bar.LeadingIcon.Icon = iconSearch
	return &chatList{
		list:   scroll.List{List: layout.List{Axis: layout.Vertical}},
		rows:   make(map[int64]*chatRow),
		search: bar,
	}
}

func (l *chatList) row(id int64) *chatRow {
	r, ok := l.rows[id]
	if !ok {
		r = new(chatRow)
		l.rows[id] = r
	}
	return r
}

// visible returns the chats the section lists.
func (l *chatList) visible(sec section, folders []model.Folder, chats []model.Chat) []model.Chat {
	var match func(model.Chat) bool
	switch sec.kind {
	case sectionFolder:
		for _, f := range folders {
			if f.ID == sec.folder {
				match = f.Contains
			}
		}
	case sectionSearch:
		query := strings.ToLower(strings.TrimSpace(l.search.GetText()))
		if query != "" {
			match = func(c model.Chat) bool { return strings.Contains(strings.ToLower(c.Title), query) }
		}
	}
	if match == nil {
		return chats
	}
	var out []model.Chat
	for _, c := range chats {
		if match(c) {
			out = append(out, c)
		}
	}
	return out
}

// searching reports whether the search panel shows its results.
func (l *chatList) searching(sec section, narrow bool) bool {
	return l.panel != nil && sec.kind == sectionSearch && !narrow
}

// Update returns what the user picked, if anything.
func (l *chatList) Update(gtx layout.Context, sec section, catalog localization.Catalog) (chatPick, bool) {
	l.search.SupportingText = catalog.T("chat.search")
	if l.search.TrailingIcon.Clickable.Clicked(gtx) {
		l.search.ClearText()
		l.search.Focus(gtx)
	}
	// The clear button shows only while there is something to clear.
	if l.search.GetText() != "" {
		l.search.TrailingIcon.Icon, l.search.TrailingIcon.Label = iconClear, catalog.T("chat.clear")
	} else {
		l.search.TrailingIcon.Icon, l.search.TrailingIcon.Label = nil, ""
	}
	if l.panel != nil && sec.kind == sectionSearch {
		l.panel.Update(gtx, l.search.GetText())
		// A chat picked from what was found goes first in the search
		// history, as does one picked from the history itself.
		typed := strings.TrimSpace(l.search.GetText()) != ""
		remember := func(c model.Chat) {
			if store := l.panel.recent.store; store != nil {
				store.BumpRecentChat(c)
			}
		}
		for _, it := range l.items {
			switch {
			case it.chat != nil && l.row(it.chat.ID).click.Clicked(gtx):
				c := *it.chat
				if typed || it.recent {
					remember(c)
				}
				return chatPick{ID: c.ID, Chat: &c}, true
			case it.found != nil && l.panel.foundRow(it.found.Message.Key).click.Clicked(gtx):
				c := it.found.Chat
				if typed {
					remember(c)
				}
				return chatPick{ID: c.ID, Chat: &c, Message: it.found.Message.Key.MessageID}, true
			}
		}
	}
	for _, id := range l.shown {
		if l.row(id).click.Clicked(gtx) {
			return chatPick{ID: id}, true
		}
	}
	return chatPick{}, false
}

// Layout draws the list in the exact size of the constraints.
func (l *chatList) Layout(gtx layout.Context, sec section, folders []model.Folder, chats []model.Chat, selected int64, narrow bool, catalog localization.Catalog) layout.Dimensions {
	end := diagnostics.From(gtx.Values).Begin("chat-list")
	defer end()
	sc := scheme(gtx)
	size := gtx.Constraints.Max
	fillRect(gtx, sc.Surface.Color, size)

	// Only search has a header; otherwise the chats take the full height
	// under a small gap.
	headerHeight := gtx.Dp(listTopGap)
	if sec.kind == sectionSearch && !narrow {
		headerHeight = gtx.Dp(searchHeader)
		offset(gtx, image.Pt(gtx.Dp(12), gtx.Dp(8)), func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(size.X-gtx.Dp(24), headerHeight-gtx.Dp(8)))
			gtx.Constraints.Min.Y = 0
			return l.search.Layout(gtx)
		})
	}

	if l.searching(sec, narrow) {
		panelTop := headerHeight
		headerHeight += l.panel.headerHeight(gtx)
		offset(gtx, image.Pt(0, panelTop), func(gtx layout.Context) layout.Dimensions {
			l.panel.layoutHeader(gtx, catalog)
			return layout.Dimensions{}
		})
		// The results of the other tab slide in, as a tab's content does.
		area := image.Rect(0, headerHeight, size.X, max(size.Y, headerHeight))
		l.panel.tabs.Slide(gtx, area, func(gtx layout.Context) layout.Dimensions {
			return l.layoutResults(gtx, selected, catalog)
		})
		if text := l.panel.failure(catalog); text != "" {
			l.toast.Show(text)
		}
		l.toast.Layout(gtx, area)
		return layout.Dimensions{Size: size}
	}
	l.items = l.items[:0]

	visible := l.visible(sec, folders, chats)
	listGtx := gtx
	listGtx.Constraints = layout.Exact(image.Pt(size.X, max(size.Y-headerHeight, 0)))
	offset(listGtx, image.Pt(0, headerHeight), func(gtx layout.Context) layout.Dimensions {
		if len(visible) == 0 {
			return layout.N.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return centeredLabel(gtx, catalog.T("chat.not_found"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 2)
				})
			})
		}
		now := time.Now()
		l.shown = l.shown[:0]
		return l.list.Layout(gtx, len(visible), func(gtx layout.Context, i int) layout.Dimensions {
			c := visible[i]
			l.shown = append(l.shown, c.ID)
			return l.layoutRow(gtx, c, c.ID == selected, narrow, now, catalog)
		})
	})
	l.toast.Layout(gtx, image.Rect(0, headerHeight, size.X, size.Y))
	return layout.Dimensions{Size: size}
}

// withMembersLine gives c, a chat found by a search or remembered from one,
// how big it is as its line, when it has no message to show.
func withMembersLine(c model.Chat, catalog localization.Catalog) model.Chat {
	if c.LastMessage == "" && c.Members > 0 {
		key := "status.members"
		if c.Kind == model.KindChannel {
			key = "status.subscribers"
		}
		c.LastMessage = catalog.Count(key, c.Members, nil)
	}
	return c
}

// layoutResults draws the results of the search panel.
func (l *chatList) layoutResults(gtx layout.Context, selected int64, catalog localization.Catalog) layout.Dimensions {
	l.items = l.panel.items(catalog)
	l.shown = l.shown[:0]
	now := time.Now()
	recent := &l.panel.recent
	dims := l.list.Layout(gtx, len(l.items), func(gtx layout.Context, i int) layout.Dimensions {
		it := l.items[i]
		switch {
		case it.recentHeading:
			return recent.layoutHeading(gtx, catalog)
		case it.recent:
			dims := l.layoutRow(gtx, withMembersLine(*it.chat, catalog), it.chat.ID == selected, false, now, catalog)
			recent.rowOp(gtx, it.chat.ID, dims.Size)
			return dims
		case it.chat != nil:
			return l.layoutRow(gtx, withMembersLine(*it.chat, catalog), it.chat.ID == selected, false, now, catalog)
		case it.found != nil:
			return l.layoutFound(gtx, l.panel, *it.found, false, now, catalog)
		case it.heading != "":
			return layoutHeading(gtx, it.heading)
		}
		return l.panel.layoutStatus(gtx, catalog)
	})
	l.panel.moreNear(l.list.Position, len(l.items))
	// Over the rows, which would keep the press from anything under them.
	recent.areaOp(gtx)
	recent.layoutMenu(gtx, catalog)
	return dims
}

func (l *chatList) layoutRow(gtx layout.Context, c model.Chat, selected, narrow bool, now time.Time, catalog localization.Catalog) layout.Dimensions {
	return l.layoutRowWith(gtx, &l.row(c.ID).surface, c, selected, narrow, now, catalog)
}

// layoutRowWith draws the row of c, which takes its clicks with r.
func (l *chatList) layoutRowWith(gtx layout.Context, r *surface, c model.Chat, selected, narrow bool, now time.Time, catalog localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	margin := gtx.Dp(chatRowMargin)
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(chatRowHeight))

	titleColor := sc.Surface.OnColor
	textColor := sc.SurfaceVariant.OnColor
	accent := sc.Primary.Color
	background := sc.Primary.Color.SetOpacity(0)
	if selected {
		background = sc.Primary.Color
		titleColor, textColor, accent = sc.Primary.OnColor, sc.Primary.OnColor, sc.Primary.OnColor
	}
	style := surfaceStyle{
		area:       image.Rectangle{Min: image.Pt(margin, 0), Max: image.Pt(size.X-margin, size.Y)},
		radius:     gtx.Dp(chatSelectRadius),
		background: background,
		content:    titleColor,
	}
	return r.Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
		pad := gtx.Dp(chatRowPadding)
		avatarPx := gtx.Dp(chatAvatarSize)
		avatarY := (size.Y - avatarPx) / 2
		avatarX := gtx.Dp(chatAvatarInset)
		offset(gtx, image.Pt(avatarX, avatarY), func(gtx layout.Context) layout.Dimensions {
			if l.avatar != nil {
				return l.avatar(gtx, c.ID, c.Kind, c.Title, chatAvatarSize)
			}
			return avatar(gtx, c.ID, c.Kind, c.Title, chatAvatarSize)
		})

		badgeBackground, badgeText := sc.Primary.Color, sc.Primary.OnColor
		if c.Muted {
			badgeBackground, badgeText = sc.Outline, sc.Surface.Color
		}
		if selected {
			badgeBackground, badgeText = sc.Primary.OnColor, sc.Primary.Color
		}
		if narrow {
			if c.Unread > 0 {
				drawBadgeRight(gtx, image.Pt(avatarX+avatarPx+gtx.Dp(4), avatarY+avatarPx-gtx.Dp(16)), c.Unread, badgeBackground, badgeText)
			}
		} else {
			textX := avatarX + avatarPx + pad
			textWidth := size.X - margin - pad - textX
			l.layoutRowText(gtx, c, image.Rect(textX, gtx.Dp(11), textX+textWidth, size.Y-gtx.Dp(11)),
				titleColor, textColor, accent, badgeBackground, badgeText, now, catalog)
		}
		return layout.Dimensions{Size: size}
	})
}

// layoutRowText draws the title, time, last message and unread counter in
// area.
func (l *chatList) layoutRowText(gtx layout.Context, c model.Chat, area image.Rectangle, titleColor, textColor, accent, badgeBackground, badgeText token.MatColor, now time.Time, catalog localization.Catalog) {
	measure := func(w layout.Widget) (layout.Dimensions, op.CallOp) {
		macro := op.Record(gtx.Ops)
		dims := w(gtx)
		return dims, macro.Stop()
	}
	width := area.Dx()
	gtx.Constraints = layout.Constraints{Max: image.Pt(width, area.Dy())}

	// Top line: title and time.
	timeDims, timeCall := measure(func(gtx layout.Context) layout.Dimensions {
		if c.LastTime.IsZero() {
			return layout.Dimensions{} // A chat found by a search.
		}
		return label(gtx, chatTime(c.LastTime, now, catalog), token.TypestyleLabelSmall, textColor, 1)
	})
	titleGtx := gtx
	titleGtx.Constraints.Max.X = max(width-timeDims.Size.X-gtx.Dp(8), 0)
	offset(titleGtx, area.Min, func(gtx layout.Context) layout.Dimensions {
		var before, after layout.Widget
		if l.badges != nil {
			before, after = l.badges(c.Badges, 16, false)
		}
		return withBadges(gtx, before, func(gtx layout.Context) layout.Dimensions {
			kind := chatKindIcon(c.Kind)
			if kind == nil {
				return label(gtx, c.Title, token.TypestyleTitleSmallEmphasized, titleColor, 1)
			}
			// As materialgram marks them: groups, channels and bots have
			// an icon of their kind before the title.
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return exact(gtx, image.Pt(gtx.Dp(16), gtx.Dp(16)), func(gtx layout.Context) layout.Dimensions {
						return kind(gtx, titleColor)
					})
				}),
				layout.Rigid(layout.Spacer{Width: 4}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, c.Title, token.TypestyleTitleSmallEmphasized, titleColor, 1)
				}),
			)
		}, after)
	})
	offset(gtx, image.Pt(area.Max.X-timeDims.Size.X, area.Min.Y+gtx.Dp(2)), func(gtx layout.Context) layout.Dimensions {
		timeCall.Add(gtx.Ops)
		return timeDims
	})

	// Bottom line: sender, message and the unread counter.
	lineY := area.Min.Y + gtx.Dp(22)
	messageWidth := width
	if c.Unread > 0 {
		badgeWidth := drawBadgeRight(gtx, image.Pt(area.Max.X, lineY+gtx.Dp(1)), c.Unread, badgeBackground, badgeText)
		messageWidth -= badgeWidth + gtx.Dp(8)
	}
	x := area.Min.X
	if c.LastSender != "" {
		senderGtx := gtx
		senderGtx.Constraints.Max.X = max(messageWidth/2, 0)
		dims := offset(senderGtx, image.Pt(x, lineY), func(gtx layout.Context) layout.Dimensions {
			return label(gtx, c.LastSender+": ", token.TypestyleBodyMedium, accent, 1)
		})
		x += dims.Size.X
		messageWidth -= dims.Size.X
	}
	messageGtx := gtx
	messageGtx.Constraints.Max.X = max(messageWidth, 0)
	offset(messageGtx, image.Pt(x, lineY), func(gtx layout.Context) layout.Dimensions {
		return label(gtx, c.LastMessage, token.TypestyleBodyMedium, textColor, 1)
	})
}

// chatKindIcon is the mark of a chat's kind, nil for private chats.
func chatKindIcon(kind model.ChatKind) wdk.IconWidget {
	switch kind {
	case model.KindGroup:
		return iconKindGroup
	case model.KindChannel:
		return iconKindChannel
	case model.KindBot:
		return iconKindBot
	}
	return nil
}
