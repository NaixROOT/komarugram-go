// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"fmt"
	"image"
	"strings"
	"time"
	"unicode/utf16"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"

	"gio-mw/token"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

var basicEmoji = strings.Fields("😀 😃 😄 😁 😆 😅 😂 🤣 😊 😇 🙂 🙃 😉 😌 😍 🥰 😘 😗 😙 😚 😋 😛 😝 😜 🤪 🤨 🧐 🤓 😎 🥸 🤩 🥳 😏 😒 😞 😔 😟 😕 🙁 ☹️ 😣 😖 😫 😩 🥺 😢 😭 😤 😠 😡 🤬 🤯 😳 🥵 🥶 😱 😨 😰 😥 😓 🤗 🤔 🫣 🤭 🫢 🫡 🤫 🫠 🤥 😶 😐 😑 😬 🙄 😯 😦 😧 😮 😲 🥱 😴 🤤 😪 😮‍💨 😵 🤐 🥴 🤢 🤮 🤧 😷 🤒 🤕 🤑 🤠 😈 👿 👹 👺 🤡 💩 👻 💀 ☠️ 👽 👾 🤖 🎃 😺 😸 😹 😻 😼 😽 🙀 😿 😾 👋 🤚 🖐️ ✋ 🖖 👌 🤌 🤏 ✌️ 🤞 🫰 🤟 🤘 🤙 👈 👉 👆 🖕 👇 ☝️ 👍 👎 ✊ 👊 🤛 🤜 👏 🙌 🫶 👐 🤲 🤝 🙏 ❤️ 🧡 💛 💚 💙 💜 🖤 🤍 🤎 💔 ❤️‍🔥 💕 💞 💓 💗 💖 💘 💝 💟 🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨 🐯 🦁 🐮 🐷 🐸 🐵 🐔 🐧 🐦 🦆 🦉 🦋 🐝 🐞 🌸 🌼 🌻 🌹 🌺 🌷 🌿 🍀 🌲 🌳 🌴 🌵 🍎 🍊 🍋 🍌 🍉 🍇 🍓 🫐 🍒 🍑 🥭 🍍 🥥 🥝 🍅 🥑 🍆 🥕 🌽 🍕 🍔 🍟 🌭 🍿 🍣 🍜 🍝 🍰 🎂 🍫 ☕ 🍵 🥤 🍺 🍷 🥂 ⚽ 🏀 🏈 ⚾ 🎾 🏐 🎱 🏓 🏸 🥊 🏆 🥇 🎮 🎲 🎯 🎨 🎭 🎤 🎧 🎼 🎹 🎸 🚗 🚕 🚌 🚎 🚓 🚑 🚒 🚲 🛵 🏍️ ✈️ 🚀 🚁 🚂 🚆 🚢 🏠 🏡 🏢 🏰 🗼 🗽 🌍 🌎 🌏 🌞 🌝 🌙 ⭐ 🌟 ✨ ⚡ 🔥 🌈 ☀️ 🌤️ 🌧️ ❄️ ☃️ 💧 🌊 💻 📱 ⌚ 📷 💡 📚 📖 ✏️ 📝 📌 📎 🔑 🔒 🔔 🎁 🎈 🎉 🎊 ✅ ❌ ❗ ❓ 💯 ♻️ 🏳️ 🏴 🏁 🚩 🇺🇦 🇬🇧 🇺🇸 🇫🇷 🇩🇪 🇮🇹 🇪🇸 🇯🇵")

func standardEmoji() []model.PickerItem {
	out := make([]model.PickerItem, 0, len(basicEmoji))
	for _, e := range basicEmoji {
		out = append(out, model.PickerItem{ID: "emoji/" + e, Emoji: e})
	}
	return out
}

type pickerRow struct {
	title    string
	items    []model.PickerItem
	featured *model.PickerPack
}

func (c *messageComposer) pickerRows(width, cell int, l localization.Catalog) []pickerRow {
	var rows []pickerRow
	addItems := func(items []model.PickerItem, featured *model.PickerPack) {
		columns := max(1, width/max(1, cell))
		for len(items) > 0 {
			n := min(columns, len(items))
			if c.tab == model.PickerGIF {
				n = gifRowCount(items, width, cell)
			}
			rows = append(rows, pickerRow{items: items[:n], featured: featured})
			items = items[n:]
		}
	}
	add := func(title string, items []model.PickerItem) {
		if len(items) == 0 {
			return
		}
		if title != "" {
			rows = append(rows, pickerRow{title: title})
		}
		addItems(items, nil)
	}
	if c.query != "" {
		add(l.T("composer.results"), c.page.Items)
		return rows
	}
	if c.tab == model.PickerGIF {
		add("", c.page.Items)
		return rows
	}
	if c.selectedPack == 0 {
		add(l.T("composer.recent"), model.PreferSaved(c.recent[c.tab], c.page.Recent))
		if c.tab == model.PickerEmoji {
			items := c.page.Items
			if len(items) == 0 {
				items = standardEmoji()
			}
			add(l.T("composer.all"), items)
		}
	}
	for _, pack := range c.page.Packs {
		if c.selectedPack == 0 || c.selectedPack == pack.ID {
			add(pack.Title, pack.Items)
		}
	}
	if len(c.page.Featured) > 0 && (c.selectedPack == 0 || c.selectedFeatured() != nil) {
		if c.selectedPack == 0 {
			rows = append(rows, pickerRow{title: l.T("composer.featured")})
		}
		for i := range c.page.Featured {
			pack := &c.page.Featured[i]
			if c.selectedPack != 0 && c.selectedPack != pack.ID {
				continue
			}
			rows = append(rows, pickerRow{title: pack.Title, featured: pack})
			addItems(pack.Items, pack)
		}
	}
	return rows
}
func gifAspect(i model.PickerItem) float32 {
	if m := i.Media.Media; m != nil && m.Width > 0 && m.Height > 0 {
		return max(.5, min(2.5, float32(m.Width)/float32(m.Height)))
	}
	return 1
}
func gifRowCount(items []model.PickerItem, width, target int) int {
	sum := float32(0)
	for n, i := range items {
		sum += gifAspect(i)
		if sum*float32(target) >= float32(width) || n == 3 {
			return n + 1
		}
	}
	return len(items)
}
func gifWidths(items []model.PickerItem, width, gap int) ([]int, int) {
	widths := make([]int, len(items))
	sum := float32(0)
	for _, i := range items {
		sum += gifAspect(i)
	}
	if sum == 0 {
		return widths, 0
	}
	available := max(0, width-gap*(len(items)-1))
	height := max(1, int(float32(available)/sum))
	left := available
	for n, i := range items {
		w := min(left, int(float32(height)*gifAspect(i)))
		if n == len(items)-1 {
			w = left
		}
		widths[n] = w
		left -= w
	}
	return widths, height
}
func (c *messageComposer) choose(gtx layout.Context, item model.PickerItem) {
	c.chooseIn(gtx, c.tab, item)
}

// chooseIn sends a sticker or GIF, or inserts an emoji, chosen from tab's
// items, which a sticker set's dialog shows too. It reports whether it did.
func (c *messageComposer) chooseIn(gtx layout.Context, tab model.PickerTab, item model.PickerItem) bool {
	d := c.draft(c.chat)
	if d.sending {
		return false
	}
	if tab == model.PickerEmoji {
		text := item.Emoji
		if text == "" {
			return false
		}
		old := d.editor.Text()
		a, b := d.editor.Selection()
		pos := min(a, b)
		d.editor.Insert(text)
		next := d.editor.Text()
		d.entities = shiftEmojiEntities(old, next, d.entities)
		if item.Custom {
			d.entities = append(d.entities, model.Entity{Kind: "emoji", Offset: len(utf16.Encode([]rune(old)[:pos])), Length: len(utf16.Encode([]rune(text))), DocumentID: item.DocumentID})
		}
		d.text = next
		d.pending = nil
		d.err = nil
		gtx.Execute(key.FocusCmd{Tag: &d.editor})
	} else if c.asksConfirmation(tab) {
		c.sendConfirm.ask(tab, item)
		c.pickerOpen = false
	} else {
		c.submit(c.chat, model.OutgoingMessage{Item: &item})
		c.pickerOpen = false
	}
	if source, ok := c.source.(model.PickerRecentStore); ok && tab != model.PickerGIF {
		go func() {
			ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
			defer cancel()
			_ = source.RememberPicker(ctx, tab, item)
		}()
	}
	c.recent[tab] = model.PreferSaved([]model.PickerItem{item}, c.recent[tab])
	if len(c.recent[tab]) > 40 {
		c.recent[tab] = c.recent[tab][:40]
	}
	gtx.Execute(op.InvalidateCmd{})
	return true
}

func (c *messageComposer) pickerLayout(gtx layout.Context, l localization.Catalog, p *chatPage, animate bool, origin image.Point) layout.Dimensions {
	size := gtx.Constraints.Max
	sc := scheme(gtx)
	defer clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(16)).Push(gtx.Ops).Pop()
	overlayFill(gtx, p.menuBackdrop(), size, origin, sc.SurfaceContainerHigh, gtx.Dp(16))
	event.Op(gtx.Ops, &c.list)
	c.hover.Update(gtx, &c.list.List, &c.strip)
	c.hover.Op(gtx)
	tabHeight := min(gtx.Dp(44), size.Y)
	searchHeight := min(gtx.Dp(48), max(0, size.Y-tabHeight))
	footer := min(gtx.Dp(48), max(0, size.Y-tabHeight-searchHeight))
	pad := gtx.Dp(10)
	inRect(gtx, image.Rect(0, 0, size.X, tabHeight), func(gtx layout.Context) layout.Dimensions {
		return c.tabs.Layout(gtx, []string{l.T("composer.emoji"), l.T("composer.stickers"), l.T("composer.gif")}, int(c.tab))
	})
	inRect(gtx, image.Rect(pad, tabHeight+gtx.Dp(4), max(pad, size.X-pad), tabHeight+searchHeight-gtx.Dp(4)), func(gtx layout.Context) layout.Dimensions {
		fillRounded(gtx, sc.SurfaceContainerLow, gtx.Constraints.Max, gtx.Dp(16))
		return layout.Inset{Top: 7, Bottom: 4, Left: 12, Right: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return flatEditor(gtx, &c.search, l.T("composer.search")) })
	})
	body := image.Rect(pad, tabHeight+searchHeight, size.X-pad, max(tabHeight+searchHeight, size.Y-footer))
	// The content of a tab switched to slides in from its side and fades in.
	sliding := func(area image.Rectangle, w layout.Widget) { c.tabs.Slide(gtx, area, w) }
	cell := gtx.Dp(68)
	if c.tab == model.PickerEmoji {
		cell = gtx.Dp(40)
	}
	if c.tab == model.PickerGIF {
		cell = gtx.Dp(100)
	}
	rows := c.pickerRows(max(0, body.Dx()), cell, l)
	sliding(body, func(gtx layout.Context) layout.Dimensions {
		if c.loading && len(rows) == 0 {
			// The first page loads in the middle of the empty picker.
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				n := min(gtx.Dp(48), gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
				gtx.Constraints = layout.Exact(image.Pt(n, n))
				return c.loader.Layout(gtx, l)
			})
		}
		return c.list.Layout(gtx, len(rows)+1, func(gtx layout.Context, n int) layout.Dimensions {
			if n == len(rows) {
				if c.loading || c.featuredLoading {
					// More loads under what is shown.
					gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, gtx.Dp(56)))
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						n := gtx.Dp(32)
						gtx.Constraints = layout.Exact(image.Pt(n, n))
						return c.loader.Layout(gtx, l)
					})
				}
				if c.pickerErr != nil {
					if c.retry.Clicked(gtx) && c.source != nil {
						if pack := c.selectedFeatured(); pack != nil {
							c.loadFeatured(*pack)
						} else {
							c.request(l, false)
						}
					}
					return textButton(gtx, &c.retry, l.T("composer.retry"))
				}
				if c.page.Next != "" {
					if c.more.Clicked(gtx) && c.source != nil {
						c.request(l, true)
					}
					return textButton(gtx, &c.more, l.T("composer.more"))
				}
				if len(rows) == 0 {
					return label(gtx, l.T("composer.empty"), token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 2)
				}
				return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, gtx.Dp(8))}
			}
			row := rows[n]
			if len(row.items) == 0 {
				if row.featured != nil {
					pack := row.featured
					click := c.packClicks[pack.ID]
					if click == nil {
						click = new(surface)
						c.packClicks[pack.ID] = click
					}
					if click.Clicked(gtx) {
						p.stickers.open(p, pack.Ref)
						c.pickerOpen = false
					}
					size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(40))
					style := surfaceStyle{radius: gtx.Dp(8), background: sc.Surface.OnColor.SetOpacity(0), content: sc.Surface.OnColor,
						button: l.T("composer.view_pack") + " " + pack.Title}
					return click.Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: 8, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return label(gtx, pack.Title, token.TypestyleTitleSmall, sc.Surface.OnColor, 1)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Inset{Left: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return label(gtx, l.T("composer.view_pack"), token.TypestyleLabelMedium, sc.Primary.Color, 1)
									})
								}),
							)
						})
					})
				}
				if row.title == "" {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: 10, Bottom: 6, Left: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return label(gtx, row.title, token.TypestyleLabelLarge, sc.SurfaceVariant.OnColor, 1)
				})
			}
			width := gtx.Constraints.Max.X
			gap := gtx.Dp(4)
			height := cell
			widths := make([]int, len(row.items))
			cols := max(1, width/max(1, cell))
			for i := range widths {
				widths[i] = width / cols
			}
			if c.tab == model.PickerGIF {
				widths, height = gifWidths(row.items, width, gap)
			}
			x := 0
			for i, item := range row.items {
				key := fmt.Sprintf("%d/%d/%s", c.tab, n, item.ID)
				click := c.itemClicks[key]
				if click == nil {
					click = new(surface)
					c.itemClicks[key] = click
				}
				// A set not added is still sent from, as in Telegram
				// Desktop; its title row opens it.
				if click.Clicked(gtx) {
					c.choose(gtx, item)
				}
				tileWidth := widths[i] - gap
				if c.tab == model.PickerGIF {
					tileWidth = widths[i]
				}
				inRect(gtx, image.Rect(x, 0, x+max(0, tileWidth), height), func(gtx layout.Context) layout.Dimensions {
					style := surfaceStyle{radius: gtx.Dp(8), background: sc.Surface.OnColor.SetOpacity(0), content: sc.Surface.OnColor}
					return click.Layout(gtx, gtx.Constraints.Max, style, func(gtx layout.Context) layout.Dimensions {
						hovered := click.click.Hovered()
						cheap := hovered && p.media.CheapToPlay(item.Media, gtx.Constraints.Max, c.tab == model.PickerGIF)
						return c.drawItem(gtx, item, p, animate || c.hover.Play(gtx, key, hovered, cheap))
					})
				})
				x += widths[i]
				if c.tab == model.PickerGIF {
					x += gap
				}
			}
			return layout.Dimensions{Size: image.Pt(width, height+gap)}
		})
	})
	footerArea := image.Rect(0, size.Y-footer, size.X, size.Y)
	inRect(gtx, footerArea, func(gtx layout.Context) layout.Dimensions {
		fillRect(gtx, sc.SurfaceContainerLow, gtx.Constraints.Max)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	sliding(footerArea, func(gtx layout.Context) layout.Dimensions {
		if c.tab == model.PickerGIF {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(gtx, l.T("composer.gif"), token.TypestyleLabelLarge, sc.SurfaceVariant.OnColor, 1)
			})
		}
		dims := c.strip.Layout(gtx, len(c.page.Packs)+len(c.page.Featured)+1, func(gtx layout.Context, n int) layout.Dimensions {
			id := int64(0)
			var item model.PickerItem
			var featured *model.PickerPack
			if n > 0 {
				var pack *model.PickerPack
				if n <= len(c.page.Packs) {
					pack = &c.page.Packs[n-1]
				} else {
					featured = &c.page.Featured[n-1-len(c.page.Packs)]
					pack = featured
				}
				id = pack.ID
				if len(pack.Items) > 0 {
					item = pack.Items[0]
				}
			}
			click := c.packClicks[id]
			if click == nil {
				click = new(surface)
				c.packClicks[id] = click
			}
			if click.Clicked(gtx) {
				c.cancelFeatured()
				c.selectedPack = id
				if c.query != "" {
					c.due = gtx.Now
				}
				c.query = ""
				c.search.SetText("")
				c.list.Position = layout.Position{}
				if !c.due.IsZero() || c.page.Next != "" {
					c.due = gtx.Now
				}
				if featured != nil && !featured.Complete {
					c.loadFeatured(*featured)
				}
				gtx.Execute(op.InvalidateCmd{})
			}
			gtx.Constraints = layout.Exact(image.Pt(gtx.Dp(44), footer))
			style := surfaceStyle{radius: gtx.Dp(10), background: sc.SecondaryContainer.Color.SetOpacity(0), content: sc.Surface.OnColor}
			if c.selectedPack == id {
				style.background, style.content = sc.SecondaryContainer.Color, sc.SecondaryContainer.OnColor
			}
			return click.Layout(gtx, gtx.Constraints.Max, style, func(gtx layout.Context) layout.Dimensions {
				if n == 0 {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return label(gtx, "◷", token.TypestyleHeadlineSmall, sc.Surface.OnColor, 1)
					})
				}
				return layout.UniformInset(6).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					hovered := click.click.Hovered()
					cheap := hovered && p.media.CheapToPlay(item.Media, gtx.Constraints.Max, c.tab == model.PickerGIF)
					return c.drawItem(gtx, item, p, animate || c.hover.Play(gtx, pickerStripKey(id), hovered, cheap))
				})
			})
		})
		dragHorizontalStrip(gtx, size.X, &c.stripDrag, &c.strip, c.invalidate)
		return dims
	})
	if c.pickerErr != nil && c.pickerErr != c.pickerTold {
		c.pickerToast.Show(mediaErrorText(c.pickerErr))
	}
	c.pickerTold = c.pickerErr
	// Over the end of the list, above the footer of packs.
	c.pickerToast.Layout(gtx, image.Rect(0, 0, size.X, size.Y-footer))
	return layout.Dimensions{Size: size}
}

// pickerShows reports whether the picker's page shows the media of id.
func (c *messageComposer) pickerShows(id string) bool {
	shows := func(items []model.PickerItem) bool {
		for _, item := range items {
			if item.Media.Media != nil && item.Media.Media.ID == id {
				return true
			}
		}
		return false
	}
	for _, pack := range c.page.Packs {
		if shows(pack.Items) {
			return true
		}
	}
	for _, pack := range c.page.Featured {
		if shows(pack.Items) {
			return true
		}
	}
	for _, items := range c.recent {
		if shows(items) {
			return true
		}
	}
	return shows(c.page.Items) || shows(c.page.Recent)
}

// pickerStripKey is the hover key of a pack in the picker's strip.
type pickerStripKey int64

func (c *messageComposer) drawItem(gtx layout.Context, item model.PickerItem, p *chatPage, animate bool) layout.Dimensions {
	size := gtx.Constraints.Max
	if item.Media.Media != nil && p.media != nil {
		status := p.media.StatusFit(item.Media, animate, size, c.tab == model.PickerGIF)
		im := status.Frame
		if im == nil {
			im = status.Preview
		}
		if im != nil && p.images != nil {
			defer clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(6)).Push(gtx.Ops).Pop()
			fit := widget.Contain
			if c.tab == model.PickerGIF {
				fit = widget.Cover
			}
			return widget.Image{Src: p.images.Op(im), Fit: fit}.Layout(gtx)
		}
	}
	if c.tab == model.PickerGIF {
		// A large GIF has no thumbnail and shows nothing until played.
		fillRounded(gtx, scheme(gtx).SurfaceContainerHighest, size, gtx.Dp(6))
		return layout.Dimensions{Size: size}
	}
	txt := item.Emoji
	if txt == "" {
		txt = "·"
	}
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// An emoji the fonts cannot draw as one glyph is wider than its
		// cell; it shrinks to fit rather than being cut to an ellipsis.
		cell := gtx.Constraints.Max
		gtx.Constraints = layout.Constraints{Max: image.Pt(1<<20, cell.Y)}
		macro := op.Record(gtx.Ops)
		dims := label(gtx, txt, token.TypestyleHeadlineSmall, scheme(gtx).Surface.OnColor, 1)
		call := macro.Stop()
		if dims.Size.X <= cell.X {
			call.Add(gtx.Ops)
			return dims
		}
		scale := float32(cell.X) / float32(dims.Size.X)
		defer op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale))).Push(gtx.Ops).Pop()
		call.Add(gtx.Ops)
		return layout.Dimensions{Size: image.Pt(cell.X, int(float32(dims.Size.Y)*scale+.5))}
	})
}
