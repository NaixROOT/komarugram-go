// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"errors"
	"slices"

	"gio-mw/token"
	"gio-mw/widget/radio"

	"gioui.org/layout"
	"gioui.org/op"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
	"komarugram/pkg/miniapp"
	"komarugram/pkg/player"
)

// playerFor picks the external player for media: the chosen one while it is
// installed, else the only one installed. ask is set when several are
// installed and none of them is chosen; kind is "" when none is installed.
func playerFor(chosen player.Kind, installed []player.Kind) (kind player.Kind, ask bool) {
	for _, k := range installed {
		if k == chosen {
			return k, false
		}
	}
	switch len(installed) {
	case 0:
		return "", false
	case 1:
		return installed[0], false
	}
	return "", true
}

// playerChoice asks which of the installed players to open media in, the
// first time there is more than one to choose from.
type playerChoice struct {
	modal  modal
	kinds  []player.Kind
	msg    model.Message
	report func(error)
	cancel surface
	pick   map[player.Kind]*surface
}

// play opens m in the external player, asking which one first when the user
// has not chosen and more than one is installed.
func (p *chatPage) play(gtx layout.Context, m model.Message, report func(error), l localization.Catalog) {
	report(nil)
	var chosen player.Kind
	if p.player != nil {
		chosen = p.player()
	}
	installed := player.Installed(p.customPlayers())
	kind, ask := playerFor(chosen, installed)
	switch {
	case ask:
		c := &p.playerChoice
		c.kinds, c.msg, c.report = installed, m, report
		c.modal.Open()
		gtx.Execute(op.InvalidateCmd{})
	case kind == "":
		report(errors.New(l.T("player.none")))
	default:
		p.media.Play(m, kind, kind.Resolve(p.customPlayers()[kind]), report)
	}
}

// customPlayers are the players the user pointed at, by kind.
func (p *chatPage) customPlayers() map[player.Kind]string {
	if p.playerPaths == nil {
		return nil
	}
	return p.playerPaths()
}

func (p *chatPage) playerDialog(gtx layout.Context, l localization.Catalog) {
	c := &p.playerChoice
	if !c.modal.Shown() {
		return
	}
	if c.pick == nil {
		c.pick = make(map[player.Kind]*surface)
	}
	for _, kind := range c.kinds {
		if c.pick[kind] == nil {
			c.pick[kind] = new(surface)
		}
	}
	if c.cancel.Clicked(gtx) {
		c.modal.Close()
	}
	for _, kind := range c.kinds {
		if c.pick[kind].Clicked(gtx) && !c.modal.closing {
			if p.setPlayer != nil {
				p.setPlayer(kind)
			}
			p.media.Play(c.msg, kind, kind.Resolve(p.customPlayers()[kind]), c.report)
			c.modal.Close()
		}
	}
	sc := scheme(gtx)
	c.modal.Layout(gtx, false, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(440))
		return card(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, l.T("player.ask"), token.TypestyleTitleMedium, sc.Surface.OnColor, 2)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if len(c.kinds) < 2 {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						body := l.Format("player.ask_body", map[string]string{"first": c.kinds[0].Title(), "second": c.kinds[1].Title()})
						return label(gtx, body, token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 4)
					})
				}),
				vspace(16),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					buttons := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return textButton(gtx, &c.cancel, l.T("history.cancel"))
					})}
					for _, kind := range c.kinds {
						buttons = append(buttons, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return textButton(gtx, c.pick[kind], kind.Title())
						}))
					}
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.Flex{Spacing: layout.SpaceStart, Alignment: layout.Middle}.Layout(gtx, buttons...)
				}),
			)
		}, defaultCardPadding)
	})
}

// playerSettings is the choice of the external player in the settings. Its
// last option, "", is to ask when a video opens, which is what the app does
// while the user has not chosen.
type playerSettings struct {
	radios *radio.Radios[player.Kind]
	// installed is looked up when the settings open, not on every frame.
	installed []player.Kind
	checked   bool
	chosen    func() player.Kind
	choose    func(player.Kind)
	// paths are the players the user pointed at; programs, where they
	// point at them.
	paths    func() map[player.Kind]string
	programs map[player.Kind]*programSetting
}

func newPlayerSettings() *playerSettings {
	s := &playerSettings{programs: make(map[player.Kind]*programSetting)}
	for _, kind := range player.Kinds {
		s.programs[kind] = &programSetting{
			title: kind.Title(),
			found: func(ctx context.Context) (string, string, error) {
				path := kind.Find()
				if path == "" {
					return "", "", nil
				}
				about, err := player.Check(ctx, kind, path)
				return path, about, err
			},
			check: func(ctx context.Context, path string) (string, error) { return player.Check(ctx, kind, path) },
			custom: func() string {
				if s.paths == nil {
					return ""
				}
				return s.paths()[kind]
			},
			save: func(string) {},
		}
	}
	s.radios = radio.NewRadios(append(slices.Clone(player.Kinds), ""), "", func(kind player.Kind) {
		if s.choose != nil {
			s.choose(kind)
		}
	})
	return s
}

// refresh looks for installed players again, as one may have been installed
// or removed since.
func (s *playerSettings) refresh() {
	var paths map[player.Kind]string
	if s.paths != nil {
		paths = s.paths()
	}
	s.installed = player.Installed(paths)
	s.checked = true
	for _, kind := range player.Kinds {
		if s.has(kind) {
			s.radios.EnableOption(kind)
		} else {
			s.radios.DisableOption(kind)
		}
	}
	// There is nothing to ask about with one player or none.
	if len(s.installed) > 1 {
		s.radios.EnableOption("")
	} else {
		s.radios.DisableOption("")
	}
}

func (s *playerSettings) has(kind player.Kind) bool {
	for _, k := range s.installed {
		if k == kind {
			return true
		}
	}
	return false
}

func (s *playerSettings) current() (player.Kind, bool) {
	if !s.checked {
		s.refresh()
	}
	var chosen player.Kind
	if s.chosen != nil {
		chosen = s.chosen()
	}
	return playerFor(chosen, s.installed)
}

func (s *playerSettings) Update(gtx layout.Context) {
	for _, kind := range player.Kinds {
		s.programs[kind].Update(gtx)
	}
	kind, _ := s.current()
	s.radios.SetValue(kind)
	s.radios.Update(gtx)
}

// subtitle is the line under the section on the main settings page.
func (s *playerSettings) subtitle(l localization.Catalog) string {
	kind, ask := s.current()
	name := kind.Title()
	switch {
	case ask:
		name = l.T("player.not_chosen")
	case kind == "":
		name = l.T("player.not_installed")
	}
	return l.Format("player.subtitle", map[string]string{"player": name})
}

func (s *playerSettings) Layout(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	kind, ask := s.current()
	hint := l.T("player.hint")
	switch {
	case ask:
		hint = l.T("player.undecided")
	case kind == "":
		hint = l.T("player.none")
	case len(s.installed) == 1:
		hint = l.Format("player.only", map[string]string{"player": kind.Title()})
	}
	labels := map[player.Kind]string{"": l.T("player.ask_option")}
	for _, k := range player.Kinds {
		labels[k] = k.Title()
		if !s.has(k) {
			labels[k] += " · " + l.T("player.not_installed")
		}
	}
	return settingsChoiceCard(gtx, l.T("player.title"), hint, func(gtx layout.Context) layout.Dimensions {
		return s.radios.Layout(gtx, radio.LeadingKind, labels)
	})
}

// newBrowserSetting is the browser Mini Apps run in, as a program the user
// can point at.
func newBrowserSetting() *programSetting {
	return &programSetting{
		found: func(ctx context.Context) (string, string, error) {
			ref, banner := miniapp.FoundBrowser()
			return ref, banner, nil
		},
		check:  miniapp.CheckBrowser,
		custom: func() string { return "" },
		save:   func(string) {},
	}
}

// refreshPrograms looks for the programs of the integrations section again.
func (p *settingsPage) refreshPrograms() {
	p.players.refresh()
	for _, kind := range player.Kinds {
		p.players.programs[kind].refresh(context.Background(), p.invalidate)
	}
	p.browser.refresh(context.Background(), p.invalidate)
	p.stickers.program.refresh(context.Background(), p.invalidate)
}

func (p *settingsPage) layoutIntegrations(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	p.browser.title = l.T("program.browser")
	children := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return p.players.Layout(gtx, l)
	}), vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return p.stickers.Layout(gtx, l)
	})}
	for _, kind := range []player.Kind{player.VLC, player.MPV} {
		children = append(children, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.players.programs[kind].Layout(gtx, l)
		}))
	}
	children = append(children, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return p.stickers.program.Layout(gtx, l)
	}), vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return p.browser.Layout(gtx, l)
	}))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}
