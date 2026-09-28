// SPDX-License-Identifier: Unlicense OR MIT

package rail

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/overlay"
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	closeIcon = wdk.RequireIconWidget(icons.NavigationClose)
)

type Rail struct {
	Active      *Button
	Buttons     []*Button
	closeButton button.Button
	CloseLabel  string
	Expandable  bool
	Expanded    bool

	overlayItem *overlay.Item
	scrollList  *widget.List
}

func (r *Rail) AsOverlayItem(buttonStyles []*ButtonStyle) *overlay.Item {
	r.scrollList = &widget.List{
		List: layout.List{Axis: layout.Vertical},
	}
	r.overlayItem = overlay.NewItem(func(gtx layout.Context) layout.Dimensions {
		r.update(gtx)
		return r.layout(gtx, buttonStyles)
	}, block.GravityMiddleStart).WithScrim()
	return r.overlayItem
}

func (r *Rail) Close(gtx layout.Context) {
	r.overlayItem.Close()
	gtx.Execute(op.InvalidateCmd{})
}

func (r *Rail) update(gtx layout.Context) {
	if r.closeButton.Clicked(gtx) {
		r.Close(gtx)
	}
}

func (r *Rail) layout(gtx layout.Context, buttonStyles []*ButtonStyle) layout.Dimensions {
	widgetTheme := BuildTheme(gtx)

	gtx.Constraints.Min.X = gtx.Dp(220)
	gtx.Constraints.Max.X = gtx.Dp(260)
	return block.Background{
		Color:        widgetTheme.CollapsedContainerColor,
		Elevation:    token.ElevationLevel1,
		CornerShapes: wdk.FromCornerShapesToken(gtx, widgetTheme.ExpandedModalContainerShape),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Container{
			MinSize: image.Point{X: gtx.Dp(widgetTheme.ExpandedContainerWidthMinimum)},
			MaxSize: image.Point{X: gtx.Dp(widgetTheme.ExpandedContainerWidthMaximum)},
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(16).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return block.Line{
					Axis:     block.AxisVertical,
					Overflow: block.OverflowClip,
					Expand:   true,
				}.Layout(gtx,
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = 0
						return r.closeButton.LayoutIconOnly(gtx, r.CloseLabel, closeIcon)
					}),
					block.NewVerticalSpacer(widgetTheme.ItemHeaderSpaceMinimum),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = 0
						return r.layoutPrimaryButtons(gtx, buttonStyles, widgetTheme)
					}),
				)
			})
		})
	})
}

func (r *Rail) layoutPrimaryButtons(gtx layout.Context, buttonStyles []*ButtonStyle, widgetTheme *Theme) layout.Dimensions {
	segments := make([]layout.Widget, len(buttonStyles))
	for idx, style := range buttonStyles {
		rButton := r.Buttons[idx]
		style.expanded = r.Expanded
		style.theme = widgetTheme
		if r.Active == rButton {
			style.active = true
		} else {
			style.active = false
		}
		segments[idx] = func(gtx layout.Context) layout.Dimensions {
			return rButton.layout(gtx, style)
		}
	}

	return r.scrollList.Layout(gtx, len(segments), func(gtx layout.Context, index int) layout.Dimensions {
		return segments[index](gtx)
	})
}
