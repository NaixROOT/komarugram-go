// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"slices"

	"gio-mw/token"
	"gio-mw/widget/toggle"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/preferences"
)

// notifySwitch is what the switch named by key turns on and off.
func notifySwitch(n *preferences.Notify, key string) *bool {
	switch key {
	case "desktop":
		return &n.Desktop
	case "sound":
		return &n.Sound
	case "name":
		return &n.Name
	case "text":
		return &n.Text
	case "private":
		return &n.Private
	case "groups":
		return &n.Groups
	case "channels":
		return &n.Channels
	case "accounts":
		return &n.AllAccounts
	}
	return new(bool)
}

// The groups of switches, as Telegram Desktop's section has them.
var (
	notifyGlobal   = []string{"desktop", "sound"}
	notifyShown    = []string{"name", "text"}
	notifyChats    = []string{"private", "groups", "channels"}
	notifyAccounts = []string{"accounts"}
)

type notifySettings struct {
	get    func() preferences.Notify
	set    func(preferences.Notify)
	groups map[*[]string]*toggle.Toggle[string]
}

func newNotifySettings() *notifySettings {
	s := &notifySettings{groups: map[*[]string]*toggle.Toggle[string]{}}
	for _, keys := range []*[]string{&notifyGlobal, &notifyShown, &notifyChats, &notifyAccounts} {
		s.groups[keys] = toggle.NewToggle(*keys, nil, func(values []string) {
			if s.get == nil || s.set == nil {
				return
			}
			n := s.get()
			for _, key := range *keys {
				*notifySwitch(&n, key) = slices.Contains(values, key)
			}
			s.set(n)
		})
	}
	return s
}

func (s *notifySettings) subtitle(l localization.Catalog) string {
	if s.get != nil && !s.get().Desktop {
		return l.T("notify.off")
	}
	return l.T("notify.on")
}

// group draws a card of switches; they follow the settings, which another
// window may have changed since the last frame.
func (s *notifySettings) group(gtx layout.Context, keys *[]string, title, hint string, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	t := s.groups[keys]
	n := s.get()
	var want []string
	for _, key := range *keys {
		if *notifySwitch(&n, key) {
			want = append(want, key)
		}
	}
	if !slices.Equal(want, t.GetValues()) {
		t.SetValues(want)
	}
	texts := map[string]string{}
	for _, key := range *keys {
		texts[key] = l.T("notify." + key)
	}
	if keys == &notifyAccounts {
		texts["accounts"] = l.T("notify.all_accounts")
	}
	return card(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return label(gtx, title, token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
			}),
			vspace(8),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return t.Layout(gtx, texts) }),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if hint == "" {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return label(gtx, hint, token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 0)
				})
			}),
		)
	}, defaultCardPadding)
}

// Layout draws the section; the accounts' switch shows only with more than
// one account.
func (s *notifySettings) Layout(gtx layout.Context, l localization.Catalog, accounts int) layout.Dimensions {
	if s.get == nil {
		return layout.Dimensions{}
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.group(gtx, &notifyGlobal, l.T("notify.global"), "", l)
		}),
		vspace(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.group(gtx, &notifyShown, l.T("notify.shown"), "", l)
		}),
		vspace(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.group(gtx, &notifyChats, l.T("notify.chats"), l.T("notify.chats_hint"), l)
		}),
	}
	if accounts > 1 {
		children = append(children, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.group(gtx, &notifyAccounts, l.T("notify.from"), l.T("notify.all_accounts_hint"), l)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}
