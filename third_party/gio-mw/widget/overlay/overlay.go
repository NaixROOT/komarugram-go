// SPDX-License-Identifier: Unlicense OR MIT

package overlay

import (
	"gioui.org/layout"
)

type Overlay struct {
	oStack   []*Item
	oUpdated bool
}

func (o *Overlay) Clear() {
	o.oStack = make([]*Item, 0)
}

// ClearItem closes the item with the given id; it animates out before it is
// removed.
func (o *Overlay) ClearItem(itemId int64) {
	for _, stackItem := range o.oStack {
		if stackItem.id == itemId {
			stackItem.Close()
			return
		}
	}
}

func (o *Overlay) Show(item *Item) {
	o.oStack = append(o.oStack, item)
}

func (o *Overlay) Update(gtx layout.Context) {
	// The top most item is the last one that is not closing.
	topMostIdx := -1
	for idx, oItem := range o.oStack {
		if !oItem.closed {
			topMostIdx = idx
		}
	}
	var newStack []*Item
	for idx, oItem := range o.oStack {
		oItem.update(gtx, idx == topMostIdx)
		if oItem.removed {
			continue
		}
		newStack = append(newStack, oItem)
	}
	o.oStack = newStack
	o.oUpdated = true
}

func (o *Overlay) Layout(gtx layout.Context) {
	if !o.oUpdated {
		panic("Overlay.Layout called before Overlay.Update")
	}
	oTheme := BuildTheme(gtx)
	for _, oItem := range o.oStack {
		oItem.layout(gtx, oTheme)
	}
}
