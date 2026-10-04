// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"komarugram/internal/messenger/historycache"
	"komarugram/internal/messenger/model"
	"komarugram/pkg/dcpool"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const streamPartSize = 128 << 10

// rangeReader fetches only aligned ranges requested by the external player (including seeks).
// The decrypted working set is bounded independently of the file's total size.
type rangeReader struct {
	cdn      *cdnRange
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	store    *Store
	message  model.Message
	location fileLocation
	size     int64
	chunks   map[int64][]byte
	order    []int64
}

func (s *Store) MediaStream(ctx context.Context, m model.Message) (io.ReaderAt, int64, func(), error) {
	if m.Media == nil || m.Media.Size <= 0 {
		return nil, 0, nil, errors.New("video size unavailable")
	}
	c := s.history
	c.mu.Lock()
	cache := c.cache
	ref, ok := c.refs[m.Media.ID]
	c.mu.Unlock()
	if cache == nil {
		return nil, 0, nil, errors.New("cache unavailable")
	}
	if b, e := cache.Media(ctx, m.Media.ID, historycache.RefOf(m, m.Media.ID)); e != nil {
		return nil, 0, nil, e
	} else if len(b) > 0 {
		return bytes.NewReader(b), int64(len(b)), func() {}, nil
	}
	if !ok {
		var e error
		ok, e = cache.Get(ctx, "ref/"+m.Media.ID, &ref)
		if e != nil {
			return nil, 0, nil, e
		}
		if !ok {
			return nil, 0, nil, errors.New("video location unavailable")
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	r := &rangeReader{ctx: ctx, cancel: cancel, store: s, message: m, location: ref, size: m.Media.Size, chunks: map[int64][]byte{}}
	return r, r.size, func() {
		cancel()
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.cdn != nil {
			r.cdn.close()
		}
	}, nil
}
func (r *rangeReader) ReadAt(p []byte, off int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if off < 0 {
		return 0, errors.New("negative video offset")
	}
	if len(p) == 0 {
		return 0, nil
	}
	if off >= r.size {
		return 0, io.EOF
	}
	n := 0
	for len(p) > 0 && off < r.size {
		part := off / int64(streamPartSize) * int64(streamPartSize)
		b, e := r.chunk(part)
		if e != nil {
			return n, e
		}
		inside := int(off - part)
		if inside >= len(b) {
			return n, io.ErrUnexpectedEOF
		}
		take := min(len(p), len(b)-inside)
		take = min(take, int(r.size-off))
		copy(p, b[inside:inside+take])
		n += take
		off += int64(take)
		p = p[take:]
	}
	if len(p) > 0 {
		return n, io.EOF
	}
	return n, nil
}
func (r *rangeReader) chunk(offset int64) ([]byte, error) {
	if e := r.ctx.Err(); e != nil {
		return nil, e
	}
	if b := r.chunks[offset]; b != nil {
		return b, nil
	}
	c := r.store.history
	c.mu.Lock()
	cache, pool := c.cache, c.pool
	c.mu.Unlock()
	key := fmt.Sprintf("%s/range/%d", r.message.Media.ID, offset)
	b, e := cache.Media(r.ctx, key, historycache.RefOf(r.message, key))
	if e != nil {
		return nil, e
	}
	if len(b) == 0 {
		if pool == nil {
			return nil, errors.New("video range is not cached; reconnect to Telegram")
		}
		ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
		defer cancel()
		origin := dcpool.DownloadClient(pool, ctx, r.location.DC)
		fetch := func() (tg.UploadFileClass, error) {
			return origin.UploadGetFile(ctx, &tg.UploadGetFileRequest{Precise: true, CDNSupported: true, Location: r.location.input(), Offset: offset, Limit: streamPartSize})
		}
		for attempt := 0; attempt < 3; attempt++ {
			if r.cdn != nil {
				b, e = r.cdn.read(ctx, offset, int(min(int64(streamPartSize), r.size-offset)))
				if invalidCDNToken(e) {
					r.cdn.close()
					r.cdn = nil
					continue
				}
				break
			}
			file, err := fetch()
			e = err
			if tgerr.Is(e, "FILE_REFERENCE_EXPIRED", "FILE_REFERENCE_EMPTY", "FILE_REFERENCE_INVALID") {
				if e = r.store.refreshReference(ctx, r.message); e != nil {
					return nil, e
				}
				c.mu.Lock()
				r.location = c.refs[r.message.Media.ID]
				c.mu.Unlock()
				continue
			}
			if e != nil {
				return nil, e
			}
			switch f := file.(type) {
			case *tg.UploadFile:
				b = f.Bytes
			case *tg.UploadFileCDNRedirect:
				provider, ok := origin.(downloader.CDNProvider)
				if !ok {
					return nil, errors.New("CDN transport unavailable")
				}
				r.cdn = &cdnRange{origin: origin, provider: provider, redirect: f}
				continue
			default:
				return nil, fmt.Errorf("unexpected file response %T", file)
			}
			break
		}
		if e != nil {
			return nil, e
		}
		if len(b) > streamPartSize {
			return nil, errors.New("oversized video range")
		}
		if len(b) == 0 {
			return nil, io.ErrUnexpectedEOF
		}
		if e = cache.SaveMedia(ctx, key, b, historycache.RefOf(r.message, key)); e != nil {
			return nil, e
		}
	}
	if len(r.order) >= 64 {
		delete(r.chunks, r.order[0])
		r.order = r.order[1:]
	}
	r.order = append(r.order, offset)
	r.chunks[offset] = b
	return b, nil
}
