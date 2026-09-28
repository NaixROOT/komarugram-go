// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"regexp"
	"slices"

	"gio-mw/token"
	"gio-mw/widget/checkbox"
	"gio-mw/widget/toggle"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/preferences"
)

// filterSettings is the part of the privacy settings that edits the
// message filters, as AyuGram's filters settings do.
type filterSettings struct {
	filters    func() preferences.Filters
	setFilters func(preferences.Filters)
	switches   *toggle.Toggle[string]
	flags      *checkbox.Checkboxes[string]
	field      *textField
	add        surface
	remove     []surface
	invalid    bool
}

func newFilterSettings() *filterSettings {
	f := &filterSettings{field: newTextField(0, "")}
	f.switches = toggle.NewToggle([]string{"enable", "chats", "blocked"}, nil, func(values []string) {
		if f.setFilters == nil {
			return
		}
		next := f.filters()
		next.Enabled = slices.Contains(values, "enable")
		next.InChats = slices.Contains(values, "chats")
		next.HideBlocked = slices.Contains(values, "blocked")
		f.setFilters(next)
	})
	f.flags = checkbox.NewCheckboxes([]string{"reversed", "case"}, []string{"case"}, nil)
	return f
}

// Update adds the pattern typed, and deletes the ones asked.
func (f *filterSettings) Update(gtx layout.Context) {
	if f.filters == nil || f.setFilters == nil {
		return
	}
	f.flags.Update(gtx)
	current := f.filters()
	for len(f.remove) < len(current.Patterns) {
		f.remove = append(f.remove, surface{})
	}
	for i := range current.Patterns {
		if f.remove[i].Clicked(gtx) {
			current.Patterns = slices.Delete(slices.Clone(current.Patterns), i, i+1)
			f.setFilters(current)
			return
		}
	}
	if f.field.Submitted(gtx) || f.add.Clicked(gtx) {
		flags := f.flags.GetValues()
		p := preferences.FilterPattern{Text: f.field.Text(), Reversed: slices.Contains(flags, "reversed"), CaseInsensitive: slices.Contains(flags, "case")}
		if _, err := regexp.Compile(p.Text); err != nil || p.Text == "" {
			f.invalid = p.Text != ""
			return
		}
		f.invalid = false
		current.Patterns = append(slices.Clone(current.Patterns), p)
		f.setFilters(current)
		f.field.Clear()
	}
}

// Layout draws the switches, the patterns and the field that adds one.
func (f *filterSettings) Layout(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	current := f.filters()
	var want []string
	for i, on := range []bool{current.Enabled, current.InChats, current.HideBlocked} {
		if on {
			want = append(want, []string{"enable", "chats", "blocked"}[i])
		}
	}
	if !slices.Equal(want, f.switches.GetValues()) {
		f.switches.SetValues(want)
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("filters.title"), token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("filters.body"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 0)
		}),
		vspace(8),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.switches.Layout(gtx, map[string]string{"enable": l.T("filters.enable"), "chats": l.T("filters.in_chats"), "blocked": l.T("filters.blocked")})
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("filters.channels_only"), token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 0)
		}),
		vspace(12),
	}
	if len(current.Patterns) == 0 {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("filters.empty"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 1)
		}))
	}
	for i, p := range current.Patterns {
		if i >= len(f.remove) {
			break
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.patternRow(gtx, i, p, l)
		}))
	}
	children = append(children,
		vspace(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := l.T("filters.expression")
			if f.invalid {
				title = l.T("filters.error")
			}
			return f.field.Layout(gtx, title, f.invalid)
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return f.flags.Layout(gtx, map[string]string{"reversed": l.T("filters.reversed"), "case": l.T("filters.case")})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return textButton(gtx, &f.add, l.T("filters.add"))
				}),
			)
		}),
	)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

// patternRow draws pattern i: its expression, how it matches, and a button
// that deletes it.
func (f *filterSettings) patternRow(gtx layout.Context, i int, p preferences.FilterPattern, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	var notes []string
	if p.Reversed {
		notes = append(notes, l.T("filters.reversed"))
	}
	if p.CaseInsensitive {
		notes = append(notes, l.T("filters.case"))
	}
	if p.Chat != 0 {
		notes = append(notes, l.T("filters.one_chat"))
	}
	button := gtx.Dp(40)
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, p.Text, token.TypestyleBodyLarge, sc.Surface.OnColor, 1)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if len(notes) == 0 {
						return layout.Dimensions{}
					}
					text := notes[0]
					for _, n := range notes[1:] {
						text += " · " + n
					}
					return label(gtx, text, token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 1)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(button, button)
			style := surfaceStyle{radius: button / 2, content: sc.Surface.OnColor, button: l.T("filters.delete")}
			return f.remove[i].Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
				px := gtx.Dp(22)
				return offset(gtx, image.Pt((button-px)/2, (button-px)/2), func(gtx layout.Context) layout.Dimensions {
					return exact(gtx, image.Pt(px, px), func(gtx layout.Context) layout.Dimensions { return iconDelete(gtx, sc.SurfaceVariant.OnColor) })
				})
			})
		}),
	)
}
