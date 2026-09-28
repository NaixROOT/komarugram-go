// SPDX-License-Identifier: Unlicense OR MIT

package player

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Stream hands an external player a file it can seek in without that file
// existing on disk. It answers byte ranges over loopback, which is what mpv
// and ffmpeg both ask for, so the bytes can come from a download that is still
// in progress: a ReaderAt that blocks until the range has arrived is enough.
type Stream struct {
	url      string
	listener net.Listener
	server   *http.Server
	requests atomic.Int64
}

// Serve publishes content on 127.0.0.1 under an unguessable path. The token
// keeps other local processes from reading the file just because they can
// reach the port.
func Serve(name string, content io.ReaderAt, size int64) (*Stream, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		_ = listener.Close()
		return nil, err
	}
	token := hex.EncodeToString(raw[:])
	path := "/" + token + "/" + name

	stream := &Stream{
		url:      fmt.Sprintf("http://%s%s", listener.Addr(), path),
		listener: listener,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		stream.requests.Add(1)
		// ServeContent turns Range headers into reads at an offset, so only
		// the part the player actually needs is ever touched.
		http.ServeContent(w, r, name, time.Time{}, io.NewSectionReader(content, 0, size))
	})
	stream.server = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go stream.server.Serve(listener) //nolint:errcheck // Serve always ends with an error on Close

	return stream, nil
}

// URL is the address to hand to the external player.
func (s *Stream) URL() string { return s.url }

// Requests counts how many times the player asked for data, which is a decent
// proxy for how often it seeked.
func (s *Stream) Requests() int64 { return s.requests.Load() }

// Close stops serving.
func (s *Stream) Close() error { return s.server.Close() }
