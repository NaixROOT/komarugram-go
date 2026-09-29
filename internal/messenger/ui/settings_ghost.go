// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"slices"

	"gio-mw/token"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/preferences"
)

// ghostOptions are the switches of what the accounts tell others, as
// AyuGram's Ghost Mode has them.
var ghostOptions = []string{"read", "online", "typing", "interact"}

func ghostFromOptions(values []string) preferences.Ghost {
	return preferences.Ghost{
		SendRead:       slices.Contains(values, "read"),
		SendOnline:     slices.Contains(values, "online"),
		SendTyping:     slices.Contains(values, "typing"),
		ReadOnInteract: slices.Contains(values, "interact"),
	}
}

func ghostToOptions(g preferences.Ghost) []string {
	var out []string
	for i, on := range []bool{g.SendRead, g.SendOnline, g.SendTyping, g.ReadOnInteract} {
		if on {
			out = append(out, ghostOptions[i])
		}
	}
	return out
}

// layoutGhost draws the switches of what the accounts tell others. The
// settings may have changed in another window since the last frame; the
// switches follow.
func (p *settingsPage) layoutGhost(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	if want := ghostToOptions(p.ghost()); !slices.Equal(want, p.ghostOpts.GetValues()) {
		p.ghostOpts.SetValues(want)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("ghost.title"), token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("ghost.body"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 0)
		}),
		vspace(8),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.ghostOpts.Layout(gtx, map[string]string{
				"read":     l.T("ghost.read"),
				"online":   l.T("ghost.online"),
				"typing":   l.T("ghost.typing"),
				"interact": l.T("ghost.interact"),
			})
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("ghost.interact_body"), token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 0)
		}),
	)
}

// layoutKeep draws the switches of what the cache keeps that Telegram
// takes back.
func (p *settingsPage) layoutKeep(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	k := p.keep()
	var want []string
	if k.Deleted {
		want = append(want, "deleted")
	}
	if k.Edits {
		want = append(want, "edits")
	}
	if !slices.Equal(want, p.keepOpts.GetValues()) {
		p.keepOpts.SetValues(want)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("keep.title"), token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("keep.body"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 0)
		}),
		vspace(8),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.keepOpts.Layout(gtx, map[string]string{"deleted": l.T("keep.deleted"), "edits": l.T("keep.edits")})
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("keep.deleted_body"), token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 0)
		}),
	)
}
