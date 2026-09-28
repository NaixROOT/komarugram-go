// SPDX-License-Identifier: Unlicense OR MIT

package chatmedia

import (
	"context"
	"image"
	"sync"
	"time"

	"komarugram/pkg/resample"
	"komarugram/pkg/vp9"
)

// A short video sticker can be decoded once, then played from immutable,
// tile-sized frames. The cap covers host pixels separately from the WASM
// decoder's own memory limit. Longer or larger files keep using streaming.
const (
	maxStickerClipBytes    = 16 << 20
	maxStickerClipFrames   = 120
	maxStickerClipDuration = 4 * time.Second
)

type stickerClip struct {
	frames        []*image.RGBA
	frameDuration time.Duration
	bytes         int64
}

func (c *stickerClip) FrameAt(elapsed time.Duration) *image.RGBA {
	index := int(elapsed/c.frameDuration) % len(c.frames)
	return c.frames[index]
}

func stickerClipFits(sticker *vp9.Sticker, box image.Point) bool {
	count := sticker.FrameCount()
	if count == 0 || count > maxStickerClipFrames || sticker.File.FrameDuration <= 0 ||
		sticker.File.FrameDuration > maxStickerClipDuration/time.Duration(count) {
		return false
	}
	size := resample.Fit(sticker.Size.X, sticker.Size.Y, box, false)
	return size.X > 0 && size.Y > 0 && int64(count)*int64(size.X)*int64(size.Y)*4 <= maxStickerClipBytes
}

func decodeStickerClip(ctx context.Context, sticker *vp9.Sticker, box image.Point, first func(*image.RGBA)) (*stickerClip, error) {
	clip := &stickerClip{frames: make([]*image.RGBA, 0, sticker.FrameCount()), frameDuration: sticker.File.FrameDuration}
	for i := range sticker.FrameCount() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		frame, err := stickerFrame(ctx, sticker, time.Duration(i)*clip.frameDuration, box)
		if err != nil {
			return nil, err
		}
		clip.frames = append(clip.frames, frame)
		clip.bytes += int64(len(frame.Pix))
		if i == 0 {
			first(frame)
		}
	}
	return clip, nil
}

// stickerFrame is the frame of sticker at elapsed, as a new image fitting
// box. It is scaled from the decoded planes, never composed at the
// sticker's own size, which is a grid cell's many times over.
func stickerFrame(ctx context.Context, sticker *vp9.Sticker, elapsed time.Duration, box image.Point) (*image.RGBA, error) {
	return sticker.FrameSized(ctx, elapsed, resample.Fit(sticker.Size.X, sticker.Size.Y, box, false))
}

// stickerPreview copies the first frame before closing both colour/alpha
// instances. Only the compiled module remains available for the next sticker.
func (m *Manager) stickerPreview(ctx context.Context, name string, data []byte, box image.Point) (image.Image, error) {
	rt, done, err := m.vp9.acquire(ctx)
	defer done()
	if err != nil {
		return nil, err
	}
	sticker, err := rt.OpenSticker(ctx, name, data)
	if err != nil {
		return nil, err
	}
	defer sticker.Close(context.Background())
	return stickerFrame(ctx, sticker, 0, box)
}

// waitStickerPlayback returns a decoder slot only while playback is requested.
func (m *Manager) waitStickerPlayback(ctx context.Context, e *entry) (func(), error) {
	for {
		e.mu.Lock()
		animate := e.animate
		e.mu.Unlock()
		if animate {
			select {
			case m.sem <- struct{}{}:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			e.mu.Lock()
			animate = e.animate
			e.mu.Unlock()
			if animate {
				return sync.OnceFunc(func() { <-m.sem }), nil
			}
			<-m.sem
		}
		select {
		case <-e.wake:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
