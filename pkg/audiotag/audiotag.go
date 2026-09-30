// SPDX-License-Identifier: Unlicense OR MIT

// Package audiotag reads what a music file says of itself, as Telegram
// Desktop reads it to send the file as a track: how long it plays, its
// title and performer, and the cover it holds. It knows the files Telegram
// Desktop sends as music: MP3 and AAC with ID3 tags, FLAC and Ogg (Vorbis
// or Opus) with Vorbis comments, and M4A with iTunes metadata.
//
// Nothing is decoded: durations come from headers, or from walking frame
// headers where a format keeps no total.
package audiotag

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"komarugram/pkg/voice"
)

// Info is what a music file says of itself.
type Info struct {
	Duration  time.Duration
	Title     string
	Performer string
	// Cover is the picture the tags hold, as its file's bytes: JPEG or
	// PNG, as the tags say; nil when there is none.
	Cover []byte
}

// Extensions are the files Telegram Desktop sends as music when it can
// read how long they play.
var Extensions = []string{".mp3", ".m4a", ".aac", ".ogg", ".oga", ".opus", ".flac"}

// MIME is the type a music file is sent with, by its name, or "" for a
// file that is not one of Extensions.
func MIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".aac":
		return "audio/aac"
	case ".ogg", ".oga", ".opus":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	}
	return ""
}

// ErrFormat is a file that is not one of Extensions' formats.
var ErrFormat = errors.New("not a music file")

// maxTag bounds what is read of a tag, covers included: a larger one is
// passed over.
const maxTag = 32 << 20

// Read reads the music file at path.
func Read(path string) (Info, error) {
	f, err := os.Open(path)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return Info{}, err
	}
	size := st.Size()
	var info Info
	switch MIME(path) {
	case "audio/mpeg":
		info = readID3(f, size)
		info.Duration, err = voice.FileDuration(path)
	case "audio/mp4":
		info = readMP4(f, size)
		info.Duration, err = voice.FileDuration(path)
	case "audio/aac":
		info = readID3(f, size)
		info.Duration, err = adtsDuration(f, size)
	case "audio/ogg":
		info, err = readOgg(f, size)
	case "audio/flac":
		info, err = readFLAC(f, size)
	default:
		return Info{}, ErrFormat
	}
	if err == nil && info.Duration <= 0 {
		err = errors.New("no sound")
	}
	if err != nil {
		return Info{}, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	info.Title = strings.TrimSpace(info.Title)
	info.Performer = strings.TrimSpace(info.Performer)
	return info, nil
}
