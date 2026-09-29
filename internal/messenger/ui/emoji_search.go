// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode"

	"komarugram/internal/messenger/model"
)

// emojiEntry is an emoji of the search, with its name and keywords of a
// language, each a phrase of words.
type emojiEntry struct {
	emoji   string
	phrases [][]string
	// words are all the words of the phrases.
	words []string
}

// emojiIndexes are the entries of every language of emojiKeywords, in the
// order of the picker's sections.
var emojiIndexes = sync.OnceValue(func() map[string][]emojiEntry {
	out := map[string][]emojiEntry{}
	for lang, lines := range emojiKeywords {
		entries := make([]emojiEntry, 0, len(lines))
		for _, line := range lines {
			emoji, text, _ := strings.Cut(line, "\t")
			e := emojiEntry{emoji: emoji}
			// The first phrase is the name, the others keywords.
			for _, phrase := range strings.Split(text, "|") {
				words := searchWords(phrase)
				e.phrases = append(e.phrases, words)
				e.words = append(e.words, words...)
			}
			entries = append(entries, e)
		}
		out[lang] = entries
	}
	return out
})

// searchWords splits text into the lowercase words that the search compares,
// with ё as е.
func searchWords(text string) []string {
	return strings.FieldsFunc(strings.ReplaceAll(strings.ToLower(text), "ё", "е"), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// Ranks of a match, the better the higher: the words of the query put
// together over the whole emoji, and then in one phrase, starting one, or
// being one, each a step of two, one more for the name over a keyword.
const (
	matchTogether = 1
	matchWords    = 2 // every word of the query starts a word of the phrase
	matchPrefix   = 4 // the query starts the phrase
	matchExact    = 6 // the query is the phrase
	matchName     = 1 // added when the phrase is the name
)

// rank returns how well the words of the query match the phrases of the
// entry, 0 for not at all.
func (e emojiEntry) rank(query []string) int {
	best := 0
	for i, phrase := range e.phrases {
		rank := 0
		switch {
		case slices.Equal(phrase, query):
			rank = matchExact
		case len(phrase) >= len(query) && startsWords(phrase[:len(query)], query):
			rank = matchPrefix
		}
		if rank == 0 && coversWords(phrase, query) {
			rank = matchWords
		}
		if i == 0 && rank > 0 {
			rank += matchName
		}
		best = max(best, rank)
	}
	if best == 0 && coversWords(e.words, query) {
		return matchTogether
	}
	return best
}

// startsWords reports whether each word of prefix starts the word of words at
// its place.
func startsWords(words, prefix []string) bool {
	for i, p := range prefix {
		if i >= len(words) || !strings.HasPrefix(words[i], p) {
			return false
		}
	}
	return true
}

// coversWords reports whether every word of query starts some word of words.
func coversWords(words, query []string) bool {
	for _, q := range query {
		found := false
		for _, w := range words {
			if strings.HasPrefix(w, q) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// searchEmoji returns the emoji that match query, by their names and keywords
// in the language and in English: the best matches first, and those of a rank
// in the order of the sections. An emoji typed is found itself.
func searchEmoji(query, language string) []model.PickerItem {
	words := searchWords(query)
	typed := strings.TrimSpace(query)
	if len(words) == 0 && typed == "" {
		return nil
	}
	type match struct {
		emoji string
		rank  int
	}
	var matches []match
	seen := map[string]bool{}
	indexes := emojiIndexes()
	langs := []string{"en"}
	if _, ok := indexes[language]; ok && language != "en" {
		langs = []string{language, "en"}
	}
	for _, lang := range langs {
		for _, e := range indexes[lang] {
			rank := 0
			if len(words) > 0 {
				rank = e.rank(words)
			}
			if typed != "" && strings.Contains(typed, strings.ReplaceAll(e.emoji, "️", "")) {
				rank = matchExact + matchName
			}
			if rank == 0 {
				continue
			}
			if seen[e.emoji] {
				// Found in another language: the better rank counts.
				for i := range matches {
					if matches[i].emoji == e.emoji {
						matches[i].rank = max(matches[i].rank, rank)
					}
				}
				continue
			}
			seen[e.emoji] = true
			matches = append(matches, match{e.emoji, rank})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].rank > matches[j].rank })
	items := make([]model.PickerItem, len(matches))
	for i, m := range matches {
		items[i] = model.PickerItem{ID: "emoji/" + m.emoji, Emoji: m.emoji}
	}
	return items
}
