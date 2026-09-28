// SPDX-License-Identifier: Unlicense OR MIT

package exp

import (
	"gio-mw/token"
)

const (
	initialStackCapacity = 16
	stackGrowthFactor    = 2
)

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

// Current returns the current theme from the stack
func (s *SurfaceThemeStack) Current() *SurfaceTheme {
	if len(s.themes) > 0 {
		return s.themes[len(s.themes)-1]
	}
	panic("SurfaceThemeStack: no current theme")
}

// defaultStack is the global instance of SurfaceThemeStack
var defaultStack = NewSurfaceThemeStack()

// GetSurfaceTheme returns the current surface theme
func GetSurfaceTheme() *SurfaceTheme {
	return defaultStack.Current()
}

// NewSurfaceTheme creates and pushes a new surface theme
func NewSurfaceTheme(colorSet token.MatColorSet) *SurfaceTheme {
	return defaultStack.Push(colorSet)
}
