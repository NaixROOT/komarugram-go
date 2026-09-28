// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"io"
	"os"
	"path/filepath"
	"time"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/chattheme"
	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"

	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/widget/scroll"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

type themeResult struct {
	appearance model.ChatAppearance
	style      *model.ChatThemeStyle
	image      *image.RGBA
	err        error
}
type themeChoicesResult struct {
	themes []model.ChatTheme
	err    error
}
type chatThemeController struct {
	loader                                         loadingIndicator
	saving                                         bool
	revision                                       uint64
	source                                         model.ChatThemeSource
	media                                          model.ConversationStore
	images                                         *imageOps
	invalidate                                     func()
	chat                                           int64
	dark                                           bool
	appearance                                     model.ChatAppearance
	style                                          *model.ChatThemeStyle
	background                                     *image.RGBA
	bubble                                         *image.RGBA
	phase                                          int64
	result                                         chan themeResult
	cancel                                         context.CancelFunc
	choicesResult                                  chan themeChoicesResult
	choicesCancel                                  context.CancelFunc
	choices                                        []model.ChatTheme
	clicks                                         []*surface
	selected                                       *model.ChatTheme
	local                                          map[int64]model.ChatTheme
	list                                           scroll.List
	apply, localButton, reset, importButton, retry surface
	path                                           *textField
	loading                                        bool
	problem                                        error
}

func newChatThemeController(source model.ConversationStore, images *imageOps, invalidate func()) *chatThemeController {
	t := &chatThemeController{media: source, images: images, invalidate: invalidate, local: map[int64]model.ChatTheme{}, path: newTextField(0, "")}
	t.source, _ = source.(model.ChatThemeSource)
	t.list.Axis = layout.Vertical
	return t
}
func (t *chatThemeController) Close() {
	if t.cancel != nil {
		t.cancel()
	}
	if t.choicesCancel != nil {
		t.choicesCancel()
	}
}
func (t *chatThemeController) Update(chat int64, dark bool) {
	revision := uint64(0)
	if source, ok := t.media.(interface{ ChatThemeRevision(int64) uint64 }); ok {
		revision = source.ChatThemeRevision(chat)
	}
	if chat != t.chat || dark != t.dark || revision != t.revision && t.selected == nil {
		t.revision = revision
		t.chat = chat
		t.dark = dark
		t.selected = nil
		t.appearance = model.ChatAppearance{}
		t.style = nil
		t.background = nil
		t.bubble = nil
		t.load()
	}
	if t.result != nil {
		select {
		case r := <-t.result:
			t.loading = false
			if r.err == nil && t.saving {
				t.selected = nil
			}
			t.saving = false
			t.result = nil
			t.problem = r.err
			if r.err == nil || r.style != nil || r.image != nil {
				t.appearance = r.appearance
				t.style = r.style
				t.background = r.image
				t.bubble = nil
			}
		default:
		}
	}
	if t.choicesResult != nil {
		select {
		case r := <-t.choicesResult:
			t.choicesResult = nil
			t.problem = r.err
			if r.err == nil {
				t.choices = r.themes
				t.clicks = make([]*surface, len(t.choices))
				for i := range t.clicks {
					t.clicks[i] = new(surface)
				}
			}
		default:
		}
	}
}
func (t *chatThemeController) job(fn func(context.Context) (model.ChatAppearance, error)) {
	if t.cancel != nil {
		t.cancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.cancel = cancel
	ch := make(chan themeResult, 1)
	t.result = ch
	t.loading = true
	t.problem = nil
	dark, chat := t.dark, t.chat
	go func() {
		defer cancel()
		defer crash.Recover("chat theme", func(e *crash.Panic) {
			select {
			case ch <- themeResult{err: e}:
				t.invalidate()
			default:
			}
		})
		a, err := fn(ctx)
		r := themeResult{appearance: a, err: err}
		if err == nil {
			style := a.Theme.Light
			if dark {
				style = a.Theme.Dark
			}
			if style == nil {
				if dark {
					style = a.Theme.Light
				} else {
					style = a.Theme.Dark
				}
			}
			r.style = style
			w := a.Wallpaper
			if w == nil && style != nil {
				w = style.Wallpaper
			}
			if w != nil {
				data := w.Image
				if w.Media != nil {
					data, err = t.media.Media(ctx, model.Message{Key: model.MessageKey{ChatID: chat}, Kind: model.MessagePhoto, Media: w.Media})
				}
				var im image.Image
				if err == nil && len(data) > 0 {
					im, err = chattheme.Decode(data)
				}
				r.image = chattheme.Wallpaper(w, im)
				// Keep the color fill if an image fails, but expose the error in appearance settings.
				r.err = err
			}
		}
		ch <- r
		t.invalidate()
	}()
}
func (t *chatThemeController) load() {
	chat := t.chat
	if a, ok := t.local[chat]; ok {
		t.job(func(context.Context) (model.ChatAppearance, error) { return model.ChatAppearance{Theme: a}, nil })
		return
	}
	if t.source != nil {
		t.job(func(ctx context.Context) (model.ChatAppearance, error) { return t.source.ChatAppearance(ctx, chat) })
	}
}
func (t *chatThemeController) LoadChoices() {
	if t.choicesResult != nil {
		return
	}
	if t.source == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.choicesCancel = cancel
	ch := make(chan themeChoicesResult, 1)
	t.choicesResult = ch
	go func() {
		defer cancel()
		defer crash.Recover("chat themes", func(e *crash.Panic) { ch <- themeChoicesResult{err: e}; t.invalidate() })
		themes, err := t.source.ChatThemes(ctx)
		ch <- themeChoicesResult{themes, err}
		t.invalidate()
	}()
}
func (t *chatThemeController) choose(theme model.ChatTheme) {
	t.selected = &theme
	t.job(func(context.Context) (model.ChatAppearance, error) { return model.ChatAppearance{Theme: theme}, nil })
}
func (t *chatThemeController) Background(gtx layout.Context) {
	if t.background != nil {
		widget.Image{Src: t.images.Op(t.background), Fit: widget.Cover}.Layout(gtx)
	}
}
func (t *chatThemeController) Bubble(gtx layout.Context, outgoing, animate bool, shape bubbleShape) {
	if t.style == nil {
		return
	}
	size := gtx.Constraints.Min
	if !outgoing {
		paint.FillShape(gtx.Ops, token.NewMatColorFromHexRGBA(t.style.Incoming).AsNRGBA(), shape.rrect(size).Op(gtx.Ops))
		return
	}
	colors := t.style.Outgoing
	if len(colors) == 0 {
		return
	}
	phase := int64(0)
	if t.style.Animated && animate && len(colors) > 2 {
		phase = gtx.Now.UnixMilli() / 100
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(100 * time.Millisecond)})
	}
	if t.bubble == nil || t.phase != phase {
		t.phase = phase
		t.bubble = chattheme.Gradient(colors, 64, 96, 0, float64(phase)/20)
	}
	defer shape.rrect(size).Push(gtx.Ops).Pop()
	gtx.Constraints = layout.Exact(size)
	widget.Image{Src: t.images.Op(t.bubble), Fit: widget.Fill}.Layout(gtx)
}
func (t *chatThemeController) messageContext(gtx layout.Context, outgoing bool) layout.Context {
	if t.style == nil {
		return gtx
	}
	theme := *wdk.GetMaterialTheme(gtx)
	sc := *theme.Scheme
	theme.Scheme = &sc
	fg := t.style.Text
	accent := t.style.Accent
	if outgoing {
		fg = t.style.OutText
		accent = t.style.OutAccent
	}
	sc.Surface.OnColor = token.NewMatColorFromHexRGBA(fg)
	sc.SurfaceVariant.OnColor = sc.Surface.OnColor.SetOpacity(.7)
	sc.Primary.Color = token.NewMatColorFromHexRGB(accent & 0xffffff)
	values := make(map[string]any, len(gtx.Values))
	for k, v := range gtx.Values {
		values[k] = v
	}
	values[wdk.ThemeNamespace] = &theme
	gtx.Values = values
	return gtx
}
func (t *chatThemeController) LayoutChoices(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	t.Update(t.chat, t.dark)
	if !t.saving && t.reset.Clicked(gtx) {
		t.choose(model.ChatTheme{})
	}
	for i, c := range t.clicks {
		if !t.saving && c.Clicked(gtx) {
			t.choose(t.choices[i])
		}
	}
	if t.retry.Clicked(gtx) {
		t.load()
		t.LoadChoices()
	}
	if !t.saving && t.importButton.Clicked(gtx) {
		path := t.path.Text()
		t.job(func(ctx context.Context) (model.ChatAppearance, error) {
			f, err := os.Open(path)
			if err != nil {
				return model.ChatAppearance{}, err
			}
			defer f.Close()
			data, err := io.ReadAll(io.LimitReader(f, chattheme.MaxBytes+1))
			if err != nil {
				return model.ChatAppearance{}, err
			}
			theme, err := chattheme.ImportDesktop(data)
			theme.Title = filepath.Base(path)
			return model.ChatAppearance{Theme: theme}, err
		})
		t.selected = &model.ChatTheme{Title: "import"}
	}
	if !t.loading && t.problem == nil && t.selected != nil && t.localButton.Clicked(gtx) {
		theme := t.appearance.Theme
		chat := t.chat
		t.saving = true
		if _, persistent := t.media.(model.LocalChatThemeSource); !persistent {
			t.local[chat] = theme
		}
		t.job(func(ctx context.Context) (model.ChatAppearance, error) {
			if source, ok := t.media.(model.LocalChatThemeSource); ok {
				if err := source.SetLocalChatTheme(ctx, chat, &theme); err != nil {
					return model.ChatAppearance{}, err
				}
			}
			return model.ChatAppearance{Theme: theme}, nil
		})
	}
	if !t.loading && t.problem == nil && t.selected != nil && t.apply.Clicked(gtx) && t.source != nil {
		theme := t.appearance.Theme
		chat := t.chat
		t.saving = true
		t.job(func(ctx context.Context) (model.ChatAppearance, error) {
			if err := t.source.SetChatTheme(ctx, chat, theme.ID); err != nil {
				return model.ChatAppearance{}, err
			}
			return t.source.ChatAppearance(ctx, chat)
		})
	}
	if t.loading || t.choicesResult != nil {
		return t.loader.page(gtx, l)
	}
	return t.list.Layout(gtx, len(t.choices)+5, func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(12).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			switch i {
			case 0:
				if t.problem != nil {
					return textButton(gtx, &t.retry, l.T("history.retry")+" · "+mediaErrorText(t.problem))
				}

				return label(gtx, t.appearance.Theme.Title, token.TypestyleTitleMedium, scheme(gtx).Surface.OnColor, 2)
			case 1:
				size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(130))
				gtx.Constraints = layout.Exact(size)
				fillRect(gtx, scheme(gtx).SurfaceContainerLow, size)
				t.Background(gtx)
				for i := 0; i < 2; i++ {
					b := gtx
					b.Constraints = layout.Exact(image.Pt(max(1, size.X*2/3), gtx.Dp(42)))
					x := gtx.Dp(8)
					if i == 1 {
						x = max(0, size.X-b.Constraints.Max.X-x)
					}
					offset(b, image.Pt(x, gtx.Dp(unit.Dp(12+i*58))), func(gtx layout.Context) layout.Dimensions {
						if t.style != nil {
							t.Bubble(gtx, i == 1, false, shapeOf(gtx, 0, i == 1))
						} else {
							fillRounded(gtx, scheme(gtx).Surface.Color, gtx.Constraints.Max, gtx.Dp(14))
						}
						return layout.UniformInset(10).Layout(t.messageContext(gtx, i == 1), func(gtx layout.Context) layout.Dimensions {
							return label(gtx, l.T("chat_theme.preview"), token.TypestyleBodyMedium, scheme(gtx).Surface.OnColor, 1)
						})
					})
				}
				return layout.Dimensions{Size: size}
			case 2:
				return textButton(gtx, &t.reset, l.T("chat_theme.default"))
			case 3:
				if t.selected == nil {
					return layout.Dimensions{}
				}
				if t.loading || t.problem != nil {
					gtx = gtx.Disabled()
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return textButton(gtx, &t.localButton, l.T("chat_theme.local"))
				}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if t.source == nil || t.appearance.Theme.ID == "" && t.appearance.Theme.Light != nil {
						return layout.Dimensions{}
					}
					return textButton(gtx, &t.apply, l.T("chat_theme.apply"))
				}))
			case 4:
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return t.path.Layout(gtx, l.T("chat_theme.path"), false) }), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return textButton(gtx, &t.importButton, l.T("chat_theme.import"))
				}))
			default:
				return textButton(gtx, t.clicks[i-5], t.choices[i-5].Title)
			}
		})
	})
}

func (t *chatThemeController) DiscardPreview() {
	if t.selected != nil && !t.saving {
		t.selected = nil
		t.load()
	}
}
