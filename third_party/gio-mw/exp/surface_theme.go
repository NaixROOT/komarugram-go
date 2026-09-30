// SPDX-License-Identifier: Unlicense OR MIT

package exp

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
)

const (
	initialStackCapacity = 16
	stackGrowthFactor    = 2
)

// SurfaceNamespace is the key of the frame's surfaces in gtx.Values.
const SurfaceNamespace = "gio-mw/exp.surfaces"

// ThemeStack defines the interface for theme stack operations
type ThemeStack interface {
	Push(theme *SurfaceTheme) *SurfaceTheme
	Pop()
	Current() *SurfaceTheme
}

// SurfaceTheme represents a themed surface with colors
type SurfaceTheme struct {
	Color   token.MatColor
	OnColor token.MatColor
	stack   *SurfaceThemeStack
}

func (t *SurfaceTheme) Pop() {
	if t.stack != nil {
		t.stack.Pop()
	}
}

// SurfaceThemeStack manages a stack of surface themes
type SurfaceThemeStack struct {
	themes   []*SurfaceTheme
	position int
}

// NewSurfaceThemeStack creates a new surface theme stack
func NewSurfaceThemeStack() *SurfaceThemeStack {
	return &SurfaceThemeStack{
		themes: make([]*SurfaceTheme, 0, initialStackCapacity),
	}
}

// Push adds a new theme to the stack and returns it
func (s *SurfaceThemeStack) Push(colorSet token.MatColorSet) *SurfaceTheme {
	theme := &SurfaceTheme{
		Color:   colorSet.Color,
		OnColor: colorSet.OnColor,
		stack:   s,
	}

	if len(s.themes) < cap(s.themes) {
		s.themes = append(s.themes, theme)
	} else {
		newCap := cap(s.themes) * stackGrowthFactor
		newStack := make([]*SurfaceTheme, len(s.themes), newCap)
		copy(newStack, s.themes)
		s.themes = append(newStack, theme)
	}
	s.position++
	return theme
}

// Pop removes the specified theme from the stack if it's at the top
func (s *SurfaceThemeStack) Pop() {
	if s.position <= 0 {
		return
	}
	s.position--
	s.themes = s.themes[:s.position]
}

// reset empties the stack.
func (s *SurfaceThemeStack) reset() {
	clear(s.themes)
	s.themes = s.themes[:0]
	s.position = 0
}

// Current returns the current theme from the stack, nil when it is empty.
func (s *SurfaceThemeStack) Current() *SurfaceTheme {
	if len(s.themes) > 0 {
		return s.themes[len(s.themes)-1]
	}
	return nil
}

// surfaces is the stack of the frame gtx draws. It lives in gtx.Values,
// which a window makes anew for every frame: each window and each frame has
// its own, where a stack for the whole program would be pushed and popped by
// the windows' goroutines at once.
func surfaces(gtx layout.Context) *SurfaceThemeStack {
	if s, ok := gtx.Values[SurfaceNamespace].(*SurfaceThemeStack); ok {
		return s
	}
	if gtx.Values == nil {
		panic("exp: the context has no Values for its surfaces")
	}
	s := NewSurfaceThemeStack()
	gtx.Values[SurfaceNamespace] = s
	return s
}

// GetSurfaceTheme returns the surface being drawn on: the one pushed last
// in the frame, or the background of the material theme, which is under
// every surface.
func GetSurfaceTheme(gtx layout.Context) *SurfaceTheme {
	if theme := surfaces(gtx).Current(); theme != nil {
		return theme
	}
	background := wdk.GetMaterialTheme(gtx).Scheme.Background
	return &SurfaceTheme{Color: background.Color, OnColor: background.OnColor}
}

// NewSurfaceTheme creates and pushes a new surface theme on the frame's
// surfaces; Pop removes it.
func NewSurfaceTheme(gtx layout.Context, colorSet token.MatColorSet) *SurfaceTheme {
	return surfaces(gtx).Push(colorSet)
}
