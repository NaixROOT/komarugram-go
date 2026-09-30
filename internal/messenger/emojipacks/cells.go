// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"
	"path/filepath"
)

// A sprite of Telegram Desktop's is a WebP of 2304 by 1152 pixels, which
// takes a tenth of a second to decode and 10 MB decoded, for one emoji of
// 72 pixels. The cells of an installed pack are therefore cut out once,
// when it is installed, into a file beside its sprites: each cell
// compressed on its own, to be read alone. The file is made from the
// pack's files and is not one of them.
//
//	"KGEC" version:u32 revision:16 bytes cell:u32 count:u32
//	offsets:u32 × (count+1), from the end of the table
//	cells: deflate of cell × cell × 4 bytes, NRGBA, each byte less the
//	       one of the pixel before it in the row

const (
	cellsName    = ".cells"
	cellsMagic   = "KGEC"
	cellsVersion = 2
)

// cellsHeader is the size of what precedes the table of offsets.
const cellsHeader = 4 + 4 + 16 + 4 + 4

// cells is an opened file of cells.
type cells struct {
	path    string
	cell    int
	offsets []uint32
	// base is where the cells begin in the file.
	base int64
}

// buildCells cuts the cells of the sprite pack p, whose files are in dir,
// into the file of cells there.
func buildCells(dir string, p Pack) error {
	set, err := OpenSprites(dir, p)
	if err != nil {
		return err
	}
	cell := set.layout.Cell
	per := set.layout.Columns * set.layout.Rows
	var (
		body    bytes.Buffer
		offsets = make([]uint32, 0, set.count+1)
		raw     = image.NewNRGBA(image.Rect(0, 0, cell, cell))
	)
	packer, err := flate.NewWriter(&body, flate.BestSpeed)
	if err != nil {
		return err
	}
	filtered := make([]byte, len(raw.Pix))
	for id := 0; id < set.count; id++ {
		sprite, err := set.sprite(id / per)
		if err != nil {
			return err
		}
		at := image.Pt(id%per%set.layout.Columns*cell, id%per/set.layout.Columns*cell).Add(sprite.Bounds().Min)
		draw.Draw(raw, raw.Bounds(), sprite, at, draw.Src)
		offsets = append(offsets, uint32(body.Len()))
		// Neighbours are alike: their differences pack better.
		for y := range cell {
			row := raw.Pix[y*raw.Stride : y*raw.Stride+cell*4]
			out := filtered[y*cell*4:]
			copy(out[:4], row[:4])
			for i := 4; i < len(row); i++ {
				out[i] = row[i] - row[i-4]
			}
		}
		packer.Reset(&body)
		if _, err := packer.Write(filtered); err != nil {
			return err
		}
		if err := packer.Close(); err != nil {
			return err
		}
	}
	offsets = append(offsets, uint32(body.Len()))

	var head bytes.Buffer
	head.WriteString(cellsMagic)
	binary.Write(&head, binary.LittleEndian, uint32(cellsVersion))
	revision := make([]byte, 16)
	copy(revision, set.revision)
	head.Write(revision)
	binary.Write(&head, binary.LittleEndian, uint32(cell))
	binary.Write(&head, binary.LittleEndian, uint32(set.count))
	binary.Write(&head, binary.LittleEndian, offsets)
	tmp := filepath.Join(dir, cellsName+".tmp")
	if err := os.WriteFile(tmp, append(head.Bytes(), body.Bytes()...), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, cellsName))
}

// openCells opens the file of cells of a set, if it is there and is of the
// set as it is.
func openCells(dir string, s *SpriteSet) (*cells, error) {
	path := filepath.Join(dir, cellsName)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	head := make([]byte, cellsHeader)
	if _, err := io.ReadFull(f, head); err != nil {
		return nil, err
	}
	revision := make([]byte, 16)
	copy(revision, s.revision)
	if string(head[:4]) != cellsMagic || binary.LittleEndian.Uint32(head[4:]) != cellsVersion || !bytes.Equal(head[8:24], revision) ||
		int(binary.LittleEndian.Uint32(head[24:])) != s.layout.Cell || int(binary.LittleEndian.Uint32(head[28:])) != s.count {
		return nil, errors.New("the cells are of another pack")
	}
	c := &cells{path: path, cell: s.layout.Cell, offsets: make([]uint32, s.count+1)}
	if err := binary.Read(f, binary.LittleEndian, c.offsets); err != nil {
		return nil, err
	}
	c.base = cellsHeader + int64(len(c.offsets))*4
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(c.offsets); i++ {
		if c.offsets[i] < c.offsets[i-1] {
			return nil, errors.New("the cells are not in order")
		}
	}
	if c.base+int64(c.offsets[s.count]) != info.Size() {
		return nil, fmt.Errorf("the file of cells is %d bytes, not %d", info.Size(), c.base+int64(c.offsets[s.count]))
	}
	return c, nil
}

// read returns cell id. The file is opened for the one read: an open file
// would keep the pack from being deleted or replaced.
func (c *cells) read(id int) (*image.NRGBA, error) {
	f, err := os.Open(c.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	packed := io.NewSectionReader(f, c.base+int64(c.offsets[id]), int64(c.offsets[id+1]-c.offsets[id]))
	img := image.NewNRGBA(image.Rect(0, 0, c.cell, c.cell))
	r := flate.NewReader(packed)
	defer r.Close()
	if _, err := io.ReadFull(r, img.Pix); err != nil {
		return nil, err
	}
	for y := range c.cell {
		row := img.Pix[y*img.Stride : y*img.Stride+c.cell*4]
		for i := 4; i < len(row); i++ {
			row[i] += row[i-4]
		}
	}
	return img, nil
}
