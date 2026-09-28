package model

import (
	"sort"
	"strings"
)

// TextRun is a non-overlapping interval with all nested styles applied.
type TextRun struct {
	Text, URL                                             string
	Bold, Italic, Code, Underline, Strike, Spoiler, Quote bool
	Emoji                                                 int64
}

func TextRuns(text string, entities []Entity) []TextRun {
	// Only codepoint boundaries are valid UTF-16 offsets. Never split a surrogate pair.
	boundary := map[int]int{0: 0}
	units := 0
	for b, r := range text {
		boundary[units] = b
		units++
		if r > 0xffff {
			units++
		}
	}
	boundary[units] = len(text)
	valid := make([]Entity, 0, len(entities))
	points := []int{0, units}
	for _, e := range entities {
		_, a := boundary[e.Offset]
		_, b := boundary[e.Offset+e.Length]
		if e.Offset < 0 || e.Length <= 0 || !a || !b {
			continue
		}
		valid = append(valid, e)
		points = append(points, e.Offset, e.Offset+e.Length)
	}
	sort.Ints(points)
	var out []TextRun
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		if a == b {
			continue
		}
		run := TextRun{Text: text[boundary[a]:boundary[b]]}
		for _, e := range valid {
			if e.Offset > a || e.Offset+e.Length < b {
				continue
			}
			switch e.Kind {
			case "bold":
				run.Bold = true
			case "italic":
				run.Italic = true
			case "code", "pre":
				run.Code = true
			case "underline":
				run.Underline = true
			case "strike":
				run.Strike = true
			case "spoiler":
				run.Spoiler = true
			case "quote":
				run.Quote = true
			case "emoji":
				run.Emoji = e.DocumentID
			case "url":
				run.URL = e.URL
				if run.URL == "" {
					run.URL = text[boundary[e.Offset]:boundary[e.Offset+e.Length]]
				}
			case "mention":
				run.URL = "https://t.me/" + strings.TrimPrefix(text[boundary[e.Offset]:boundary[e.Offset+e.Length]], "@")
			}
		}
		out = append(out, run)
	}
	return out
}
