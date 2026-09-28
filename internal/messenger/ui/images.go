// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gioui.org/op/paint"
)

type imageTexture struct {
	op   paint.ImageOp
	seen uint64
}

// ImageOp owns Gio's texture identity. Recreate it only when decoded pixels
// change, otherwise even an unchanged photo is converted/uploaded every frame.
type imageOps struct {
	generation uint64
	entries    map[image.Image]imageTexture
}

func (c *imageOps) BeginFrame() { c.generation++ }
func (c *imageOps) Op(im image.Image) paint.ImageOp {
	if c.entries == nil {
		c.entries = make(map[image.Image]imageTexture)
	}
	texture, ok := c.entries[im]
	if !ok {
		texture.op = paint.NewImageOp(im)
	}
	texture.seen = c.generation
	c.entries[im] = texture
	return texture.op
}

// Release forgets every image, so that a hidden window keeps no pixels
// alive through its texture handles.
func (c *imageOps) Release() { c.entries = nil }
func (c *imageOps) EndFrame() {
	for im, texture := range c.entries {
		if texture.seen+1 < c.generation {
			delete(c.entries, im)
		}
	}
}
