// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	_ "image/png" // sprites may be PNG
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // Telegram Desktop's sprites are WebP

	"gioui.org/text"
)

// selector is U+FE0F, which asks for the emoji presentation of the
// character before it. The emoji of a set are listed without it, and it is
// passed over in a text wherever it stands in an emoji.
const selector = '️'

// SpriteSet draws emoji from the sprites of a pack: text.EmojiImages. It
// is used by the shapers of all windows at once.
type SpriteSet struct {
	dir      string
	layout   Sprites
	revision string
	// rows are how many rows each picture has.
	rows []int
	// count is how many emoji the set has.
	count int

	// nodes and edges are the trie of the emoji's sequences: a node is an
	// emoji when its id is not negative, and an edge leads from a node by
	// a character.
	nodes []int32
	edges map[uint64]int32

	// cells are the cells cut out of the pictures when the pack was
	// installed, nil for a pack that was not: its pictures are decoded.
	cells *cells

	mu sync.Mutex
	// decoded are the pictures decoded last, for a set without cells: a
	// picture of Telegram Desktop's takes a tenth of a second to decode
	// and 10 MB decoded.
	decoded [decodedSprites]decodedSprite
	// uses counts the pictures asked for, to tell the one used longest ago.
	uses uint64
}

// decodedSprites is how many decoded pictures are kept.
const decodedSprites = 2

type decodedSprite struct {
	index int
	img   image.Image
	used  uint64
}

// OpenSprites opens the sprite set of pack p, whose files are in dir.
func OpenSprites(dir string, p Pack) (*SpriteSet, error) {
	if p.Kind != KindSprites || p.Sprites == nil {
		return nil, errors.New("not a sprite pack")
	}
	s := &SpriteSet{dir: dir, layout: *p.Sprites, revision: p.Revision(), nodes: []int32{-1}, edges: map[uint64]int32{}}
	cells := 0
	for i, name := range s.layout.Images {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		config, _, err := image.DecodeConfig(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		rows := config.Height / s.layout.Cell
		if config.Width != s.layout.Columns*s.layout.Cell || rows <= 0 || rows > s.layout.Rows || config.Height%s.layout.Cell != 0 {
			return nil, fmt.Errorf("%s is %dx%d, not a grid of %d columns of %d pixels", name, config.Width, config.Height, s.layout.Columns, s.layout.Cell)
		}
		if rows < s.layout.Rows && i != len(s.layout.Images)-1 {
			return nil, fmt.Errorf("%s has %d rows, and is not the last picture", name, rows)
		}
		s.rows = append(s.rows, rows)
		cells += rows * s.layout.Columns
	}
	f, err := os.Open(filepath.Join(dir, s.layout.Order))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	lines := bufio.NewScanner(f)
	lines.Buffer(nil, 1<<20)
	for lines.Scan() {
		sequences := strings.Fields(lines.Text())
		if len(sequences) == 0 {
			return nil, fmt.Errorf("%s: line %d is empty", s.layout.Order, s.count+1)
		}
		for _, sequence := range sequences {
			if err := s.add(sequence, s.count); err != nil {
				return nil, fmt.Errorf("%s: line %d: %w", s.layout.Order, s.count+1, err)
			}
		}
		s.count++
	}
	if err := lines.Err(); err != nil {
		return nil, err
	}
	if s.count == 0 || s.count > cells {
		return nil, fmt.Errorf("%d emoji are listed for %d cells", s.count, cells)
	}
	s.cells, _ = openCells(dir, s)
	return s, nil
}

func edge(node int32, r rune) uint64 { return uint64(node)<<32 | uint64(uint32(r)) }

// add puts a sequence into the trie as emoji id.
func (s *SpriteSet) add(sequence string, id int) error {
	node := int32(0)
	points := 0
	for _, r := range sequence {
		if r == selector {
			continue
		}
		points++
		next, ok := s.edges[edge(node, r)]
		if !ok {
			next = int32(len(s.nodes))
			s.nodes = append(s.nodes, -1)
			s.edges[edge(node, r)] = next
		}
		node = next
	}
	if points == 0 {
		return errors.New("an empty emoji")
	}
	if s.nodes[node] >= 0 && s.nodes[node] != int32(id) {
		return fmt.Errorf("%q is listed twice", sequence)
	}
	s.nodes[node] = int32(id)
	return nil
}

// Count is how many emoji the set has.
func (s *SpriteSet) Count() int { return s.count }

// Match implements text.EmojiImages: the longest emoji of the set that the
// text begins with. U+FE0F is passed over inside an emoji and taken with
// it after one. A lone character is an emoji only with U+FE0F after it,
// unless it is drawn as one on its own: a digit or © in a text is not.
func (s *SpriteSet) Match(runes []rune) (id, length int) {
	node := int32(0)
	points, best, bestPoints, selected := 0, -1, 0, false
	for i, r := range runes {
		if r == selector && i > 0 {
			if best >= 0 && length == i {
				length, selected = i+1, true
			}
			continue
		}
		next, ok := s.edges[edge(node, r)]
		if !ok {
			break
		}
		node = next
		points++
		if s.nodes[node] >= 0 {
			best, length, bestPoints, selected = int(s.nodes[node]), i+1, points, false
		}
	}
	if best < 0 || (bestPoints == 1 && !selected && !text.EmojiPresentation(runes[0])) {
		return 0, 0
	}
	return best, length
}

// Image implements text.EmojiImages: the picture of an emoji, scaled to
// size pixels square.
func (s *SpriteSet) Image(id, size int) image.Image {
	if id < 0 || id >= s.count || size <= 0 || size > 4096 {
		return nil
	}
	out := image.NewRGBA(image.Rect(0, 0, size, size))
	if s.cells != nil {
		if cell, err := s.cells.read(id); err == nil {
			draw.CatmullRom.Scale(out, out.Bounds(), cell, cell.Bounds(), draw.Src, nil)
			return out
		}
	}
	per := s.layout.Columns * s.layout.Rows
	sprite, err := s.sprite(id / per)
	if err != nil {
		return nil
	}
	cell := s.layout.Cell
	at := image.Pt(id%per%s.layout.Columns*cell, id%per/s.layout.Columns*cell).Add(sprite.Bounds().Min)
	draw.CatmullRom.Scale(out, out.Bounds(), sprite, image.Rectangle{Min: at, Max: at.Add(image.Pt(cell, cell))}, draw.Src, nil)
	return out
}

// sprite returns picture n decoded.
func (s *SpriteSet) sprite(n int) (image.Image, error) {
	if n < 0 || n >= len(s.layout.Images) {
		return nil, errors.New("no such picture")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uses++
	oldest := 0
	for i := range s.decoded {
		d := &s.decoded[i]
		if d.img != nil && d.index == n {
			d.used = s.uses
			return d.img, nil
		}
		if d.used < s.decoded[oldest].used {
			oldest = i
		}
	}
	f, err := os.Open(filepath.Join(s.dir, s.layout.Images[n]))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	s.decoded[oldest] = decodedSprite{index: n, img: img, used: s.uses}
	return img, nil
}
