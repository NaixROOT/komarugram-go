// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"errors"
	"fmt"
	"strings"
)

// The emoji sets of Telegram Desktop are sprites without a list of what is
// in them: the order of the cells is the order in which its build lists
// the emoji, from a text file of its sources, lib_ui/emoji.txt. This is
// that order, worked out from the file the way the build does, for
// cmd/emoji-pack to write beside the sprites of a pack.
//
// The file has sections separated by lines of '=' or '-', each of parts
// separated by blank lines, each of lines of quoted emoji:
//
//	0: the categories of the panel, eight parts; an emoji with skin tones
//	   is a line of itself and its five tones, a pair of people with two
//	   tones a line of 26
//	1: parts that replace categories of section 0, found by their first
//	   emoji: the same emoji without the tones
//	2: the emoji that have skin tones, and in a second part those of two
//	   people
//	3: emoji outside the panel
//
// The cells follow the categories, the replaced ones as replaced, then
// section 3: an emoji, then its five tones, or the 25 pairs of tones of two
// people. An emoji met again has the cell it has.

// TelegramLayout is the layout of the sprites of Telegram Desktop's sets.
var TelegramLayout = Sprites{Cell: 72, Columns: 32, Rows: 16}

const (
	firstTone  = '\U0001F3FB'
	secondTone = '\U0001F3FC'
	lastTone   = '\U0001F3FF'
	joiner     = '‍'
)

// telegramFile is the list parsed: sections of parts of lines of emoji.
type telegramFile [][][][]string

func parseTelegramList(data string) (telegramFile, error) {
	var (
		file    telegramFile
		section [][][]string
		part    [][]string
	)
	endPart := func() {
		if len(part) > 0 {
			section = append(section, part)
		}
		part = nil
	}
	endSection := func() {
		endPart()
		if len(section) > 0 {
			file = append(file, section)
		}
		section = nil
	}
	for n, line := range strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			endPart()
			continue
		case strings.HasPrefix(line, "========") || strings.HasPrefix(line, "--------"):
			endSection()
			continue
		}
		var emoji []string
		for _, entry := range strings.Split(line, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if len(entry) < 3 || entry[0] != '"' || entry[len(entry)-1] != '"' || strings.Contains(entry[1:len(entry)-1], `"`) {
				return nil, fmt.Errorf("line %d: expected quoted emoji separated by commas, got %s", n+1, entry)
			}
			emoji = append(emoji, entry[1:len(entry)-1])
		}
		if len(emoji) > 0 {
			part = append(part, emoji)
		}
	}
	endSection()
	return file, nil
}

// bare is an emoji without U+FE0F.
func bare(runes []rune) string {
	out := make([]rune, 0, len(runes))
	for _, r := range runes {
		if r != selector {
			out = append(out, r)
		}
	}
	return string(out)
}

func count(runes []rune, r rune) int {
	n := 0
	for _, c := range runes {
		if c == r {
			n++
		}
	}
	return n
}

// twoTones are the patterns of an emoji of two people: with the first tone
// for both, and with the first and the second.
type twoTones struct{ same, different []rune }

// TelegramOrder returns the emoji of the cells of Telegram Desktop's
// sprites, in order, from the text of its emoji list. Each cell has its
// emoji and then the other sequences drawn the same; all are without
// U+FE0F.
func TelegramOrder(list string) ([][]string, error) {
	file, err := parseTelegramList(list)
	if err != nil {
		return nil, err
	}
	if len(file) < 3 || len(file[0]) != 8 || len(file[1]) > 8 {
		return nil, fmt.Errorf("the list has %d sections, the first of %d parts: not the list of Telegram Desktop", len(file), len(file[0]))
	}
	// The line of section 0 that begins with an emoji: itself with tones.
	lineOf := func(emoji string) []string {
		for _, part := range file[0] {
			for _, line := range part {
				if line[0] == emoji {
					return line
				}
			}
		}
		return nil
	}

	// aliases are the other sequences drawn as an emoji.
	aliases := map[string][]string{}
	alias := func(emoji, other string) {
		for _, a := range aliases[emoji] {
			if a == other {
				return
			}
		}
		aliases[emoji] = append(aliases[emoji], other)
	}

	toned := map[string]bool{}
	for _, line := range file[2][0] {
		for _, emoji := range line {
			tones := lineOf(emoji)
			if len(tones) != 6 {
				return nil, fmt.Errorf("%s is listed as having skin tones, and its line has %d emoji, not 6", emoji, len(tones))
			}
			first := []rune(tones[1])
			if len(first) < 2 || first[1] != firstTone || count(first, firstTone) != 1 {
				return nil, fmt.Errorf("%s: the tone is not the second character of %s", emoji, tones[1])
			}
			toned[bare(append(first[:1:1], first[2:]...))] = true
		}
	}

	pairs := map[string]twoTones{}
	if len(file[2]) > 1 {
		for _, line := range file[2][1] {
			for _, emoji := range line {
				tones := lineOf(emoji)
				if len(tones) != 26 {
					return nil, fmt.Errorf("%s is listed as two people with skin tones, and its line has %d emoji, not 26", emoji, len(tones))
				}
				original, same, different := []rune(tones[0]), []rune(tones[1]), []rune(tones[2])
				if len(original) < 1 || len(same) < 2 || len(different) < 5 {
					return nil, fmt.Errorf("%s: its line is not of two people", emoji)
				}
				withTone := func(pattern []rune, tone rune) string {
					out := append([]rune(nil), pattern...)
					for i, r := range out {
						if r == firstTone || r == secondTone {
							out[i] = tone
						}
					}
					return bare(out)
				}
				if count(same, firstTone) == 1 {
					// One glyph of both people takes one tone: the emoji
					// with a tone after its first character. The pair
					// spelled out with that tone twice is drawn the same.
					if same[1] != firstTone {
						return nil, fmt.Errorf("%s: the tone is not the second character of %s", emoji, tones[1])
					}
					for tone := rune(firstTone); tone <= lastTone; tone++ {
						alias(withTone(same, tone), withTone(different, tone))
					}
				} else if len(original) == 1 {
					// The pair is always spelled out; the emoji with one
					// tone after it is drawn as the pair of that tone.
					for tone := rune(firstTone); tone <= lastTone; tone++ {
						alias(bare(append(original[:1:1], tone)), withTone(same, tone))
					}
				}
				pattern := twoTones{same: same, different: different}
				if len(original) == 1 {
					pattern.same = []rune{original[0], firstTone}
				}
				if different[1] != firstTone || different[len(different)-1] != secondTone {
					return nil, fmt.Errorf("%s: %s does not have the first tone and then the second", emoji, tones[2])
				}
				pairs[bare(original)] = pattern
			}
		}
	}

	var order [][]string
	cell := map[string]int{}
	add := func(emoji string) error {
		if _, ok := cell[emoji]; ok {
			return nil
		}
		cell[emoji] = len(order)
		sequences := []string{emoji}
		for _, other := range aliases[emoji] {
			if _, taken := cell[other]; taken {
				return fmt.Errorf("%s, drawn as %s, has a cell of its own", other, emoji)
			}
			cell[other] = len(order)
			sequences = append(sequences, other)
		}
		order = append(order, sequences)
		return nil
	}
	category := func(lines [][]string) error {
		for _, line := range lines {
			for _, emoji := range line {
				runes := []rune(emoji)
				id := bare(runes)
				if id == "" {
					return errors.New("an empty emoji")
				}
				if count(runes, firstTone) > 0 {
					return fmt.Errorf("%s: an emoji with a tone among those without", emoji)
				}
				if err := add(id); err != nil {
					return err
				}
				if toned[id] {
					for tone := rune(firstTone); tone <= lastTone; tone++ {
						if err := add(bare(append(append(runes[:1:1], tone), runes[1:]...))); err != nil {
							return err
						}
					}
				} else if pattern, ok := pairs[id]; ok {
					for first := rune(firstTone); first <= lastTone; first++ {
						for second := rune(firstTone); second <= lastTone; second++ {
							from := pattern.different
							if first == second {
								from = pattern.same
							}
							out := append([]rune(nil), from...)
							for i, r := range out {
								switch r {
								case firstTone:
									out[i] = first
								case secondTone:
									out[i] = second
								}
							}
							if err := add(bare(out)); err != nil {
								return err
							}
						}
					}
				}
			}
		}
		return nil
	}
	for _, part := range file[0] {
		lines := part
		// The part of section 1 that begins with the same emoji replaces it.
		for _, replacement := range file[1] {
			if replacement[0][0] == part[0][0] {
				lines = replacement
			}
		}
		if err := category(lines); err != nil {
			return nil, err
		}
	}
	if len(file) > 3 {
		for _, part := range file[3] {
			if err := category(part); err != nil {
				return nil, err
			}
		}
	}
	return order, nil
}
